package cmd

import (
	"fmt"
	"log"
	"securesign/sigstore-ocp/tas-installer/pkg/certs"
	"securesign/sigstore-ocp/tas-installer/pkg/helm"
	"securesign/sigstore-ocp/tas-installer/pkg/keycloak"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
	"securesign/sigstore-ocp/tas-installer/pkg/secrets"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs Trusted Artifact Signer",
	Long: `The 'install' command is designed to set up the Trusted Artifact Signer (TAS).

	This command performs a series of actions:
	1. Initializes the Kubernetes client to interact with your cluster.
	2. Installs Keycloak for SSO.
	3. Sets up necessary certificates.
	4. Configures secrets.
	5. Deploys TAS to openshift`,

	Run: func(cmd *cobra.Command, args []string) {
		err := installTas()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func installTas() error {
	installSteps := []func() error{
		kubernetes.InitKubeClient,
		keycloak.InstallSSOKeycloak,
		certs.SetupCerts,
		func() error { return kubernetes.CreateNamespace("sigstore-monitoring") },
		func() error { return secrets.ConfigurePullSecret("pull-secret", "sigstore-monitoring") },
		func() error { return kubernetes.CreateNamespace("fulcio-system") },
		func() error {
			return secrets.ConfigureSystemSecrets("fulcio-system", "fulcio-secret-rh", getFulcioLiteralSecrets(), getFulcioFileSecrets())
		},
		func() error { return kubernetes.CreateNamespace("rekor-system") },
		func() error {
			return secrets.ConfigureSystemSecrets("rekor-system", "rekor-private-key", nil, getRekorSecrets())
		},
		func() error { return helm.InstallTrustedArtifactSigner(certs.CommonName) },
	}
	for _, step := range installSteps {
		if err := step(); err != nil {
			return fmt.Errorf("Installation step failed: %v", err)
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
		"password": certs.CertPassword,
	}
}

func getRekorSecrets() map[string]string {
	return map[string]string{
		"private": "./keys-cert/rekor_key.pem",
	}
}
