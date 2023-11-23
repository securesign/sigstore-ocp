package keycloak

import (
	"fmt"
	"os/exec"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
)

func InstallSSOKeycloak(kc *kubernetes.KubernetesClient) error {

	fmt.Println("Installing keycloak operator")
	if err := applyKustomize("keycloak/operator/base"); err != nil {
		return err
	}
	if err := kc.WaitForPodStatusRunning("keycloak-system", "rhsso-operator"); err != nil {
		return err
	}

	fmt.Println("Installing keycloak resources")
	if err := applyKustomize("keycloak/resources/base"); err != nil {
		return err
	}
	if err := kc.WaitForPodStatusRunning("keycloak-system", "keycloak-postgresql"); err != nil {
		return err
	}
	fmt.Println("Keycloak has successfully been installed")

	if err := kc.WaitForPodStatusRunning("keycloak-system", "keycloak"); err != nil {
		return err
	}
	fmt.Println("Keycloak is up and running")

	return nil
}

func applyKustomize(path string) error {
	cmd := exec.Command("oc", "apply", "--kustomize", path)
	err := cmd.Run()
	return err
}
