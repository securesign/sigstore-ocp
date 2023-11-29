package helm

import (
	"os"
	"os/exec"
)

func InstallTrustedArtifactSigner(commonName string) error {
	executeCommand := func(cmd *exec.Cmd) error {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			return err
		}

		if err := cmd.Wait(); err != nil {
			return err
		}

		return nil
	}

	installCmd := exec.Command("sh", "-c", "envsubst < examples/values-sigstore-openshift.yaml | helm upgrade -i trusted-artifact-signer --debug charts/trusted-artifact-signer  -n trusted-artifact-signer --create-namespace --values -")
	installCmd.Env = append(os.Environ(), "OPENSHIFT_APPS_SUBDOMAIN="+commonName)
	if err := executeCommand(installCmd); err != nil {
		return err
	}

	return nil
}
