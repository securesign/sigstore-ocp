package certs

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
	"strings"
	"syscall"

	"golang.org/x/term"
)

var (
	CertPassword = ""
)

func SetupCerts(kc *kubernetes.KubernetesClient) error {
	orgName, email, commonName, password, err := promptForCertInfo(kc)
	if err != nil {
		return err
	}

	err = os.MkdirAll("./keys-cert", 0755)
	if err != nil {
		return err
	}

	commands := []string{
		"openssl ecparam -genkey -name prime256v1 -noout -out keys-cert/unenc.key",
		"openssl ec -in keys-cert/unenc.key -out keys-cert/file_ca_key.pem -des3 -passout pass:" + password,
		"openssl ec -in keys-cert/file_ca_key.pem -passin pass:" + password + " -pubout -out keys-cert/file_ca_pub.pem",
		"openssl req -new -x509 -days 365 -key keys-cert/file_ca_key.pem -passin pass:" + password + " -out keys-cert/fulcio-root.pem -subj \"/CN=" + commonName + "/emailAddress=" + email + "/O=" + orgName + "\"",
		"openssl ecparam -name prime256v1 -genkey -noout -out keys-cert/rekor_key.pem",
	}

	for _, cmd := range commands {
		err := executeCommand(cmd)
		if err != nil {
			log.Fatalf("command failed: %v", err)
		}
	}

	return nil
}

func executeCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func promptForCertInfo(kc *kubernetes.KubernetesClient) (string, string, string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter the organization name for the certificate: ")
	orgName, err := reader.ReadString('\n')
	if err != nil {
		return "", "", "", "", fmt.Errorf("error reading organization name: %v", err)
	}
	orgName = strings.TrimSpace(orgName)

	fmt.Print("Enter the email address for the certificate: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return "", "", "", "", fmt.Errorf("error reading email address: %v", err)
	}
	email = strings.TrimSpace(email)

	fmt.Print("Enter the password for the private key: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", "", "", fmt.Errorf("\nerror reading password: %v", err)
	}

	password := string(bytePassword)
	CertPassword = password

	fmt.Println("\nOrganization Name:", orgName)
	fmt.Println("Email Address:", email)
	fmt.Println("Common Name (CN):", kc.ClusterCommonName)

	return orgName, email, kc.ClusterCommonName, password, nil
}
