package ca

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"testing"
)

var ()

var (
	testKey = []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIMgdcx/E0yUexfJ7j5H1gikqMxU8LY9Nyz1NBtirp/5poAoGCCqGSM49
AwEHoUQDQgAEV0Dah48tLY7YK7TysUBTS8VuJ1e7QEXXlRgSLhOFVpOcnUIQidOS
mI9UKm05VL4kGD24SmoiL1hnwLoi0+1qsw==
-----END EC PRIVATE KEY-----
`)

	testCa = []byte(`-----BEGIN CERTIFICATE-----
MIIBeDCCAR+gAwIBAgIRAMJJH5QMGmWgisxpH53YigwwCgYIKoZIzj0EAwIwITEf
MB0GA1UEChMWRklYTUU6IElzIHRoaXMgbmVlZGVkPzAeFw0xNzAxMDExNjM0MjZa
Fw0zNjEyMjcxNjM0MjZaMCExHzAdBgNVBAoTFkZJWE1FOiBJcyB0aGlzIG5lZWRl
ZD8wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARXQNqHjy0tjtgrtPKxQFNLxW4n
V7tARdeVGBIuE4VWk5ydQhCJ05KYj1QqbTlUviQYPbhKaiIvWGfAuiLT7Wqzozgw
NjAOBgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/
BAUwAwEB/zAKBggqhkjOPQQDAgNHADBEAiAwPmNW8yWylX/4vGH3c1edMvL5ZRus
9J5j/lQtE3vnbwIgRTygXaPQMVoinUNfOjJiORuVjONDWAIB1kpSaLMRq34=
-----END CERTIFICATE-----
`)

	testFingerprint = `9693998d872a743f2a70ba899ebc0e48b7739b5b9d81e950c681580d4ba6bee0`
)

func openTestCA() (*CA, error) {
	return OpenCA([]byte(testKey), []byte(testCa))
}

func TestInitCA(t *testing.T) {
	ca, err := InitCA()
	if err != nil {
		t.Fatalf("InitCA() failed: %s", err.Error())
	}

	if ca == nil {
		t.Fatalf("OpenCA() returned nil")
	}

	b, _ := ca.CertificatePEM()
	ioutil.WriteFile("/tmp/cert2.pem", []byte(b), 0600)
}

func TestOpenCA(t *testing.T) {
	ca, err := openTestCA()

	if err != nil {
		t.Fatalf("OpenCA() failed: %s", err.Error())
	}

	if ca == nil {
		t.Fatalf("OpenCA() returned nil")
	}
}

func TestOpenCAFail(t *testing.T) {
	k := []byte(testKey)
	c := []byte(testCa)
	e := []byte("") // empty
	brokenCert := []byte(testCa[0:100])
	brokenCert2 := []byte(`-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
    `)
	brokenCert3 := []byte(`-----BEGIN CERTIFICATE-----
    MIIBeDCCAR+gAwIBAgIRAP+UTvdh9jddMH7KSwVsBD4wCgYIKoZIzj0EAwIwITEf
    MB0GA1UEChMWRklYTUU6IElzIHRoaXMgbmVlZGVkPzAeFw0xNzAxMDExNzE5NTZa
    Fw0zNjEyMjcxNzE5NTZaMCExHzAdBgNVBAoTFkZJWE1FOiBJcyB0aGlzIG5lZWRl
    ZD8wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARMy0IZ4iWiUThz07Ic4cnyny1K
    NBzJHBYB+vGqM2XdL7OfvNgIxDsStO2YdjIH/xhcXyk0HpkLr7Isq3F7HdB4ozgw
-----END CERTIFICATE-----
`)
	mismatch := []byte(`-----BEGIN CERTIFICATE-----
MIIBeDCCAR+gAwIBAgIRAP+UTvdh9jddMH7KSwVsBD4wCgYIKoZIzj0EAwIwITEf
MB0GA1UEChMWRklYTUU6IElzIHRoaXMgbmVlZGVkPzAeFw0xNzAxMDExNzE5NTZa
Fw0zNjEyMjcxNzE5NTZaMCExHzAdBgNVBAoTFkZJWE1FOiBJcyB0aGlzIG5lZWRl
ZD8wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARMy0IZ4iWiUThz07Ic4cnyny1K
NBzJHBYB+vGqM2XdL7OfvNgIxDsStO2YdjIH/xhcXyk0HpkLr7Isq3F7HdB4ozgw
NjAOBgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/
BAUwAwEB/zAKBggqhkjOPQQDAgNHADBEAiAOruOFFHNoa4jbNyLTnda3JVsHsu3J
nXv93Y9HMQzmjAIgfRmMBVxRpxp47SJmNN5dBjzpxve1MJo5AQDE8DQeVfc=
-----END CERTIFICATE-----
`)
	brokenKey := []byte(`-----BEGIN EC PRIVATE KEY-----
MIIBeDCCAR+gAwIBAgIRAP+UTvdh9jddMH7KSwVsBD4wCgYIKoZIzj0EAwIwITEf
MB0GA1UEChMWRklYTUU6IElzIHRoaXMgbmVlZGVkPzAeFw0xNzAxMDExNzE5NTZa
Fw0zNjEyMjcxNzE5NTZaMCExHzAdBgNVBAoTFkZJWE1FOiBJcyB0aGlzIG5lZWRl
ZD8wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARMy0IZ4iWiUThz07Ic4cnyny1K
NBzJHBYB+vGqM2XdL7OfvNgIxDsStO2YdjIH/xhcXyk0HpkLr7Isq3F7HdB4ozgw
NjAOBgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/
BAUwAwEB/zAKBggqhkjOPQQDAgNHADBEAiAOruOFFHNoa4jbNyLTnda3JVsHsu3J
nXv93Y9HMQzmjAIgfRmMBVxRpxp47SJmNN5dBjzpxve1MJo5AQDE8DQeVfc=
-----END CERTIFICATE-----
`)

	nonSigned := []byte(`-----BEGIN CERTIFICATE-----
MIIFczCCBFugAwIBAgIQAYdAdtYc/yHec/7saKg3XTANBgkqhkiG9w0BAQsFADBC
MQswCQYDVQQGEwJVUzEWMBQGA1UEChMNR2VvVHJ1c3QgSW5jLjEbMBkGA1UEAxMS
UmFwaWRTU0wgU0hBMjU2IENBMB4XDTE2MTAwNTAwMDAwMFoXDTE3MTAwNTIzNTk1
OVowGzEZMBcGA1UEAwwQKi5nYW5zb2ktZGV2LmNvbTCCASIwDQYJKoZIhvcNAQEB
BQADggEPADCCAQoCggEBALTDvIJh3UidNxnqV9St2dJDmaNSEtHncq8wtir23xX7
qU4cbkvwc8IQuMbsUAdUGyRPqKHQpAzPLPIuWMrMb4fyxjBAIezm5WeQY9uFhf5U
iXN3TTIlQg6TA0OTbblIV7tajYZ8s2mUIKdvix7YRZKWEbQwffvH1g5riy1UAi97
gZC0JzjQheh8UL25n1eatw6cAt7UMBo18T4kn6ZlLoPxscfzxKswt9ugyXy5U5Q3
Bz4wyknzQ6dHLJ/dNqvb0jtZOCA2MtbVLFucrAwCLggfVhj47o655BspEXq3eRyO
VbWYne4V9wXlAR+me2e6R4aWSkj7S/6kBSG+7vMMDWUCAwEAAaOCAoowggKGMCsG
A1UdEQQkMCKCECouZ2Fuc29pLWRldi5jb22CDmdhbnNvaS1kZXYuY29tMAkGA1Ud
EwQCMAAwKwYDVR0fBCQwIjAgoB6gHIYaaHR0cDovL2dwLnN5bWNiLmNvbS9ncC5j
cmwwbwYDVR0gBGgwZjBkBgZngQwBAgEwWjAqBggrBgEFBQcCARYeaHR0cHM6Ly93
d3cucmFwaWRzc2wuY29tL2xlZ2FsMCwGCCsGAQUFBwICMCAMHmh0dHBzOi8vd3d3
LnJhcGlkc3NsLmNvbS9sZWdhbDAfBgNVHSMEGDAWgBSXwidQnsLJ7AyIMsh8reKm
AU/abzAOBgNVHQ8BAf8EBAMCBaAwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUF
BwMCMFcGCCsGAQUFBwEBBEswSTAfBggrBgEFBQcwAYYTaHR0cDovL2dwLnN5bWNk
LmNvbTAmBggrBgEFBQcwAoYaaHR0cDovL2dwLnN5bWNiLmNvbS9ncC5jcnQwggED
BgorBgEEAdZ5AgQCBIH0BIHxAO8AdgDd6x0reg1PpiCLga2BaHB+Lo6dAdVciI09
EcTNtuy+zAAAAVeV5eNpAAAEAwBHMEUCIDfjN2FPpp9IQmH4GHoXFt6m2Zpnh53P
KFqDNdnKV9RgAiEAyzR+PRfIVEAthACr7JNrnMzcRO0xerv7XfmnIZfTIq0AdQBo
9pj4H2SCvjqM7rkoHUz8cVFdZ5PURNEKZ6y7T0/7xAAAAVeV5eOAAAAEAwBGMEQC
IAmQ6r1H4O0Di0C4xQ2gBhoenbsVejtd3EnL5ddYZav8AiB5tK6cXfDB3K4e8m3u
ArLae2XACKMNsYjOlbaAsmp1pzANBgkqhkiG9w0BAQsFAAOCAQEAbHnLGGbucDZV
B697kCFkeZAVUJWAVKACEs+T3Tya5DWUbJIBRmAXprUp64kcT92KC/N2wSXfE4qj
q1G4hEK3TjWaAeu26V/+tPLGNXa6HwpTIuVnENIrEQ7vnQYp6Q9zOieGobhGUP47
1vPsCKYEzggi7nN1jQOJEKpBm14332vu2tsS1EgsPbGon/+xdkojeWJ3nbrgcrK/
mMYOrNvq1ikQ+PcOdAnULwJklMExdjkRT/p+TizF9Nw3/7i2X6T9TYZ1lQYyfGC4
Bqa7BzipeYqlnCR75GGnkbTP3OPWkGpLMTTokyV9libM+LaGshgL1up0t5QQ/kFF
cOhykCUsFg==
-----END CERTIFICATE-----
`)

	nonECDSA := []byte(`-----BEGIN CERTIFICATE-----
MIIDBzCCAe+gAwIBAgIJALUXBHwdePpOMA0GCSqGSIb3DQEBCwUAMBoxGDAWBgNV
BAMMD3d3dy5leGFtcGxlLmNvbTAeFw0xNzAxMDExODMzMjJaFw0yNjEyMzAxODMz
MjJaMBoxGDAWBgNVBAMMD3d3dy5leGFtcGxlLmNvbTCCASIwDQYJKoZIhvcNAQEB
BQADggEPADCCAQoCggEBALgZwNblbXUMXu4xbZXPk0RtFbUyBj2hUizunxeqP4pT
R9oMheiTGXUKR+f/WsqlPFn1Gn7+gier4lRybwYvgzOE/kSSNsuemGCoP9iIc04m
NrxMInzgrve8CYvHhd32rZH1Mson3LcIsnRtAqBdpFGbfAbWqvDOd0LKcBCDq1/O
/CLhyDstrgJ7sbN2W3sUSWrb2KGSote/vcwpDHCXW51uSULvHtS2ZC2HU/N7ki6B
EFllTHVc1/uDo9lZaChBzIIRWvoolypZMVNSBMkJxYNnH22NY4MR8JlP1p90A6PD
+8uL3cUpgV8v1x6oRDTjqmTWBiSly+A2bi5Z+Tu3WJMCAwEAAaNQME4wHQYDVR0O
BBYEFLjhsIcjOIzZy+XjlGYmvIzgN8oFMB8GA1UdIwQYMBaAFLjhsIcjOIzZy+Xj
lGYmvIzgN8oFMAwGA1UdEwQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBABurHa7D
H9rTClaKDQXCJAtl3NK0aNO5FWEMyP56Vfem1v7ZANGTqCkDzRH+eZ73Fys0YV/9
pTfAuYo9BHOBS0dXtDxbZY2bzgI2zX1SFN+VXUPcM+cUps0W2kpfPzOaWVLFbZE4
ZkaVFNHbF1K0mVORLQnNmLCSTwvAlB+8NYN0bxLeJXXHEMkSEC/laySfCVahfkB5
+8ZeX7PTwCPD7R/+DLhouIXQdqnPtCHfDlgSZN5iq5WyhmT5+qKPJFChSNIG7wHs
XLb+GEMinuabyI+hWdluTXbyWU0T1bxi2bdIpHJjC7dZTroGAQ+2F+4v8odaIx/T
N71u95VZi7yvJoc=
-----END CERTIFICATE-----
`)

	cases := []struct {
		key  []byte
		cert []byte
	}{
		{nil, nil},
		{k, nil},
		{k, e},
		{k, k},
		{nil, c},
		{e, c},
		{c, c},
		{e, e},
		{k, brokenCert},
		{k, brokenCert2},
		{k, brokenCert3},
		{k, mismatch},
		{k, nonSigned},
		{k, nonECDSA},
		{brokenKey, c},
	}

	for i, input := range cases {
		ca, err := OpenCA(input.key, input.cert)
		if err == nil {
			t.Fatalf("OpenCA() failed to detect wrong input (case %d)", i)
		}

		if ca != nil {
			t.Fatalf("OpenCA() failed to return nil")
		}
	}
}

func TestFingerprint256(t *testing.T) {
	ca, _ := openTestCA()
	fingerprint := ca.Fingerprint256()
	if len(fingerprint) != 64 {
		t.Fatalf("Fingerprint256() returned wrong length: %d", len(fingerprint))
	}

	if fingerprint != testFingerprint {
		t.Fatalf("Fingerprint256() returned wrong string, got '%s', expected '%s'", fingerprint, testFingerprint)
	}
}

func TestCertificatePEM(t *testing.T) {
	ca, _ := openTestCA()

	certPem, err := ca.CertificatePEM()
	if err != nil {
		t.Fatalf("CertificatePEM() failed: %s", err.Error())
	}

	if bytes.Compare(certPem, testCa) != 0 {
		t.Fatalf("CertificatePEM() returned wrong output, got '%s', expected '%s'", certPem, testCa)
	}
}

func TestCertificatePEMFail(t *testing.T) {
	ca := &CA{}

	certPem, err := ca.CertificatePEM()
	if err == nil {
		t.Fatalf("CertificatePEM() failed to detect empty cert")
	}

	if certPem != nil {
		t.Fatalf("CertificatePEM() returned wrong output, got '%s', expected ''", certPem)
	}
}

func TestKeyPEM(t *testing.T) {
	ca, _ := openTestCA()

	keyPem, err := ca.KeyPEM()
	if err != nil {
		t.Fatalf("KeyPEM() failed: %s", err.Error())
	}

	if bytes.Compare(keyPem, testKey) != 0 {
		t.Fatalf("KeyPEM() returned wrong output, got '%s', expected '%s'", keyPem, testKey)
	}
}

func TestKeyPEMFail(t *testing.T) {
	ca := &CA{}

	keyPem, err := ca.KeyPEM()
	if err == nil {
		t.Fatalf("KeyPEM() failed to detect empty cert")
	}

	if keyPem != nil {
		t.Fatalf("KeyPEM() returned wrong output, got '%s', expected ''", keyPem)
	}
}

func TestSignCSR(t *testing.T) {
	cn := "test123"
	key, _ := GenerateKey()
	csr, err := GenerateCSR(key, cn, nil)
	if err != nil {
		t.Fatalf("GenerateCSR() failed: %s", err.Error())
	}

	if csr.Subject.CommonName != cn {
		t.Fatalf("CommonName is wrong. got '%s', expected '%s'", csr.Subject.CommonName, cn)
	}

	ca, _ := openTestCA()
	cert, err := ca.SignCSR(csr)
	if err != nil {
		t.Fatalf("SignCSR() failed: %s", err.Error())
	}

	signed, _ := ca.Verify(cert)
	if !signed {
		t.Fatalf("The certificate is not signed")
	}
}

func TestSignCSRFail(t *testing.T) {
	key, _ := GenerateKey()
	csr, _ := GenerateCSR(key, "hello", nil)

	ca, _ := openTestCA()
	csr.Signature = nil
	cert, err := ca.SignCSR(csr)
	if err == nil {
		t.Fatalf("SignCSR() failed to catch broken signature")
	}

	if cert != nil {
		t.Fatalf("SignCSR() returned a cert from broken CSR")
	}
}

func TestVerify(t *testing.T) {
	cn := "test123"
	key, _ := GenerateKey()
	csr, err := GenerateCSR(key, cn, nil)

	ca, _ := openTestCA()
	cert, err := ca.SignCSR(csr)
	if err != nil {
		t.Fatalf("SignCSR() failed: %s", err.Error())
	}

	verified, err := ca.Verify(cert)
	if err != nil {
		t.Fatalf("Verify() returned an error: %s", err.Error())
	}

	if !verified {
		t.Fatalf("Verify() failed to verify the signed cert")
	}
}

func TestVerifyFail(t *testing.T) {
	ca, _ := openTestCA()
	key, _ := GenerateKey()
	csr, _ := GenerateCSR(key, "hostname", nil)
	cert, _ := ca.SignCSR(csr)

	// Start a new CA with a new keypair.
	ca, _ = InitCA()

	verified, err := ca.Verify(cert)
	if err == nil {
		t.Fatalf("Verify() did not generate error for cert signed by another CA")
	}

	if verified {
		t.Fatalf("Verify() verified cert from another CA")
	}

	// Check nil argument.
	_, err = ca.Verify(nil)
	if err == nil {
		t.Fatalf("Verify() failed to detect nil cert")
	}
}

func TestCertPool(t *testing.T) {
	ca, _ := openTestCA()
	pool := ca.CertPool()

	if len(pool.Subjects()) != 1 {
		t.Fatalf("CertPol() returned the wrong number of certificates")
	}
}

func TestCertPoolFail(t *testing.T) {
	ca := &CA{}
	pool := ca.CertPool()

	if len(pool.Subjects()) != 0 {
		t.Fatalf("CertPol() returned the wrong number of certificates")
	}
}

func TestVerifyHTTPRequest(t *testing.T) {
	ca := &CA{}
	req := &http.Request{}
	req.TLS = &tls.ConnectionState{}

	// Test blank CA
	_, err := ca.VerifyHTTPRequest(req)
	if err == nil {
		t.Fatalf("VerifyHTTPRequest() failed to error on 'blank' CA")
	}

	// Add certificate
	cn := "test123"
	key, _ := GenerateKey()
	csr, err := GenerateCSR(key, cn, nil)
	if err != nil {
		t.Fatalf("GenerateCSR() failed: %s", err.Error())
	}

	if csr.Subject.CommonName != cn {
		t.Fatalf("CommonName is wrong. got '%s', expected '%s'", csr.Subject.CommonName, cn)
	}

	// Test request without cert
	ca, _ = openTestCA()
	_, err = ca.VerifyHTTPRequest(req)
	if err == nil {
		t.Fatalf("VerifyHTTPRequest() failed to error on missing cert")
	}

	cert, err := ca.SignCSR(csr)
	if err != nil {
		t.Fatalf("SignCSR() failed: %s", err.Error())
	}

	req.TLS.PeerCertificates = []*x509.Certificate{cert}
	_, err = ca.VerifyHTTPRequest(req)
	if err != nil {
		t.Fatalf("VerifyHTTPRequest() errored on signed request")
	}

	ca, _ = InitCA()
	_, err = ca.VerifyHTTPRequest(req)
	if err == nil {
		t.Fatalf("VerifyHTTPRequest() failed to catch unsigned cert")
	}
}
