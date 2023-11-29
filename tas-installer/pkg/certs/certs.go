package certs

import (
	"log"
	"os"
	"os/exec"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
	"securesign/sigstore-ocp/tas-installer/ui"
)

var (
	certPassword string
)

func SetupCerts(kc *kubernetes.KubernetesClient) error {
	certConfig, err := ui.PromptForCertInfo(kc)
	if err != nil {
		return err
	}
	certPassword = certConfig.CertPassword

	err = os.MkdirAll("./keys-cert", 0755)
	if err != nil {
		return err
	}

	commands := []string{
		"openssl ecparam -genkey -name prime256v1 -noout -out keys-cert/unenc.key",
		"openssl ec -in keys-cert/unenc.key -out keys-cert/file_ca_key.pem -des3 -passout pass:" + certConfig.CertPassword,
		"openssl ec -in keys-cert/file_ca_key.pem -passin pass:" + certConfig.CertPassword + " -pubout -out keys-cert/file_ca_pub.pem",
		"openssl req -new -x509 -days 365 -key keys-cert/file_ca_key.pem -passin pass:" + certConfig.CertPassword + " -out keys-cert/fulcio-root.pem -subj \"/CN=" + certConfig.ClusterCommonName + "/emailAddress=" + certConfig.OrganizationEmail + "/O=" + certConfig.OrganizationName + "\"",
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

func GetCertPassword() string {
	return certPassword
}