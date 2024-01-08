package helm

import (
	"embed"
	"log"
	"os"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
	"securesign/sigstore-ocp/tas-installer/pkg/oidc"
	"text/template"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
)

const templateValuesFile = "values-openshift.tmpl"

var (
	//go:embed values-openshift.tmpl
	templateFS embed.FS
	values     = make(map[string]interface{})
)

type templatedValues struct {
	OpenShiftAppsSubdomain string
	OIDCconfig             oidc.OIDCConfig
}

func UninstallTrustedArtifactSigner(tasNamespace, tasReleaseName string) (*release.UninstallReleaseResponse, error) {
	actionConfig, _, err := actionConfig(tasNamespace)
	if err != nil {
		return nil, err
	}
	return action.NewUninstall(actionConfig).Run(tasReleaseName)
}

func InstallTrustedArtifactSigner(kc *kubernetes.KubernetesClient, oidcConfig oidc.OIDCConfig, tasNamespace, tasReleaseName, pathToValuesFile, chartLocation, chartVersion string) error {

	tv := templatedValues{
		OpenShiftAppsSubdomain: kc.ClusterCommonName,
		OIDCconfig:             oidcConfig,
	}

	tmpl, err := template.ParseFS(templateFS, templateValuesFile)
	if err != nil {
		return err
	}

	if pathToValuesFile != "" {
		if err := parseValuesFile(pathToValuesFile); err != nil {
			return err
		}
	} else {
		// if no values passed, use the default templated values
		tmpFile, err := os.CreateTemp("", "values-*.yaml")
		if err != nil {
			return err
		}
		defer tmpFile.Close()
		err = tmpl.Execute(tmpFile, tv)
		if err != nil {
			return err
		}
		if err := parseValuesFile(tmpFile.Name()); err != nil {
			return err
		}
	}

	client, err := registry.NewClient()
	if err != nil {
		return err
	}
	actionConfig, settings, err := actionConfig(tasNamespace)
	if err != nil {
		return err
	}

	lister := action.NewList(actionConfig)
	lister.AllNamespaces = true
	releases, err := lister.Run()
	if err != nil {
		return err
	}
	exists := false
	for _, rel := range releases {
		if rel.Name == tasReleaseName && rel.Namespace == tasNamespace {
			exists = true
			if err := upgradeRelease(actionConfig, client, settings, tasNamespace, chartLocation, chartVersion, values); err != nil {
				return err
			}
		}
	}
	if !exists {
		if err := installNewRelease(actionConfig, client, settings, tasNamespace, tasReleaseName, chartLocation, chartVersion, values); err != nil {
			return err
		}
	}
	return nil
}

func actionConfig(tasNamespace string) (*action.Configuration, *cli.EnvSettings, error) {
	settings := cli.New()
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), tasNamespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return nil, nil, err
	}
	return actionConfig, settings, nil
}

func installNewRelease(actionConfig *action.Configuration, client *registry.Client, settings *cli.EnvSettings, tasNamespace, tasReleaseName, chartLocation, chartVersion string, values map[string]interface{}) error {
	install := action.NewInstall(actionConfig)
	install.ReleaseName = tasReleaseName
	install.Namespace = tasNamespace
	install.CreateNamespace = true
	install.Version = chartVersion
	install.SetRegistryClient(client)

	chartPath, err := install.LocateChart(chartLocation, settings)
	if err != nil {
		return err
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return err
	}

	_, err = install.Run(chart, values)
	if err != nil {
		return err
	}

	return nil
}

func upgradeRelease(actionConfig *action.Configuration, client *registry.Client, settings *cli.EnvSettings, tasNamespace, chartLocation, chartVersion string, values map[string]interface{}) error {
	upgrade := action.NewUpgrade(actionConfig)
	upgrade.Namespace = tasNamespace
	upgrade.Version = chartVersion
	upgrade.SetRegistryClient(client)

	chartPath, err := upgrade.LocateChart(chartLocation, settings)
	if err != nil {
		return err
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return err
	}

	_, err = upgrade.Run(tasNamespace, chart, values)
	if err != nil {
		return err
	}

	return nil
}

func parseValuesFile(pathToValuesFile string) error {
	data, err := os.ReadFile(pathToValuesFile)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal([]byte(data), &values); err != nil {
		return err
	}

	return nil
}
