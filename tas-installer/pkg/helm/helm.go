package helm

import (
	"context"
	"log"
	"os"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
)

func InstallTrustedArtifactSigner(kc *kubernetes.KubernetesClient, pathToValuesFile, chartVersion string) error {
	chartUrl := "oci://quay.io/redhat-user-workloads/arewm-tenant/sigstore-ocp/trusted-artifact-signer"

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
		if err := upgradeRelease(actionConfig, client, settings, kc.ClusterCommonName, chartUrl, chartVersion, pathToValuesFile); err != nil {
			return err
		}
	} else {
		if err := installNewRelease(actionConfig, client, settings, kc.ClusterCommonName, chartUrl, chartVersion, pathToValuesFile); err != nil {
			return err
		}
	}
	return nil
}

func installNewRelease(actionConfig *action.Configuration, client *registry.Client, settings *cli.EnvSettings, clusterCommonName, chartURL, chartVersion, pathToValuesFile string) error {
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

	values, err := parseValuesFile(settings, clusterCommonName, pathToValuesFile)
	if err != nil {
		return err
	}

	_, err = install.Run(chart, values)
	if err != nil {
		return err
	}

	return nil
}

func upgradeRelease(actionConfig *action.Configuration, client *registry.Client, settings *cli.EnvSettings, clusterCommonName, chartURL, chartVersion, pathToValuesFile string) error {
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

	values, err := parseValuesFile(settings, clusterCommonName, pathToValuesFile)
	if err != nil {
		return err
	}

	_, err = upgrade.Run("trusted-artifact-signer", chart, values)
	if err != nil {
		return err
	}

	return nil
}

func parseValuesFile(settings *cli.EnvSettings, clusterCommonName, pathToValuesFile string) (map[string]interface{}, error) {
	var valueOpts values.Options

	defaultValues := []string{
		"global.appsSubdomain=" + clusterCommonName,
		"scaffold.fulcio.server.ingress.http.hosts[0].host=fulcio." + clusterCommonName,
		"scaffold.fulcio.server.ingress.http.hosts[0].path=/",
		"scaffold.rekor.server.ingress.hosts[0].host=rekor." + clusterCommonName,
		"scaffold.rekor.server.ingress.hosts[0].path=/",
		"scaffold.tuf.ingress.http.hosts[0].host=tuf." + clusterCommonName,
		"scaffold.tuf.ingress.http.hosts[0].path=/",
	}

	if pathToValuesFile != "" {
		valueOpts = values.Options{
			ValueFiles: []string{pathToValuesFile},
			Values:     defaultValues,
		}
	} else {
		valueOpts = values.Options{
			Values: defaultValues,
		}
	}

	mergedValues, err := valueOpts.MergeValues(getter.All(settings))
	if err != nil {
		return nil, err
	}

	return mergedValues, nil
}