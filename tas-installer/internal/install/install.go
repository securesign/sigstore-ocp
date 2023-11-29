package install

import (
	"context"
	"fmt"
	"path/filepath"
	"securesign/sigstore-ocp/tas-installer/pkg/certs"
	"securesign/sigstore-ocp/tas-installer/pkg/helm"
	"securesign/sigstore-ocp/tas-installer/pkg/keycloak"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
	"securesign/sigstore-ocp/tas-installer/pkg/secrets"
	"securesign/sigstore-ocp/tas-installer/ui"
	"time"
)

func HandleKeycloakInstall(kc *kubernetes.KubernetesClient, operatorConfig, resourceConfig string) error {
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

func HandleHelmChartInstall(clusterCommonName string) error {
	if err := helm.InstallTrustedArtifactSigner(clusterCommonName); err != nil {
		return err
	}
	fmt.Println("Helm Chart Successfully installed ")
	return nil
}

func HandleNamespaceCreate(kc *kubernetes.KubernetesClient, namespace string) error {
	if err := kc.CreateNamespaceIfExists(namespace); err != nil {
		if err == kubernetes.ErrNamespaceAlreadyExists {
			fmt.Printf("namespace %s already exists skipping create", namespace)
		}
		return err
	}
	fmt.Printf("namespace: %s successfully created \n", namespace)
	return nil
}

func HandlePullSecretSetup(kc *kubernetes.KubernetesClient, pullSecretName, namespace string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secretExistsInCluster, err := kc.SecretExists(ctx, pullSecretName, namespace)
	if err != nil {
		return err
	}

	if secretExistsInCluster {
		overWrite, err := ui.PromptForPullSecretOverwrite(pullSecretName, namespace)
		if err != nil {
			return err
		}

		if overWrite {
			pullSecretPath, err := ui.PromptForPullSecretPath()
			if err != nil {
				return err
			}

			err = secrets.OverwritePullSecret(kc, pullSecretName, namespace, pullSecretPath)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("Skipping secret overwrite")
			return nil
		}

	} else {
		pullSecretPath, err := ui.PromptForPullSecretPath()
		if err != nil {
			return err
		}

		fileName := filepath.Base(pullSecretPath)
		err = secrets.ConfigureSystemSecrets(kc, namespace, pullSecretName, nil, map[string]string{fileName: pullSecretPath})
		if err != nil {
			return err
		}
	}

	return nil
}

func HandleCertSetup(kc *kubernetes.KubernetesClient) error {
	certConfig, err := ui.PromptForCertInfo(kc)
	if err != nil {
		return err
	}
	certs.SetupCerts(kc, certConfig)
	return nil
}

func DeleteSegmentBackupJobIfExists(kc *kubernetes.KubernetesClient, namespace, jobName string) error {
	if err := kc.DeleteJobIfExists(namespace, jobName); err != nil {
		return err
	}
	return nil
}
