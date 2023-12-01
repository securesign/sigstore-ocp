package ui

import (
	"bufio"
	"fmt"
	"os"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
	"strings"
	"syscall"

	"golang.org/x/term"
)

type CertConfig struct {
	OrganizationName  string
	OrganizationEmail string
	ClusterCommonName string
	CertPassword      string
}

func PromptForCertInfo(kc *kubernetes.KubernetesClient) (*CertConfig, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter the organization name for the certificate: ")
	orgName, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading organization name: %v", err)
	}
	orgName = strings.TrimSpace(orgName)

	fmt.Print("Enter the email address for the certificate: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading email address: %v", err)
	}
	email = strings.TrimSpace(email)

	fmt.Print("Enter the password for the private key: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, fmt.Errorf("\nerror reading password: %v", err)
	}

	password := string(bytePassword)
	fmt.Println("\nOrganization Name:", orgName)
	fmt.Println("Email Address:", email)
	fmt.Println("Common Name (CN):", kc.ClusterCommonName)

	return &CertConfig{OrganizationName: orgName, OrganizationEmail: email, ClusterCommonName: kc.ClusterCommonName, CertPassword: password}, nil
}
