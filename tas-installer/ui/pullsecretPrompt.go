package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func PromptForPullSecretPath() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter the absolute path to the pull-secret.json file: ")
	filePath, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(filePath), nil
}

func PromptForPullSecretOverwrite(secretName, namespace string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("Secret %s in namespace %s already exists. Overwrite it (Y/N)?: ", secretName, namespace)
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
