package ca

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"net"
	"testing"
)

func TestDecodeKey(t *testing.T) {
	_, err := DecodeKey(testKey)
	if err != nil {
		t.Errorf("DecodeKey() failed: %s", err.Error())
	}
}

func TestDecodeKeyFail(t *testing.T) {
	unparsable := []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIMgdcx/E0yUexfJ7j5H1gikqMxU8LY9Nyz1NBtirp/5poaoGCCqGSM49
AwEHoUQDQgAEV0Dah48tLY7YK7TysUBTS8VuJ1e7QEXXlRgSLhOFVpOcnUIQidOS
mI9UKm05VL4kGD24SmoiL1hnwLoi0+1qsw==
-----END EC PRIVATE KEY-----
`)

	_, err := DecodeKey(unparsable)
	if err == nil {
		t.Errorf("DecodeKey() failet to detect broken key")
	}
}

func TestEncodeKey(t *testing.T) {
	key, _ := DecodeKey(testKey)
	encoded, err := EncodeKey(key)
	if err != nil {
		t.Fatalf("EncodeKey() failed: %s", err.Error())
	}

	if bytes.Compare(testKey, encoded) != 0 {
		t.Errorf("Unexpected output from EncodeKey()")
	}
}

func TestEncodeKeyFail(t *testing.T) {
	key, _ := DecodeKey(testKey)
	key.Curve = &unknownCurve{}
	_, err := EncodeKey(key)
	if err == nil {
		t.Fatalf("EncodeKey() did not fail")
	}
}

func TestRandomString(t *testing.T) {
	for l := 0; l < 100; l++ {
		str := RandomString(l)

		if len(str) != l {
			t.Errorf("String is '%s' len()=%d, should be '%d'", str, len(str), l)
		}
	}
}

func TestRandomStringFail(t *testing.T) {
	defer func() {
		recover()
	}()
	randSource = &failReader{failAt: 1}

	RandomString(20)

}

func TestGenerateKey(t *testing.T) {
	randSource = rand.Reader

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
