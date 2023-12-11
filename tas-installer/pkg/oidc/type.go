package oidc

// defines the type for the OIDC provider
type OIDCConfig struct {
	IssuerURL string
	ClientID  string
	Type      string
}
