package cmd

import (
	"fmt"
	"log"
	"os"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"

	"github.com/spf13/cobra"
)

var envVarsCmd = &cobra.Command{
	Use:   "envVars",
	Short: "Generates env vars to communicate with TAS",
	Long: `The 'envVars' command will generate the following Environmental Variables that will allow you to communicate with the TAS stack
	
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
	rootCmd.AddCommand(envVarsCmd)
}

func generateEnvVars() error {

	kc, err := kubernetes.InitKubeClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Kubernetes client: %v", err)
	}

	keycloakRealm := "sigstore"
	envVars := map[string]string{
		"BASE_HOSTNAME":   kc.ClusterCommonName,
		"KEYCLOAK_REALM":  keycloakRealm,
		"FULCIO_URL":      "https://fulcio." + kc.ClusterCommonName,
		"KEYCLOAK_URL":    "https://keycloak-keycloak-system." + kc.ClusterCommonName,
		"REKOR_URL":       "https://rekor." + kc.ClusterCommonName,
		"TUF_URL":         "https://tuf." + kc.ClusterCommonName,
		"OIDC_ISSUER_URL": "https://" + "keycloak-keycloak-system." + kc.ClusterCommonName + "/auth/realms/" + keycloakRealm,
	}

	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set env var %s: %w", key, err)
		}
		fmt.Printf("Set %s=%s\n", key, value)
	}

	return nil
}
