package cmd

import (
	"log"
	"securesign/sigstore-ocp/tas-installer/internal/uninstall"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Removes installation of Trusted Artifact Signer",
	Long:  `Removes installation of Trusted Artifact Signer (TAS) on a Kubernetes cluster.`,

	Run: func(cmd *cobra.Command, args []string) {
		if err := uninstallTas(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func uninstallTas() error {
	if err := uninstall.HandleHelmChartUninstall(tasNamespace, tasReleaseName); err != nil {
		log.Print(err.Error())
	}
	if err := uninstall.HandleNamespacesDelete(kc, tasNamespacesAll); err != nil {
		return err
	}
	return nil
}
