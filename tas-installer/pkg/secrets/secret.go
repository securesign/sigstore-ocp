package secrets

import (
	"fmt"
	"os"
	"path/filepath"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ConfigureSystemSecrets(kc *kubernetes.KubernetesClient, namespace, secretName string, literals, filepaths map[string]string) error {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: make(map[string][]byte),
	}

	for key, filePath := range filepaths {
		secretData, _, err := processSecretFile(filePath)
		if err != nil {
			return err
		}
		secret.Data[key] = secretData
	}

	for key, value := range literals {
		secret.Data[key] = []byte(value)
	}

	err := kc.CreateSecretIfNotExists(secretName, namespace, secret)
	if err != nil {
		return err
	}
	return nil
}

func OverwritePullSecret(kc *kubernetes.KubernetesClient, pullSecretName, namespace, pullSecretPath string) error {
	secretData, fileName, err := processSecretFile(pullSecretPath)
	if err != nil {
		return err
	}
	err = kc.UpdateSecretData(pullSecretName, namespace, fileName, secretData)
	return nil
}

func processSecretFile(secretPath string) ([]byte, string, error) {
	if secretPath == "" {
		return nil, "", fmt.Errorf("no secret path provided")
	}

	secretFileExists := pullSecretFileExists(secretPath)
	if !secretFileExists {
		return nil, "", fmt.Errorf("pull secret was not found based on the path provided: %s", secretPath)
	}

	secretData, err := readSecretFile(secretPath)
	if err != nil {
		return nil, "", err
	}

	fileName := filepath.Base(secretPath)
	return secretData, fileName, nil
}

func pullSecretFileExists(secretFilePath string) bool {
	if _, err := os.Stat(secretFilePath); err == nil {
		return true
	}
	return false
}

func readSecretFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}
