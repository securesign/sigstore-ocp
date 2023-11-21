package main

import (
	"log"
	"securesign/sigstore-ocp/tas-install/pkg/keycloak"
	"securesign/sigstore-ocp/tas-install/pkg/kubernetes"
)

func main() {
	// Initialize the Kubernetes client
	err := kubernetes.InitKubeClient()
	if err != nil {
		log.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}

	if err := keycloak.InstallSSOKeycloak(); err != nil {
		log.Fatalf("Failed to install keycloak: %v", err)
	}
}
