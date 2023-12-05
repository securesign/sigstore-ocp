package install

import (
	"context"
	"fmt"
	"path/filepath"
	"securesign/sigstore-ocp/tas-installer/pkg/certs"
	"securesign/sigstore-ocp/tas-installer/pkg/helm"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
	"securesign/sigstore-ocp/tas-installer/pkg/secrets"
	"securesign/sigstore-ocp/tas-installer/ui"
	"time"
)

var OIDCConfig ui.OIDCConfig

func HandleHelmChartInstall(kc *kubernetes.KubernetesClient, helmValuesFile, helmChartVersion string) error {
	fmt.Println("Installing helm chart")
	if err := helm.InstallTrustedArtifactSigner(kc, helmValuesFile, helmChartVersion, OIDCConfig); err != nil {
		return err
	}
	fmt.Println("Helm Chart Successfully installed")
	return nil
}

func HandleNamespaceCreate(kc *kubernetes.KubernetesClient, namespace string) error {
	if err := kc.CreateNamespaceIfNotExists(namespace); err != nil {
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

func HandleOIDCInfo() error {
	useCustomOIDC, err := ui.PromptForDefaultOIDCOption()
	if err != nil {
		return err
	}

	if useCustomOIDC {
		config, err := ui.PromptForOIDCInfo()
		if err != nil {
			return err
		}
		OIDCConfig = *config
	} else {
		OIDCConfig.ClientID = "sigstore"
		OIDCConfig.IssuerURL = "https://oauth2.sigstore.dev/auth"
		OIDCConfig.Type = "email"
		fmt.Println("Using default OIDC provider")
	}
	return nil
}
