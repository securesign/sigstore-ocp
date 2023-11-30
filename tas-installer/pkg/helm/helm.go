package helm

import (
	"context"
	"log"
	"os"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
)

var (
	values = make(map[string]interface{})
)

func InstallTrustedArtifactSigner(kc *kubernetes.KubernetesClient, pathToValuesFile, chartVersion string) error {
	chartUrl := "oci://quay.io/redhat-user-workloads/arewm-tenant/sigstore-ocp/trusted-artifact-signer"
	if pathToValuesFile != "" {
		if err := parseValuesFile(pathToValuesFile, kc.ClusterCommonName); err != nil {
			return err
		}
	}

	settings := cli.New()
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), "trusted-artifact-signer", os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return err
	}

	client, err := registry.NewClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	exists, err := kc.NamespaceExists(ctx, "trusted-artifact-signer")
	if err != nil {
		return err
	}

	if exists {
		upgradeRelease(actionConfig, client, settings, chartUrl, chartVersion, values)
	} else {
		installNewRelease(actionConfig, client, settings, chartUrl, chartVersion, values)
	}
	return nil
}

func installNewRelease(actionConfig *action.Configuration, client *registry.Client, settings *cli.EnvSettings, chartURL, chartVersion string, values map[string]interface{}) error {
	install := action.NewInstall(actionConfig)
	install.ReleaseName = "trusted-artifact-signer"
	install.Namespace = "trusted-artifact-signer"
	install.CreateNamespace = true
	install.Version = chartVersion
	install.SetRegistryClient(client)

	chartPath, err := install.LocateChart(chartURL, settings)
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

func upgradeRelease(actionConfig *action.Configuration, client *registry.Client, settings *cli.EnvSettings, chartURL, chartVersion string, values map[string]interface{}) error {
	upgrade := action.NewUpgrade(actionConfig)
	upgrade.Namespace = "trusted-artifact-signer"
	upgrade.Version = chartVersion
	upgrade.SetRegistryClient(client)

	chartPath, err := upgrade.LocateChart(chartURL, settings)
	if err != nil {
		return err
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		return err
	}

	_, err = upgrade.Run("trusted-artifact-signer", chart, values)
	if err != nil {
		return err
	}

	return nil
}

func parseValuesFile(pathToValuesFile, commonName string) error {
	data, err := os.ReadFile(pathToValuesFile)
	if err != nil {
		return err
	}
	modifiedData := replaceOpenShiftAppsSubdomain(data, commonName)

	if err = yaml.Unmarshal([]byte(modifiedData), &values); err != nil {
		return err
	}

	return nil
}

func replaceOpenShiftAppsSubdomain(data []byte, commonName string) string {
	return strings.ReplaceAll(string(data), "$OPENSHIFT_APPS_SUBDOMAIN", commonName)
}