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
		func() error { return keycloak.InstallSSOKeycloak(kc, "keycloak-system") },
		func() error { return certs.SetupCerts(kc) },
		func() error { return checkSegmentBackupJob(kc, "sigstore-monitoring", "segment-backup-job") },
		func() error { return kc.CreateNamespace("sigstore-monitoring") },
		func() error { return secrets.ConfigurePullSecret(kc, "pull-secret", "sigstore-monitoring") },
		func() error { return kc.CreateNamespace("fulcio-system") },
		func() error {
			return secrets.ConfigureSystemSecrets(kc, "fulcio-system", "fulcio-secret-rh", getFulcioLiteralSecrets(), getFulcioFileSecrets())
		},
		func() error { return kc.CreateNamespace("rekor-system") },
		func() error {
			return secrets.ConfigureSystemSecrets(kc, "rekor-system", "rekor-private-key", nil, getRekorSecrets())
		},
		func() error { return helm.InstallTrustedArtifactSigner(kc.ClusterCommonName) },
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
		"password": certs.GetCertPassword(),
	}
}

func getRekorSecrets() map[string]string {
	return map[string]string{
		"private": "./keys-cert/rekor_key.pem",
	}
}

func checkSegmentBackupJob(kc *kubernetes.KubernetesClient, namespace, jobName string) error {
	job, err := kc.GetJob(namespace, jobName)
	if err != nil {
		return err
	}

	if job != nil {
		err := kc.DeleteJob(namespace, jobName)
		if err != nil {
			return err
		}
	}
	return nil
}
