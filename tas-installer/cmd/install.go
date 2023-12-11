package cmd

import (
	"fmt"
	"log"
	"path/filepath"

	"securesign/sigstore-ocp/tas-installer/internal/install"
	"securesign/sigstore-ocp/tas-installer/pkg/certs"
	"securesign/sigstore-ocp/tas-installer/pkg/oidc"
	"securesign/sigstore-ocp/tas-installer/pkg/secrets"

	"github.com/spf13/cobra"
)

const (
	keysCertDir = "keys-cert"
)

var (
	helmChartLocation string
	helmChartVersion  string
	helmValuesFile    string
	oidcConfig        oidc.OIDCConfig
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
			if err := install.HandleHelmChartInstall(kc, oidcConfig, tasNamespace, tasReleaseName, helmValuesFile, helmChartLocation, helmChartVersion); err != nil {
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
	installCmd.PersistentFlags().StringVar(&helmChartVersion, "chart-version", "0.1.29", "Version of the Helm chart")
	installCmd.PersistentFlags().StringVar(&helmChartLocation, "chart-location", "./charts/trusted-artifact-signer", "/local/path/to/chart or oci://registry/repo location of Helm chart")
	installCmd.PersistentFlags().StringVar(&helmValuesFile, "values", "", "path to custom values file for chart configuration")
	installCmd.PersistentFlags().StringVar(&oidcConfig.IssuerURL, "oidc-issuer-url", "", "Specify the OIDC issuer URL e.g for keycloak: https://[keycloak-domain]/auth/realms/[realm-name]")
	installCmd.PersistentFlags().StringVar(&oidcConfig.ClientID, "oidc-client-id", "", "Specify the OIDC client ID")
	installCmd.PersistentFlags().StringVar(&oidcConfig.Type, "oidc-type", "", "Specify the OIDC type")
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
