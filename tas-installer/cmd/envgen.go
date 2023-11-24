package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var envgenCmd = &cobra.Command{
	Use:   "envgen",
	Short: "Generates a script to define env vars to communicate with TAS",
	Long: `The 'envgen' command will generate a script which will define the following Environmental Variables that will allow you to communicate with the TAS stack
	
	Env Vars Generated:
	1. BASE_HOSTNAME=apps.$(oc get dns cluster -o jsonpath='{ .spec.baseDomain }')
	2. KEYCLOAK_REALM=sigstore
	3. FULCIO_URL=https://fulcio.\$BASE_HOSTNAME
	4. KEYCLOAK_URL=https://keycloak-keycloak-system.\$BASE_HOSTNAME
	5. REKOR_URL=https://rekor.\$BASE_HOSTNAME
	6. TUF_URL=https://tuf.\$BASE_HOSTNAME
	7. OIDC_ISSUER_URL=\$KEYCLOAK_URL/auth/realms/\$KEYCLOAK_REALM`,

	Run: func(cmd *cobra.Command, args []string) {
		err := generateEnvVars()
		if err != nil {
			log.Fatal("Failed to generate env vars")
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
