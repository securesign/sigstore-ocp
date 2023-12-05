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

func HandleHelmChartInstall(kc *kubernetes.KubernetesClient, tasNamespace, tasReleaseName, helmValuesFile, helmChartVersion string) error {
	if err := helm.InstallTrustedArtifactSigner(kc, tasNamespace, tasReleaseName, helmValuesFile, helmChartVersion); err != nil {
		return err
	}
	return nil
}

func HandleNamespacesCreate(kc *kubernetes.KubernetesClient, namespaces []string) ([]string, error) {
	createns := []string{}
	for _, ns := range namespaces {
		if err := kc.CreateNamespaceIfNotExists(ns); err != nil {
			if err != kubernetes.ErrNamespaceAlreadyExists {
				return createns, err
			}
		}
		createns = append(createns, ns)
	}
	return createns, nil
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
