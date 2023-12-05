package uninstall

import (
	"fmt"

	"securesign/sigstore-ocp/tas-installer/pkg/helm"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
)

func HandleHelmChartUninstall(tasNamespace, tasReleaseName string) error {
	fmt.Println("Uninstalling helm chart")
	info, err := helm.UninstallTrustedArtifactSigner(tasNamespace, tasReleaseName)
	if err != nil {
		return err
	}
	fmt.Printf("Uninstalled helm release: %s namespace: %s %s\n", info.Release.Name, info.Release.Namespace, info.Info)
	return nil
}

func HandleNamespacesDelete(kc *kubernetes.KubernetesClient, namespaces []string) error {
	for _, ns := range namespaces {
		deleted, err := kc.DeleteNamespaceIfExists(ns)
		if err != nil {
			return err
		}
		if deleted {
			fmt.Printf("namespace: %s successfully deleted \n", ns)
		}
	}
	return nil
}
