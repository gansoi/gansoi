package ca

import (
	"crypto/ecdsa"
	"net"
	"testing"
)

func TestRandomString(t *testing.T) {
	for l := 0; l < 100; l++ {
		str := RandomString(l)

		if len(str) != l {
			t.Errorf("String is '%s' len()=%d, should be '%d'", str, len(str), l)
		}
	}
}

func TestGenerateKey(t *testing.T) {
	_, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() failed: %s", err.Error())
	}
}

func TestGenerateCSR(t *testing.T) {
	const hostname = "hostname.example.com"
	key, _ := GenerateKey()

	csr, err := GenerateCSR(key, hostname, []net.IP{net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatalf("GenerateKey() failed: %s", err.Error())
	}

	if csr.Subject.CommonName != hostname {
		t.Fatalf("GenerateKey() did not return a CSR with correct subject %s != %s", csr.Subject.CommonName, hostname)
	}
}

func TestGenerateCSRFail(t *testing.T) {
	key, _ := GenerateKey()

	_, err := GenerateCSR(key, "", nil)
	if err == nil {
		t.Fatalf("GenerateKey() did not catch empty common name")
	}

	cn := "test123"
	csr, err := GenerateCSR(nil, cn, nil)
	if err == nil {
		t.Fatalf("GenerateCSR() failed to detect nil key")
	}

	if csr != nil {
		t.Fatalf("GenerateCSR() returned non-nil when given nil key")
	}

	csr, err = GenerateCSR(&ecdsa.PrivateKey{}, cn, nil)
	if err == nil {
		t.Fatalf("GenerateCSR() failed to detect empty key")
	}

	if csr != nil {
		t.Fatalf("GenerateCSR() returned non-nil when given empty key")
	}
}
