package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type OIDCConfig struct {
	IssuerURL string
	ClientID  string
	Type      string
}

func PromptForOIDCInfo() (*OIDCConfig, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Please enter the URL of your OIDC provider: ")
	issuerURL, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading OIDC URL: %v", err)
	}
	issuerURL = strings.TrimSpace(issuerURL)

	fmt.Print("Please enter the ClientID of your OIDC provider: ")
	clientID, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading OIDC ClientID: %v", err)
	}
	clientID = strings.TrimSpace(clientID)

	return &OIDCConfig{IssuerURL: issuerURL, ClientID: clientID, Type: "email"}, nil
}

func PromptForDefaultOIDCOption() (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Would you like to configure trusted artifact signer with a custom OIDC provider (Y/N)?: ")
		useCustomOIDC, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}

		useCustomOIDC = strings.TrimSpace(strings.ToLower(useCustomOIDC))

		if useCustomOIDC == "y" {
			return true, nil
		} else if useCustomOIDC == "n" {
			return false, nil
		} else {
			fmt.Println("Invalid input. Please enter 'Y' for Yes or 'N' for No.")
		}
	}
}
