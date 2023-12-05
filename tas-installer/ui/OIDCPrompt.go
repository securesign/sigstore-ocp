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
