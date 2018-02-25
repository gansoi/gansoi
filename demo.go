package main

import (
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"github.com/gansoi/gansoi/logger"
	"github.com/spf13/cobra"
)

var (
	demoCert = []byte(`-----BEGIN CERTIFICATE-----
MIIFdzCCBF+gAwIBAgIQQhQJcGJQyihkOE0ronzRbzANBgkqhkiG9w0BAQsFADBH
MQswCQYDVQQGEwJVUzEWMBQGA1UEChMNR2VvVHJ1c3QgSW5jLjEgMB4GA1UEAxMX
UmFwaWRTU0wgU0hBMjU2IENBIC0gRzIwHhcNMTcxMDA2MDAwMDAwWhcNMTgxMTA1
MjM1OTU5WjAbMRkwFwYDVQQDDBAqLmdhbnNvaS1kZXYuY29tMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtMO8gmHdSJ03GepX1K3Z0kOZo1IS0edyrzC2
KvbfFfupThxuS/BzwhC4xuxQB1QbJE+oodCkDM8s8i5Yysxvh/LGMEAh7OblZ5Bj
24WF/lSJc3dNMiVCDpMDQ5NtuUhXu1qNhnyzaZQgp2+LHthFkpYRtDB9+8fWDmuL
LVQCL3uBkLQnONCF6HxQvbmfV5q3DpwC3tQwGjXxPiSfpmUug/Gxx/PEqzC326DJ
fLlTlDcHPjDKSfNDp0csn902q9vSO1k4IDYy1tUsW5ysDAIuCB9WGPjujrnkGykR
erd5HI5VtZid7hX3BeUBH6Z7Z7pHhpZKSPtL/qQFIb7u8wwNZQIDAQABo4ICiTCC
AoUwKwYDVR0RBCQwIoIQKi5nYW5zb2ktZGV2LmNvbYIOZ2Fuc29pLWRldi5jb20w
CQYDVR0TBAIwADArBgNVHR8EJDAiMCCgHqAchhpodHRwOi8vZ3Muc3ltY2IuY29t
L2dzLmNybDBvBgNVHSAEaDBmMGQGBmeBDAECATBaMCoGCCsGAQUFBwIBFh5odHRw
czovL3d3dy5yYXBpZHNzbC5jb20vbGVnYWwwLAYIKwYBBQUHAgIwIAweaHR0cHM6
Ly93d3cucmFwaWRzc2wuY29tL2xlZ2FsMB8GA1UdIwQYMBaAFEz0v+g7vsIk8xtH
O7VuSI4Wq68SMA4GA1UdDwEB/wQEAwIFoDAdBgNVHSUEFjAUBggrBgEFBQcDAQYI
KwYBBQUHAwIwVwYIKwYBBQUHAQEESzBJMB8GCCsGAQUFBzABhhNodHRwOi8vZ3Mu
c3ltY2QuY29tMCYGCCsGAQUFBzAChhpodHRwOi8vZ3Muc3ltY2IuY29tL2dzLmNy
dDCCAQIGCisGAQQB1nkCBAIEgfMEgfAA7gB1AN3rHSt6DU+mIIuBrYFocH4ujp0B
1VyIjT0RxM227L7MAAABXvAj61IAAAQDAEYwRAIgNXmBiqdkNyUo/VA+Nb5erQ2G
FfnkGC8Zw1abu6EPGsECIFNtaOJvYrn3nKBI1r9ebahFeRJ7oqrCHCM/XO00k5mq
AHUApLkJkLQYWBSHuxOizGdwCjw1mAT5G9+443fNDsgN3BAAAAFe8CPrkwAABAMA
RjBEAiAWJk465qxBD/f17vTCqws6sA7FhKsa7Bmg+1MDezKZUQIgVnY1mmppis8X
aNmpbvtqmNAch0INNTJDbRU8T5WgYp4wDQYJKoZIhvcNAQELBQADggEBAJPBPIcr
m89+XazLmVotvUby71AWqO18NrCV1ytYCgyAsRJYkF/NukEQL7WDxyT3cCLV41+i
GLNp83X9t1+YUPMXSABHyAqgAmctiw5CBFEpTl/iay3T0gmB/aisPlLlUR0FaVvR
W+8CjNIDBCVnt4I5XOV8izEwKYqta8Hirhl2eeYTHNsmknIUu0iUqfHlNcKbwzXA
uH82p04qcxtw5UhbhGJw2yIpiiu5Ndl7RulZjliJgWW7GFokPB4+uQfsiyEYfpqL
fd9ca+WLwnOOmO3UqRWZydaDH4ZLBstaQu6rMZUIkF00IhJxY7Mas4XHca6D1Jns
h8PMwU9r0aStaEY=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIEtTCCA52gAwIBAgIQSOmUQNQ2SRy4uII9CUOUxzANBgkqhkiG9w0BAQsFADCBmDELMAkGA1UE
BhMCVVMxFjAUBgNVBAoTDUdlb1RydXN0IEluYy4xOTA3BgNVBAsTMChjKSAyMDA4IEdlb1RydXN0
IEluYy4gLSBGb3IgYXV0aG9yaXplZCB1c2Ugb25seTE2MDQGA1UEAxMtR2VvVHJ1c3QgUHJpbWFy
eSBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eSAtIEczMB4XDTE0MDYxMDAwMDAwMFoXDTI0MDYwOTIz
NTk1OVowRzELMAkGA1UEBhMCVVMxFjAUBgNVBAoTDUdlb1RydXN0IEluYy4xIDAeBgNVBAMTF1Jh
cGlkU1NMIFNIQTI1NiBDQSAtIEcyMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxJVj
KNBOMEWvi5c0FEX4XFhK+jOObpxgq/OG/zR0siu+oYzVoqNgekC54fwiyme6YKrHmvkGf+73uoUF
sAP/cq4VQUqYZNcXS1TvBcaYB5MnPk/cD8Z7i+fzBl6N6LSuKbQeHi0WkNPqqueMO22vNln/xQr6
x0y9NotkxEr1zjP5B75/RZCoCBSw0KVP34KA2hvuwxOwmPUP+X52tea5XWi5XFCQiaQ2sXAW6rEQ
tWp23+G7/HjycpnPyaLUc1R3v8A5d+WuEsV4WhlF1EEZ03z1b5lr14u8LQmdSxBhwNpSw68iQ8br
N35jdDANanGO3l1bisjF15sp6K62JWGB6wIDAQABo4IBSTCCAUUwLgYIKwYBBQUHAQEEIjAgMB4G
CCsGAQUFBzABhhJodHRwOi8vZy5zeW1jZC5jb20wEgYDVR0TAQH/BAgwBgEB/wIBADBMBgNVHSAE
RTBDMEEGCmCGSAGG+EUBBzYwMzAxBggrBgEFBQcCARYlaHR0cDovL3d3dy5nZW90cnVzdC5jb20v
cmVzb3VyY2VzL2NwczA2BgNVHR8ELzAtMCugKaAnhiVodHRwOi8vZy5zeW1jYi5jb20vR2VvVHJ1
c3RQQ0EtRzMuY3JsMA4GA1UdDwEB/wQEAwIBBjApBgNVHREEIjAgpB4wHDEaMBgGA1UEAxMRU3lt
YW50ZWNQS0ktMS02OTcwHQYDVR0OBBYEFEz0v+g7vsIk8xtHO7VuSI4Wq68SMB8GA1UdIwQYMBaA
FMR5yo6hTgMdHNxr2zFblD4/MH8tMA0GCSqGSIb3DQEBCwUAA4IBAQB6U7Xetu9So1+K9YnxQsxe
RoiupQiHUd4PDwLrDIJ443N9cb1D6cqKP+AlkpszM3RJXgDZcxQcC0Z2HIoNTYxsfkv3YNiBeKB4
0CViqxDKIugcGd1Sg2QF5Ydmrud6pDs+2HB6dqJnOdTJ+uW3HkHiCTmIHBhVCsRBr7Lz8w9CFGF0
gePah1qaTYvTyY+JZhMpEeT/4t+Olgxaoaprm/38AztVDaaiJUgXH0Ko2mx+aW6g32fSbfQOahJ5
9XzIpTIcxDGy5ruoa2qiimBpwFd9svIxDJhlMuwIWs7GmOkhlz8seSkD9faUK1Mx85NoV+HXTzrR
YaFgzrmrmK41VGOL
-----END CERTIFICATE-----
`)

	demoKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAtMO8gmHdSJ03GepX1K3Z0kOZo1IS0edyrzC2KvbfFfupThxu
S/BzwhC4xuxQB1QbJE+oodCkDM8s8i5Yysxvh/LGMEAh7OblZ5Bj24WF/lSJc3dN
MiVCDpMDQ5NtuUhXu1qNhnyzaZQgp2+LHthFkpYRtDB9+8fWDmuLLVQCL3uBkLQn
ONCF6HxQvbmfV5q3DpwC3tQwGjXxPiSfpmUug/Gxx/PEqzC326DJfLlTlDcHPjDK
SfNDp0csn902q9vSO1k4IDYy1tUsW5ysDAIuCB9WGPjujrnkGykRerd5HI5VtZid
7hX3BeUBH6Z7Z7pHhpZKSPtL/qQFIb7u8wwNZQIDAQABAoIBAFc3ngv2tjMkEAME
C+7FFzUZgtbHcecvWilnQm4GgWr06yKSGzGuyduX/9+TA6YVkab6fG3e4lh/2brc
W+E1tJaOr8t2Fihc29EVXOj9SsKE/XDl3ixUx8OKcWe4iZd9bT8rmN+L6XEGlJ7U
9fYi+aaZm98qCo9iQ0jf4N41C3zmtN2iXq8qoXqa9ad/RMAa8I17Yil6m/YRAZac
3VxSH7GVYyia8c983I9Pdrkk1iLJbpNDXAbs6eA6ioEkPB4B0GJywVcpdOLQwxOW
7aE+XnZLooQkMrf4yZbSfMAGyp3NidXel93U4jrkmc26nEZKBHQo6DsY/J5ZVSl0
iB7rcP0CgYEA7gKEr2L0XJtIy33/S9PIFA1D6wo0We589ieeav7H/EqPClG3p1Fq
RCsLitXgAjXNCBkLn09Yi5lP5nVO2HkkJfBFp1rwHKdIFdLC2LivzhF2qkLnFOBt
h6KnbRLc90SfzlLwb0+gffVC26cldKNa4iz89Kd1gy3rR2OhXIhDUT8CgYEAwm2G
U4dy5LYYmX44eWzIz6R+L+aXSrjHB6wNNG0yxH6xjwlgcAGUzrwWSOeBlqwx2Uh5
5abL1pt2YbPyBrZGsyogArZ30b2tYBMi/Hm4d2M7ZP2tV02nUnFuE/JWXnnQ2i2/
M2qZQZgBjP15v4pQjuedM3+n3tNHxF3l/TCq1FsCgYBU0Hzr6v1dStDEAyBIqy1v
R9LeHQLO0VeieDfRtP0bAI68hKZHb5HIvPYeAV0ULIvlyNcFbEcHaBi67S6toW2q
P1by7ksGSu47KKHajOXJLxv0TGcAX4FohiPXkJNBYij4Y0HeyKdOe2nZ0FRenh+y
3Yk+vbX4ixJ+nBhSWxRyDQKBgHpaJkAGav0WwuBFGBEBrlVNNMO8HtU89rMTSLQH
S/1vpRlYU0HpHNYEcxmp5lkFP9F21I76qigBaTwO223x4wf4qHBMl5Z8ANEG+etc
RgOLhOMG6MCZ84PkMduHk2aczhue0kXu/UbT+5XYJXet+QgVJU41NVT6LJ2cYZE1
wSslAoGBAOk5UUTHm+HVy17ztc38moMKsqDmEFvLcLFoiGqdZhHgouw+gUQg21Uy
o+7hG62xVvzquorK1CB0MKx92GIyGtk0319Ta0bt6oWwe/MoLtslgcYbjk8UdN5P
/eElJPzec9D3uSLCUihW5O8AjcrYOsTZftRnIQTxuaF2eobj0sLR
-----END RSA PRIVATE KEY-----
`)

	demoConfig = `
bind: "127.0.0.1:0"
datadir: "{{ .Datadir }}"

http:
bind: "gansoi-dev.com:9002"
hostnames:
- "gansoi-dev.com"
cert: "{{ .Datadir }}/cert.pem"
key: "{{ .Datadir }}/key.pem"

checks:
gansoi.com-https:
agent: "http"
arguments:
  url: "https://gansoi.com/"
expressions:
  - "StatusCode == 200"
  - "TimeAccumulated < 1500"

gansoi-com-http:
agent: "http"
arguments:
  url: "http://gansoi.com/"
  followRedirect: false
expressions:
  - "StatusCode == 302"
  - "TimeAccumulated < 1500"

gansoi-com-ssh:
agent: "ssh"
arguments:
  address: "gansoi.com"
expressions:
  - "HandshakeTime < 1000"
`
)

func runDemo(cmd *cobra.Command, arguments []string) {
	cwd, _ := os.Getwd()

	dataDir := path.Join(cwd, "demo-data")
	configFile = path.Join(dataDir, "config.yml")

	os.Mkdir(dataDir, 0755)

	ioutil.WriteFile(path.Join(dataDir, "cert.pem"), demoCert, 0644)
	ioutil.WriteFile(path.Join(dataDir, "key.pem"), demoKey, 0644)

	wr, _ := os.OpenFile(configFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)

	data := map[string]interface{}{
		"Datadir": dataDir,
	}

	err := template.Must(template.New("name").Parse(demoConfig)).Execute(wr, data)
	if err != nil {
		logger.Info("demo", "Failed to write demo configuration: %s", err.Error())
		exit(1)
	}

	_, err = os.Stat(path.Join(dataDir, "cluster.json"))
	if err != nil {
		initCore(cmd, arguments)
	}

	runCore(cmd, arguments)
}
