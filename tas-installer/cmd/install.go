package cmd

import (
	"fmt"
	"log"
	"securesign/sigstore-ocp/tas-installer/internal/install"
	"securesign/sigstore-ocp/tas-installer/pkg/certs"
	"securesign/sigstore-ocp/tas-installer/pkg/secrets"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs Trusted Artifact Signer",
	Long: `Installs Trusted Artifact Signer (TAS) on a Kubernetes cluster.

	This command performs a series of actions:
	1. Initializes the Kubernetes client to interact with your cluster
	2. Installs Keycloak for SSO
	3. Sets up necessary certificates
	4. Configures secrets
	5. Deploys TAS to openshift`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := installTas(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func installTas() error {
	installSteps := []func() error{
		func() error {
			return install.HandleKeycloakInstall(kc, "keycloak/operator/base", "keycloak/resources/base")
		},
		func() error { return install.HandleCertSetup(kc) },
		func() error {
			return install.DeleteSegmentBackupJobIfExists(kc, "sigstore-monitoring", "segment-backup-job")
		},
		func() error { return install.HandleNamespaceCreate(kc, "sigstore-monitoring") },
		func() error { return install.HandlePullSecretSetup(kc, "pull-secret", "sigstore-monitoring") },
		func() error { return install.HandleNamespaceCreate(kc, "fulcio-system") },
		func() error {
			return secrets.ConfigureSystemSecrets(kc, "fulcio-system", "fulcio-secret-rh", getFulcioLiteralSecrets(), getFulcioFileSecrets())
		},
		func() error { return install.HandleNamespaceCreate(kc, "rekor-system") },
		func() error {
			return secrets.ConfigureSystemSecrets(kc, "rekor-system", "rekor-private-key", nil, getRekorSecrets())
		},
		func() error { return install.HandleHelmChartInstall(kc.ClusterCommonName) },
	}
	for _, step := range installSteps {
		if err := step(); err != nil {
			return fmt.Errorf("install step failed: %v", err)
		}
	}
	return nil
}

func getFulcioFileSecrets() map[string]string {
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

func getRekorSecrets() map[string]string {
	return map[string]string{
		"private": "./keys-cert/rekor_key.pem",
	}
}
