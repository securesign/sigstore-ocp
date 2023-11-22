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

	// Create fulcio-system namespace
	if err := kubernetes.CreateNamespace("fulcio-system"); err != nil {
		log.Fatalf("Failed to create fulcio-system namespace: %v", err)
	}

	if err := secrets.ConfigureSystemSecrets("fulcio-system", "fulcio-secret-rh", map[string]string{"password": certs.CertPassword}, map[string]string{"private": "./keys-cert/file_ca_key.pem", "public": "./keys-cert/file_ca_pub.pem", "cert": "./keys-cert/fulcio-root.pem"}); err != nil {
		log.Fatalf("Failed to create secrets: %v", err)
	}

	// Create rekor-system namespace
	if err := kubernetes.CreateNamespace("rekor-system"); err != nil {
		log.Fatalf("Failed to create rekor-system namespace: %v", err)
	}

	if err := secrets.ConfigureSystemSecrets("rekor-system", "rekor-private-key", nil, map[string]string{"private": "./keys-cert/rekor_key.pem"}); err != nil {
		log.Fatalf("Failed to create secrets: %v", err)
	}
}
