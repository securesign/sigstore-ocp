package keycloak

import (
	"fmt"
	"os/exec"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
)

func InstallSSOKeycloak(kc *kubernetes.KubernetesClient, namespace string) error {
	fmt.Printf("Installing keycloak operator in namespace %s \n", namespace)

	if err := applyKustomize("keycloak/operator/base"); err != nil {
		return err
	}
	if err := kc.WaitForPodStatusRunning(namespace, "rhsso-operator"); err != nil {
		return err
	}

	fmt.Printf("Installing keycloak resources in namespace %s \n", namespace)
	if err := applyKustomize("keycloak/resources/base"); err != nil {
		return err
	}
	if err := kc.WaitForPodStatusRunning(namespace, "keycloak-postgresql"); err != nil {
		return err
	}
	fmt.Println("Keycloak has successfully been installed")

	if err := kc.WaitForPodStatusRunning(namespace, "keycloak"); err != nil {
		return err
	}

	return nil
}

func applyKustomize(path string) error {
	cmd := exec.Command("oc", "apply", "--kustomize", path)
	err := cmd.Run()
	return err
}
