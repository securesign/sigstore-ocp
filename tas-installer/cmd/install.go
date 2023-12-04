package cmd

import (
	"fmt"
	"log"
	"securesign/sigstore-ocp/tas-installer/internal/install"
	"securesign/sigstore-ocp/tas-installer/pkg/certs"
	"securesign/sigstore-ocp/tas-installer/pkg/secrets"

	"github.com/spf13/cobra"
)

var (
	helmChartVersion string
	helmValuesFile   string
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs Trusted Artifact Signer",
	Long: `Installs Trusted Artifact Signer (TAS) on a Kubernetes cluster.

	This command performs a series of actions:
	1. Initializes the Kubernetes client to interact with your cluster
	2. Sets up necessary certificates
	3. Configures secrets
	4. Deploys TAS to openshift`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := installTas(tasNamespace); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func installTas(tasNamespace string) error {
	installSteps := []func() error{
		func() error { return install.HandleCertSetup(kc) },
		func() error { return install.HandleNamespacesCreate(kc, tasNamespacesAll) },
		func() error {
			return install.DeleteSegmentBackupJobIfExists(kc, monitoringNamespace, segmentBackupJob)
		},
		func() error { return install.HandlePullSecretSetup(kc, pullSecret, monitoringNamespace) },
		func() error {
			return secrets.ConfigureSystemSecrets(kc, fulcioNamespace, fulcioCertSecretName, getFulcioLiteralSecrets(), getFulcioSecretFiles())
		},
		func() error {
			return secrets.ConfigureSystemSecrets(kc, rekorNamespace, rekorPrivateKey, nil, getRekorSecretFiles())
		},
		func() error {
			return install.HandleHelmChartInstall(kc, tasNamespace, tasReleaseName, helmValuesFile, helmChartVersion)
		},
	}
	for _, step := range installSteps {
		if err := step(); err != nil {
			return fmt.Errorf("install step failed: %v", err)
		}
	}
	return nil
}

func init() {
	installCmd.PersistentFlags().StringVar(&helmChartVersion, "chartVersion", "0.1.24", "Version of the Helm chart")
	installCmd.PersistentFlags().StringVar(&helmValuesFile, "valuesFile", "", "Custom values file for chart configuration")
}

func getFulcioSecretFiles() map[string]string {
	return map[string]string{
		"private": "./keys-cert/file_ca_key.pem",
		"public":  "./keys-cert/file_ca_pub.pem",
		"cert":    "./keys-cert/fulcio-root.pem",
	}
}

func getFulcioLiteralSecrets() map[string]string {
	return map[string]string{
		"password": certs.GetCertPassword(),
	}
}

func getRekorSecretFiles() map[string]string {
	return map[string]string{
		"private": "./keys-cert/rekor_key.pem",
	}
}
