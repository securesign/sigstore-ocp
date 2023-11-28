package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var envgenCmd = &cobra.Command{
	Use:   "envgen",
	Short: "Creates a shell script defining configuration environment variables for TAS",
	Long: `Creates a shell script defining configuration environment variables for TAS command line binaries. This script can be used to configure "cosign" and other CLI binaries that communicate with the TAS infrastructure.
	
	Environment Variables:
	1. KEYCLOAK_REALM=sigstore
	2. FULCIO_URL=https://fulcio.\$BASE_HOSTNAME
	3. KEYCLOAK_URL=https://keycloak-keycloak-system.\$BASE_HOSTNAME
	4. REKOR_URL=https://rekor.\$BASE_HOSTNAME
	5. TUF_URL=https://tuf.\$BASE_HOSTNAME
	6. OIDC_ISSUER_URL=\$KEYCLOAK_URL/auth/realms/\$KEYCLOAK_REALM`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := generateEnvVars(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(envgenCmd)
}

func generateEnvVars() error {
	baseHostname := kc.ClusterCommonName

	scriptContent := `#!/bin/bash
	export BASE_HOSTNAME=` + baseHostname + `
	echo "Base hostname = $BASE_HOSTNAME"
	export KEYCLOAK_REALM=sigstore
	export FULCIO_URL=https://fulcio.` + baseHostname + `
	export KEYCLOAK_URL=https://keycloak-keycloak-system.` + baseHostname + `
	export REKOR_URL=https://rekor.` + baseHostname + `
	export TUF_URL=https://tuf.` + baseHostname + `
	export OIDC_ISSUER_URL=https://keycloak-keycloak-system.` + baseHostname + `/auth/realms/sigstore`

	fileName := "tas-env-variables.sh"
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	_, err = file.WriteString(scriptContent)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	err = os.Chmod(fileName, 0755)
	if err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	fmt.Printf("A script '%s' to set environment variables has been created.\n", fileName)
	fmt.Println("To initialize the environment variables, run 'source ./" + fileName + "' from the terminal.")
	return nil
}
