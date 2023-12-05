package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"

	"github.com/spf13/cobra"
)

const (
	fulcioNamespace      = "fulcio-system"
	rekorNamespace       = "rekor-system"
	monitoringNamespace  = "trusted-artifact-signer-monitoring"
	tasNamespace         = "trusted-artifact-signer"
	tasReleaseName       = "trusted-artifact-signer"
	fulcioCertSecretName = "fulcio-secret-rh"
	rekorPrivateKey      = "rekor-private-key"
	pullSecret           = "pull-secret"
	segmentBackupJob     = "segment-backup-job"
)

var (
	kc             *kubernetes.KubernetesClient
	kubeconfigPath string
	// tasNamespacesAll are those not managed by Helm
	tasNamespacesAll = []string{fulcioNamespace, rekorNamespace, monitoringNamespace, tasNamespace}
)

var rootCmd = &cobra.Command{
	Use:   "tas-installer",
	Short: "Installer for Red Hat Trusted Artifact Signer (TAS) on Kubernetes",
	Long: `Installs Red Hat Trusted Artifact Signer (TAS) on Kubernetes
	
	For a successful installation, you must have provide the path to a kubeconfig file, or have 
	one in $HOME/.kube/config. Additionally, the following CLI tools must all be in your $PATH environment.
	
	oc - used to install Keycloak `,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		kc, err = kubernetes.InitKubeClient(kubeconfigPath)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Using kube config found at %s\n", kubeconfigPath)
		return nil
	},
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	defaultKubeConfigPath := filepath.Join(homeDir, ".kube/config")
	rootCmd.PersistentFlags().StringVar(&kubeconfigPath, "kubeconfig", defaultKubeConfigPath, "Specify the kubeconfig path")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
