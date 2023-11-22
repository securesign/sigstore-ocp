package main

import (
	"log"
	"securesign/sigstore-ocp/tas-install/pkg/certs"
	"securesign/sigstore-ocp/tas-install/pkg/keycloak"
	"securesign/sigstore-ocp/tas-install/pkg/kubernetes"
	"securesign/sigstore-ocp/tas-install/pkg/secrets"
)

func main() {
	// Initialize the Kubernetes client
	err := kubernetes.InitKubeClient()
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}

	// Install keycloak
	if err := keycloak.InstallSSOKeycloak(); err != nil {
		log.Fatalf("Failed to install keycloak: %v", err)
	}

	// Setup certs
	if err := certs.SetupCerts(); err != nil {
		log.Fatalf("Failed to setup certs: %v", err)
	}

	// Create sigstore-monitoring namespace
	if err := kubernetes.CreateNamespace("sigstore-monitoring"); err != nil {
		log.Fatalf("Failed to create sigstore-monitoring namespace: %v", err)
	}

	// Configure Pull Secret
	if err := secrets.ConfigurePullSecret("pull-secret", "sigstore-monitoring"); err != nil {
		log.Fatalf("Failed to create pull secret: %v", err)
	}
}
