package cmd

import (
	"fmt"
	"log"
	"securesign/sigstore-ocp/tas-installer/pkg/certs"
	"securesign/sigstore-ocp/tas-installer/pkg/helm"
	"securesign/sigstore-ocp/tas-installer/pkg/keycloak"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
	"securesign/sigstore-ocp/tas-installer/pkg/secrets"
	"securesign/sigstore-ocp/tas-installer/ui"

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
		func() error { return handleKeycloakInstall(kc, "keycloak/operator/base", "keycloak/resources/base") },
		func() error { return handleCertSetup(kc) },
		func() error { return deleteSegmentBackupJobIfExists(kc, "sigstore-monitoring", "segment-backup-job") },
		func() error { return handleNamespaceCreate(kc, "sigstore-monitoring") },
		func() error { return secrets.ConfigurePullSecret(kc, "pull-secret", "sigstore-monitoring") },
		func() error { return handleNamespaceCreate(kc, "fulcio-system") },
		func() error {
			return secrets.ConfigureSystemSecrets(kc, "fulcio-system", "fulcio-secret-rh", getFulcioLiteralSecrets(), getFulcioFileSecrets())
		},
		func() error { return handleNamespaceCreate(kc, "rekor-system") },
		func() error {
			return secrets.ConfigureSystemSecrets(kc, "rekor-system", "rekor-private-key", nil, getRekorSecrets())
		},
		func() error { return handleHelmChartInstall(kc.ClusterCommonName) },
	}
	for _, step := range installSteps {
		if err := step(); err != nil {
			return fmt.Errorf("install step failed: %v", err)
		}
	}
	return nil
}

func handleKeycloakInstall(kc *kubernetes.KubernetesClient, operatorConfig, resourceConfig string) error {
	fmt.Println("Installing keycloak operator in namespace: 'keycloak-system'")

	if err := keycloak.ApplyAndWaitForKeycloakResources(kc, operatorConfig, "keycloak-system", "rhsso-operator", func(err error) {
		switch {
		case err == kubernetes.ErrPodNotFound:
			fmt.Println("No pods with the prefix 'rhsso-operator' found in namespace keycloak-system. Retrying in 10 seconds...")
		case err == kubernetes.ErrPodNotRunning:
			fmt.Println("Waiting for pod with prefix 'rhsso-operator' to reach a running state...")
		}
	}); err != nil {
		return err
	}
	fmt.Println("Pod with prefix 'rhsso-operator' has reached a running state")

	fmt.Println("Installing keycloak resources in namespace: 'keycloak-system'")
	if err := keycloak.ApplyAndWaitForKeycloakResources(kc, resourceConfig, "keycloak-system", "keycloak-postgresql", func(err error) {
		switch {
		case err == kubernetes.ErrPodNotFound:
			fmt.Println("No pods with the prefix 'keycloak-postgresql' found in namespace keycloak-system. Retrying in 10 seconds...")
		case err == kubernetes.ErrPodNotRunning:
			fmt.Println("Waiting for pod with prefix 'keycloak-postgresql' to reach a running state...")
		}
	}); err != nil {
		return err
	}
	fmt.Println("Pod with prefix 'keycloak-postgresql' has reached a running state")

	fmt.Println("Keycloak installed successfully")
	return nil
}

func handleHelmChartInstall(clusterCommonName string) error {
	if err := helm.InstallTrustedArtifactSigner(clusterCommonName); err != nil {
		return err
	}
	fmt.Println("Helm Chart Successfully installed ")
	return nil
}

func handleNamespaceCreate(kc *kubernetes.KubernetesClient, namespace string) error {
	if err := kc.CreateNamespaceIfExists(namespace); err != nil {
		if err == kubernetes.ErrNamespaceAlreadyExists {
			fmt.Printf("namespace %s already exists skipping create", namespace)
		}
		return err
	}
	fmt.Printf("namespace: %s successfully created \n", namespace)
	return nil
}

func handleCertSetup(kc *kubernetes.KubernetesClient) error {
	certConfig, err := ui.PromptForCertInfo(kc)
	if err != nil {
		return err
	}
	certs.SetupCerts(kc, certConfig)
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

func deleteSegmentBackupJobIfExists(kc *kubernetes.KubernetesClient, namespace, jobName string) error {
	if err := kc.DeleteJobIfExists(namespace, jobName); err != nil {
		return err
	}
	return nil
}
