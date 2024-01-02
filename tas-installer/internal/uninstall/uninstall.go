package uninstall

import (
	"fmt"

	"securesign/sigstore-ocp/tas-installer/pkg/helm"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
)

func HandleHelmChartUninstall(kc *kubernetes.KubernetesClient, tasNamespace, tasReleaseName string) (string, error) {
	info, err := helm.UninstallTrustedArtifactSigner(kc, tasNamespace, tasReleaseName)
	if err != nil {
		return "", err
	}
	msg := fmt.Sprintf("Uninstalled helm release: %s namespace: %s %s\n", info.Release.Name, info.Release.Namespace, info.Info)
	return msg, nil
}

func HandleNamespacesDelete(kc *kubernetes.KubernetesClient, namespaces []string) ([]string, error) {
	deletens := []string{}
	for _, ns := range namespaces {
		deleted, err := kc.DeleteNamespaceIfExists(ns)
		if err != nil {
			return deletens, err
		}
		if deleted {
			deletens = append(deletens, ns)
		}
	}
	return deletens, nil
}
