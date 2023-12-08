package cmd

import (
	"fmt"
	"log"
	"path/filepath"

	"securesign/sigstore-ocp/tas-installer/internal/install"
	"securesign/sigstore-ocp/tas-installer/pkg/certs"
	"securesign/sigstore-ocp/tas-installer/pkg/secrets"

	"github.com/spf13/cobra"
)

const (
	keysCertDir = "keys-cert"
)

var (
	helmChartVersion string
	helmValuesFile   string
	helmChartUrl     = "./charts/trusted-artifact-signer"
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
		func() error { return install.HandleCertSetup(kc, keysCertDir) },
		func() error {
			createns, err := install.HandleNamespacesCreate(kc, tasNamespacesAll)
			if err != nil {
				return err
			}
			for _, ns := range createns {
				log.Printf("namespace: %s successfully created", ns)
			}
			return nil
		},
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
			log.Print("installing helm chart")
			if err := install.HandleHelmChartInstall(kc, tasNamespace, tasReleaseName, helmValuesFile, helmChartUrl, helmChartVersion); err != nil {
				return err
			}
			return nil
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
	installCmd.PersistentFlags().StringVar(&helmChartVersion, "chartVersion", helmChartVersion, "Version of the Helm chart")
	installCmd.PersistentFlags().StringVar(&helmValuesFile, "valuesFile", "", "Custom values file for chart configuration")
	installCmd.PersistentFlags().StringVar(&helmChartUrl, "chartUrl", helmChartUrl, "URL to Trusted Artifact Signer Helm chart")
}

func getFulcioSecretFiles() map[string]string {
	return map[string]string{
		"private": filepath.Join(keysCertDir, certs.FulcioPrivateKey),
		"public":  filepath.Join(keysCertDir, certs.FulcioPublicKey),
		"cert":    filepath.Join(keysCertDir, certs.FulcioRootCert),
	}
}

func getFulcioLiteralSecrets() map[string]string {
	return map[string]string{
		"password": certs.GetCertPassword(),
	}
}

func getRekorSecretFiles() map[string]string {
	return map[string]string{
		"private": filepath.Join(keysCertDir, certs.RekorSigningKey),
	}
}
