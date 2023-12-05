package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"securesign/sigstore-ocp/tas-installer/pkg/kubernetes"
	"securesign/sigstore-ocp/tas-installer/ui"
	"time"
)

var (
	certPassword string
)

func SetupCerts(kc *kubernetes.KubernetesClient, certConfig *ui.CertConfig) error {
	certPassword = certConfig.CertPassword
	err := os.MkdirAll("./keys-cert", 0755)
	if err != nil {
		return err
	}

	cakey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	if err = createCAKey(cakey, certConfig); err != nil {
		return err
	}
	if err = createCAPub(cakey, certConfig); err != nil {
		return err
	}
	if err = createFulcioCA(cakey, certConfig); err != nil {
		return err
	}
	if err = createRekorKey(certConfig); err != nil {
		return err
	}

	return nil
}

func GetCertPassword() string {
	return certPassword
}

func createCAKey(key *ecdsa.PrivateKey, certConfig *ui.CertConfig) error {
	mKey, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}

	block, err := x509.EncryptPEMBlock(rand.Reader, "EC PRIVATE KEY", mKey, []byte(certConfig.CertPassword), x509.PEMCipherAES256)
	if err != nil {
		return err
	}

	file, err := os.Create("./keys-cert/file_ca_key.pem")
	if err != nil {
		return err
	}
	defer file.Close()
	if err = pem.Encode(file, block); err != nil {
		return err
	}
	return nil
}

func createCAPub(key *ecdsa.PrivateKey, certConfig *ui.CertConfig) error {
	mPubKey, err := x509.MarshalPKIXPublicKey(key.Public())
	if err != nil {
		return err
	}

	publicF, err := os.Create("./keys-cert/file_ca_pub.pem")
	if err != nil {
		return err
	}
	defer publicF.Close()

	err = pem.Encode(publicF, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: mPubKey,
	})
	if err != nil {
		return err
	}

	return nil
}

func createRekorKey(certConfig *ui.CertConfig) error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	mKey, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}

	file, err := os.Create("./keys-cert/rekor_key.pem")
	if err != nil {
		return err
	}
	defer file.Close()

	err = pem.Encode(file, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: mKey,
	})
	if err != nil {
		return err
	}

	return nil
}

func createFulcioCA(key *ecdsa.PrivateKey, certConfig *ui.CertConfig) error {
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * 10 * time.Hour)

	issuer := pkix.Name{
		CommonName:   certConfig.ClusterCommonName,
		Organization: []string{certConfig.OrganizationName},
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(0),
		Subject:               issuer,
		EmailAddresses:        []string{certConfig.OrganizationEmail},
		SignatureAlgorithm:    x509.ECDSAWithSHA256,
		BasicConstraintsValid: true,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		Issuer:                issuer,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
	}

	fulcioRoot, err := x509.CreateCertificate(rand.Reader, &template, &template, key.Public(), key)
	if err != nil {
		return err
	}

	f, err := os.Create("./keys-cert/fulcio-root.pem")
	if err != nil {
		return err
	}
	defer f.Close()
	err = pem.Encode(f, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: fulcioRoot,
	})

	if err != nil {
		return err
	}
	return nil
}
