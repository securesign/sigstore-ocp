package secrets

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"securesign/sigstore-ocp/tas-install/pkg/kubernetes"
	"strings"
)

func ConfigurePullSecret(pullSecretName, namespace string) error {
	secretExistsInCluster, err := kubernetes.PullSecretExists(pullSecretName, namespace)
	if err != nil {
		return err
	}

	if secretExistsInCluster {
		overWrite, err := promptForSecretOverwrite(pullSecretName, namespace)
		if err != nil {
			return err
		}

		if overWrite {
			err := handleSecretCreation(pullSecretName, namespace)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("Skipping overwriting pull-secret...")
			return nil
		}

	} else {
		err := handleSecretCreation(pullSecretName, namespace)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Secret: %s created successfully\n", pullSecretName)
	return nil
}

func handleSecretCreation(pullSecretName, namespace string) error {
	secretPath, err := promptForSecretPath()
	if err != nil {
		return err
	}
	secretFileExists := pullSecretFileExists(secretPath)
	if secretFileExists {
		secretData, err := readSecretFile(secretPath)
		if err != nil {
			return err
		}
		fileName := filepath.Base(secretPath)
		err = kubernetes.CreatePullSecret(pullSecretName, namespace, fileName, secretData)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("pull secret was not found based on the path provided: %s\n", secretPath)
	}
	return nil
}

func promptForSecretPath() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter the absolute path to the pull-secret.json file:")
	filePath, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(filePath), nil
}

func promptForSecretOverwrite(secretName, namespace string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("Secret %s in namespace %s already exists. Overwrite it (Y/N)?: \n", secretName, namespace)
		overwrite, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}

		overwrite = strings.TrimSpace(strings.ToLower(overwrite))

		if overwrite == "y" {
			return true, nil
		} else if overwrite == "n" {
			return false, nil
		} else {
			fmt.Println("Invalid input. Please enter 'Y' for Yes or 'N' for No.")
		}
	}
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
