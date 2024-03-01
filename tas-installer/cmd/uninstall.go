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
	log.Print("uninstalling helm chart")
	msg, err := uninstall.HandleHelmChartUninstall(kc, tasNamespace, tasReleaseName)
	if err != nil {
		log.Print(err.Error())
	} else {
		log.Print(msg)
	}
	deletens, err := uninstall.HandleNamespacesDelete(kc, tasNamespacesAll)
	if err != nil {
		return err
	}
	for _, ns := range deletens {
		log.Printf("namespace: %s successfully deleted", ns)
	}
	return nil
}
