package ca

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

type (
	// CA represents a certificate authority.
	CA struct {
		privateKey  *ecdsa.PrivateKey
		Certificate *x509.Certificate
	}
)

// InitCA will start a new certificate authority.
func InitCA() (*CA, error) {
	ca := &CA{}

	// Generate private key.
	priv, err := GenerateKey()
	if err != nil {
		return nil, err
	}

	// Make it usable an hour earlier to account for clock skew.
	notBefore := time.Now().Add(-time.Hour)
	notAfter := notBefore.Add(20 * 365 * 24 * time.Hour) // 20 years

	template := x509.Certificate{
		SerialNumber:          randomSerial(),
		Subject:               pkix.Name{Organization: []string{"Gansoi Cluster CA"}},
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(randSource, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("CreateCertificate: %s", err.Error())
	}

	ca.privateKey = priv

	// We ignore errors here, we created the CSR, it's error-free (tm).
	ca.Certificate, _ = x509.ParseCertificate(derBytes)

	return ca, nil
}

// OpenCA will instantiate a new CA from an existing key/cert pair.
func OpenCA(keyPEM []byte, certPEM []byte) (*CA, error) {
	var err error
	ca := &CA{}

	ca.privateKey, err = DecodeKey(keyPEM)
	if err != nil {
		return nil, fmt.Errorf("key parse failed: %s", err.Error())
	}

	ca.Certificate, err = DecodeCert(certPEM)
	if err != nil {
		return nil, fmt.Errorf("parse cert: %s", err.Error())
	}

	// Check that we're self-signed.
	err = ca.Certificate.CheckSignatureFrom(ca.Certificate)
	if err != nil {
		return nil, errors.New("the certificate is not self-signed")
	}

	pub, ok := ca.Certificate.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not a ECDSA key")
	}

	if pub.X.Cmp(ca.privateKey.X) != 0 || pub.Y.Cmp(ca.privateKey.Y) != 0 {
		return nil, errors.New("tls: private key does not match certificate")
	}

	return ca, nil
}

// Fingerprint256 returns the certificate sha256 fingerprint as a string.
func (ca *CA) Fingerprint256() string {
	hash := sha256.Sum256(ca.Certificate.Raw)

	return fmt.Sprintf("%x", hash)
}

// CertificatePEM will return the root certificate in PEM format.
func (ca *CA) CertificatePEM() ([]byte, error) {
	if ca.Certificate == nil {
		return nil, fmt.Errorf("no certificate set")
	}

	return EncodeCert(ca.Certificate)
}

// KeyPEM will return the private CA key PEM encoded.
func (ca *CA) KeyPEM() ([]byte, error) {
	if ca.privateKey == nil {
		return nil, fmt.Errorf("no private key set")
	}

	return EncodeKey(ca.privateKey)
}

// SignCSR will sign a CSR and generate a new certificate.
func (ca *CA) SignCSR(csr *x509.CertificateRequest) (*x509.Certificate, error) {
	template := &x509.Certificate{
		SerialNumber: randomSerial(),
		Subject:      csr.Subject,
		PublicKey:    csr.PublicKey,
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(20 * 365 * 24 * time.Hour),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		IPAddresses:  csr.IPAddresses,
	}

	err := csr.CheckSignature()
	if err != nil {
		return nil, err
	}

	derBytes, err := x509.CreateCertificate(randSource, template, ca.Certificate, csr.PublicKey, ca.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %s", err.Error())
	}

	// Ignore errors from parsing a valid cert created by us.
	cert, _ := x509.ParseCertificate(derBytes)

	return cert, nil
}

// Verify that cert is signed by our root.
func (ca *CA) Verify(cert *x509.Certificate) (bool, error) {
	if cert == nil {
		return false, errors.New("no certificate")
	}

	opts := x509.VerifyOptions{}
	opts.Roots = x509.NewCertPool()
	opts.Roots.AddCert(ca.Certificate)

	_, err := cert.Verify(opts)

	if err != nil {
		return false, err
	}

	return true, nil
}

// randomSerial generates a random serial for certificates.
func randomSerial() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)

	return serialNumber
}

// CertPool returns a pool suitable for use in http clients.
func (ca *CA) CertPool() *x509.CertPool {
	pool := x509.NewCertPool()

	if ca.Certificate != nil {
		pool.AddCert(ca.Certificate)
	}

	return pool
}

// VerifyHTTPRequest verifies that a HTTP remote has presented a certificate
// signed by this CA.
func (ca *CA) VerifyHTTPRequest(req *http.Request) (string, error) {
	if ca.Certificate == nil {
		return "", errors.New("certificate authority not ready")
	}

	if len(req.TLS.PeerCertificates) < 1 {
		return "", errors.New("no certificate provided")
	}

	clientCert := req.TLS.PeerCertificates[0]
	err := clientCert.CheckSignatureFrom(ca.Certificate)
	if err != nil {
		return "", err
	}

	return clientCert.Subject.CommonName, nil
}
