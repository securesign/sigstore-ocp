package helm

import (
	"context"
	"fmt"
	"log"
	"os"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
	"securesign/sigstore-ocp/tas-installer/ui"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
)

func InstallTrustedArtifactSigner(kc *kubernetes.KubernetesClient, pathToValuesFile, chartVersion string, oidcConfig ui.OIDCConfig) error {
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
		if err := upgradeRelease(actionConfig, client, settings, oidcConfig, kc.ClusterCommonName, chartUrl, chartVersion, pathToValuesFile); err != nil {
			return err
		}
	} else {
		if err := installNewRelease(actionConfig, client, settings, oidcConfig, kc.ClusterCommonName, chartUrl, chartVersion, pathToValuesFile); err != nil {
			return err
		}
	}
	return nil
}

func installNewRelease(actionConfig *action.Configuration, client *registry.Client, settings *cli.EnvSettings, oidcConfig ui.OIDCConfig, clusterCommonName, chartURL, chartVersion, pathToValuesFile string) error {
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

	defaultValueOpts, err := parseValuesFile(clusterCommonName, pathToValuesFile)
	if err != nil {
		return err
	}

	oidcValueOpts, err := parseOIDCProvider(oidcConfig)
	if err != nil {
		return err
	}

	values, err := mergeValueOpts(settings, defaultValueOpts, oidcValueOpts)
	if err != nil {
		return err
	}

	_, err = install.Run(chart, values)
	if err != nil {
		return err
	}

	return nil
}

func upgradeRelease(actionConfig *action.Configuration, client *registry.Client, settings *cli.EnvSettings, oidcConfig ui.OIDCConfig, clusterCommonName, chartURL, chartVersion, pathToValuesFile string) error {
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

	defaultValueOpts, err := parseValuesFile(clusterCommonName, pathToValuesFile)
	if err != nil {
		return err
	}

	oidcValueOpts, err := parseOIDCProvider(oidcConfig)
	if err != nil {
		return err
	}

	values, err := mergeValueOpts(settings, defaultValueOpts, oidcValueOpts)
	if err != nil {
		return err
	}

	_, err = upgrade.Run("trusted-artifact-signer", chart, values)
	if err != nil {
		return err
	}

	return nil
}

func parseValuesFile(clusterCommonName, pathToValuesFile string) (values.Options, error) {
	var defaultValueOpts values.Options

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
		defaultValueOpts = values.Options{
			ValueFiles: []string{pathToValuesFile},
			Values:     defaultValues,
		}
	} else {
		defaultValueOpts = values.Options{
			Values: defaultValues,
		}
	}

	return defaultValueOpts, nil
}

func parseOIDCProvider(oidcConfig ui.OIDCConfig) (values.Options, error) {
	if oidcConfig.IssuerURL == "" || oidcConfig.ClientID == "" {
		return values.Options{}, fmt.Errorf("invalid OIDC configuration")
	}

	oidcConfig.IssuerURL = strings.ReplaceAll(oidcConfig.IssuerURL, ".", "\\.")

	oidcValues := []string{
		fmt.Sprintf("scaffold.fulcio.config.contents.OIDCIssuers.%s.IssuerURL=%s", oidcConfig.IssuerURL, oidcConfig.IssuerURL),
		fmt.Sprintf("scaffold.fulcio.config.contents.OIDCIssuers.%s.ClientID=%s", oidcConfig.IssuerURL, oidcConfig.ClientID),
		fmt.Sprintf("scaffold.fulcio.config.contents.OIDCIssuers.%s.Type=%s", oidcConfig.IssuerURL, oidcConfig.Type),
	}

	return values.Options{Values: oidcValues}, nil
}

func mergeValueOpts(settings *cli.EnvSettings, defaultValueOpts, oidcValueOpts values.Options) (map[string]interface{}, error) {
	var combinedValueOpts values.Options

	combinedValueOpts = values.Options{
		ValueFiles:    append(defaultValueOpts.ValueFiles, oidcValueOpts.ValueFiles...),
		StringValues:  append(defaultValueOpts.StringValues, oidcValueOpts.StringValues...),
		Values:        append(defaultValueOpts.Values, oidcValueOpts.Values...),
		FileValues:    append(defaultValueOpts.FileValues, oidcValueOpts.FileValues...),
		JSONValues:    append(defaultValueOpts.JSONValues, oidcValueOpts.JSONValues...),
		LiteralValues: append(defaultValueOpts.LiteralValues, oidcValueOpts.LiteralValues...),
	}

	mergedValues, err := combinedValueOpts.MergeValues(getter.All(settings))
	if err != nil {
		return nil, err
	}
	return mergedValues, nil
}
