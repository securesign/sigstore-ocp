package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"

	"github.com/spf13/cobra"
)

var (
	kc             *kubernetes.KubernetesClient
	kubeconfigPath string
)

var rootCmd = &cobra.Command{
	Use:   "tas-installer",
	Short: "Allows for easy installation of TAS",
	Long:  `The tas-installer cli tool allows for easy installation of the Trusted Artifact Signer stack.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		kc, err = kubernetes.InitKubeClient(kubeconfigPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding user home directory: %v\n", err)
		os.Exit(1)
	}
	defaultKubeConfigPath := filepath.Join(homeDir, ".kube/config")
	rootCmd.PersistentFlags().StringVar(&kubeconfigPath, "kubeconfig", defaultKubeConfigPath, "Specify the kubeconfig path")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
