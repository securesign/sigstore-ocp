package keycloak

import (
	"os/exec"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
)

func applyKustomize(path string) error {
	cmd := exec.Command("oc", "apply", "--kustomize", path)
	err := cmd.Run()
	return err
}

func ApplyAndWaitForKeycloakResources(kc *kubernetes.KubernetesClient, configFilePath, namespace, podNamePrefix string, status func(error)) error {
	if err := applyKustomize(configFilePath); err != nil {
		return err
	}

	if err := kc.WaitForPodStatusRunning(namespace, podNamePrefix, func(err error) {
		status(err)
	}); err != nil {
		return err
	}

	return nil
}
