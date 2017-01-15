package ca

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
)

// DecodeCert will decode a PEM encoded certificate.
func DecodeCert(certPEM []byte) (*x509.Certificate, error) {
	b, _ := pem.Decode(certPEM)
	if b == nil {
		return nil, errors.New("Certificate PEM decode failed")
	}

	// Read certificate.
	cert, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		return nil, fmt.Errorf("ParseCertificate: %s", err.Error())
	}

	return cert, nil
}

// DecodeKey will decode a PEM encoded private key.
func DecodeKey(keyPEM []byte) (*ecdsa.PrivateKey, error) {
	// Decode key.
	b, _ := pem.Decode(keyPEM)
	if b == nil || len(b.Bytes) < 1 {
		return nil, errors.New("Key PEM decode failed")
	}

	if b.Type != "EC PRIVATE KEY" {
		return nil, errors.New("Wrong key type")
	}

	// Read key.
	key, err := x509.ParseECPrivateKey(b.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Key parse failed: %s", err.Error())
	}

	return key, nil
}

// EncodeCert is a small helper function that will PEM encode a certificate.
func EncodeCert(cert *x509.Certificate) ([]byte, error) {
	b := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})

	return b, nil
}

// EncodeKey is a helper that will PEM encode a private key.
func EncodeKey(privateKey *ecdsa.PrivateKey) ([]byte, error) {
	b, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	b = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})

	return b, nil
}

// RandomString will generate a pseudo-random string consisting of length
// alphanumeric runes.
func RandomString(length int) string {
	if length == 0 {
		return ""
	}

	var vocabulary = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // storage for random bytes.
	vlen := byte(len(vocabulary))
	maxrb := byte(256 - (256 % len(vocabulary)))
	i := 0
	for {
		if _, err := io.ReadFull(rand.Reader, r); err != nil {
			panic("error reading from random source: " + err.Error())
		}
		for _, c := range r {
			if c >= maxrb {
				// Skip to avoid modulo bias.
				continue
			}
			b[i] = vocabulary[c%vlen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}

// GenerateKey will generate a new ECDSA private key.
func GenerateKey() (*ecdsa.PrivateKey, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

// GenerateCSR will generate a new certificate signing request.
func GenerateCSR(key *ecdsa.PrivateKey, commonName string, ips []net.IP) (*x509.CertificateRequest, error) {
	if key == nil {
		return nil, errors.New("key is nil")
	}

	if commonName == "" {
		return nil, errors.New("Empty commonName")
	}

	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{""},
		},
		SignatureAlgorithm: x509.ECDSAWithSHA256,
		DNSNames:           []string{commonName},
	}

	for _, ip := range ips {
		template.IPAddresses = append(template.IPAddresses, ip)
	}

	derBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, key)
	if err != nil {
		return nil, err
	}

	csr, err := x509.ParseCertificateRequest(derBytes)
	if err != nil {
		return nil, err
	}

	return csr, nil
}
