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
MIIGGTCCBQGgAwIBAgISA4aQb3/pJ1qEMuzkvUQkwfsxMA0GCSqGSIb3DQEBCwUA
MEoxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MSMwIQYDVQQD
ExpMZXQncyBFbmNyeXB0IEF1dGhvcml0eSBYMzAeFw0xODEwMjcwODI0MjdaFw0x
OTAxMjUwODI0MjdaMBkxFzAVBgNVBAMTDmdhbnNvaS1kZXYuY29tMIIBIjANBgkq
hkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAscbrApl7z5XTOMza7nca1JUvdl4/tf4K
pJz0LAk55/2NQsXzCN6AJu6XWcAIkt5oU5J5UiKWWZKe4dMfy4oHD39iTlKPft98
zS7mJIyEEVH+tq/xg70H7XuBlgB9OGm4fBtHrOGsRPCowIsQXsn1CAUD7UlUccZG
HPons9peeTuup/boL57kW9SBRYQYC6i+ztfJZibZgZXh2qtBzE32JYSJZUwh5BdK
+o+sI8arxkcWhElbVfAD7gexHmZLebgCxIHI4X/Dk0BxjF2mwqtjfcr1JHy5bh4v
/uTByrsoM7VBvtsTHC5hub6XnGhBQVEKU+ml1gE9DKZXp9+9bUhGnwIDAQABo4ID
KDCCAyQwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEF
BQcDAjAMBgNVHRMBAf8EAjAAMB0GA1UdDgQWBBSo4+cC0LQ9wXqFEQA2mYcM+7U6
wjAfBgNVHSMEGDAWgBSoSmpjBH3duubRObemRWXv86jsoTBvBggrBgEFBQcBAQRj
MGEwLgYIKwYBBQUHMAGGImh0dHA6Ly9vY3NwLmludC14My5sZXRzZW5jcnlwdC5v
cmcwLwYIKwYBBQUHMAKGI2h0dHA6Ly9jZXJ0LmludC14My5sZXRzZW5jcnlwdC5v
cmcvMCsGA1UdEQQkMCKCECouZ2Fuc29pLWRldi5jb22CDmdhbnNvaS1kZXYuY29t
MIH+BgNVHSAEgfYwgfMwCAYGZ4EMAQIBMIHmBgsrBgEEAYLfEwEBATCB1jAmBggr
BgEFBQcCARYaaHR0cDovL2Nwcy5sZXRzZW5jcnlwdC5vcmcwgasGCCsGAQUFBwIC
MIGeDIGbVGhpcyBDZXJ0aWZpY2F0ZSBtYXkgb25seSBiZSByZWxpZWQgdXBvbiBi
eSBSZWx5aW5nIFBhcnRpZXMgYW5kIG9ubHkgaW4gYWNjb3JkYW5jZSB3aXRoIHRo
ZSBDZXJ0aWZpY2F0ZSBQb2xpY3kgZm91bmQgYXQgaHR0cHM6Ly9sZXRzZW5jcnlw
dC5vcmcvcmVwb3NpdG9yeS8wggEEBgorBgEEAdZ5AgQCBIH1BIHyAPAAdgApPFGW
VMg5ZbqqUPxYB9S3b79Yeily3KTDDPTlRUf0eAAAAWa01swcAAAEAwBHMEUCIQC5
BQlM5O98vkopD1pb/Q6IZc0sLKcqrPBEko26qTJGVwIgOHd703THZiBQX8crLAdh
f/C5+piAIcUkiumx1gVxFNUAdgDiaUuuJujpQAnohhu2O4PUPuf+dIj7pI8okwGd
3fHb/gAAAWa01s2DAAAEAwBHMEUCIQDapfuDrxwzLIqWdqLquNBvH59MiRyYXQJ7
Qa5MoXCaLQIgHGn82+ZKYP3q3bH4B7fPk17v4ZaAOYQGtL3TD9obU38wDQYJKoZI
hvcNAQELBQADggEBAEa245KAdZJYkAj4CfKOhqKucuWv7wDk+YzAugIqbRXaB+yk
U5uPFH2Uc4UjmLEUDU9jK5/UnD/G/vJSRcrFhjdlq+9IM4LOMBk+gNBl/0LyuVOg
caOh0eWfQcqoqxnrCq+zc/f5eWAcwTk0FKi81nVp9elZMUqaAmfQgoyZIThBU3Bv
nYKBIqKIfGqzga0vQnjmnpcdNafHYapIsx29wzqc1Bid/yztSXO0mwPZdfi/qYd/
C7/3k2y1RayYYgmdhWxs9mYn9RnUyQ+DpNENxE93rRuFNYuSKp/kpZN+nJkPP7Um
LvXt+2QdtGIrnIx62+sGcy5nvaKS7NrykhNPdXA=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIEkjCCA3qgAwIBAgIQCgFBQgAAAVOFc2oLheynCDANBgkqhkiG9w0BAQsFADA/
MSQwIgYDVQQKExtEaWdpdGFsIFNpZ25hdHVyZSBUcnVzdCBDby4xFzAVBgNVBAMT
DkRTVCBSb290IENBIFgzMB4XDTE2MDMxNzE2NDA0NloXDTIxMDMxNzE2NDA0Nlow
SjELMAkGA1UEBhMCVVMxFjAUBgNVBAoTDUxldCdzIEVuY3J5cHQxIzAhBgNVBAMT
GkxldCdzIEVuY3J5cHQgQXV0aG9yaXR5IFgzMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAnNMM8FrlLke3cl03g7NoYzDq1zUmGSXhvb418XCSL7e4S0EF
q6meNQhY7LEqxGiHC6PjdeTm86dicbp5gWAf15Gan/PQeGdxyGkOlZHP/uaZ6WA8
SMx+yk13EiSdRxta67nsHjcAHJyse6cF6s5K671B5TaYucv9bTyWaN8jKkKQDIZ0
Z8h/pZq4UmEUEz9l6YKHy9v6Dlb2honzhT+Xhq+w3Brvaw2VFn3EK6BlspkENnWA
a6xK8xuQSXgvopZPKiAlKQTGdMDQMc2PMTiVFrqoM7hD8bEfwzB/onkxEz0tNvjj
/PIzark5McWvxI0NHWQWM6r6hCm21AvA2H3DkwIDAQABo4IBfTCCAXkwEgYDVR0T
AQH/BAgwBgEB/wIBADAOBgNVHQ8BAf8EBAMCAYYwfwYIKwYBBQUHAQEEczBxMDIG
CCsGAQUFBzABhiZodHRwOi8vaXNyZy50cnVzdGlkLm9jc3AuaWRlbnRydXN0LmNv
bTA7BggrBgEFBQcwAoYvaHR0cDovL2FwcHMuaWRlbnRydXN0LmNvbS9yb290cy9k
c3Ryb290Y2F4My5wN2MwHwYDVR0jBBgwFoAUxKexpHsscfrb4UuQdf/EFWCFiRAw
VAYDVR0gBE0wSzAIBgZngQwBAgEwPwYLKwYBBAGC3xMBAQEwMDAuBggrBgEFBQcC
ARYiaHR0cDovL2Nwcy5yb290LXgxLmxldHNlbmNyeXB0Lm9yZzA8BgNVHR8ENTAz
MDGgL6AthitodHRwOi8vY3JsLmlkZW50cnVzdC5jb20vRFNUUk9PVENBWDNDUkwu
Y3JsMB0GA1UdDgQWBBSoSmpjBH3duubRObemRWXv86jsoTANBgkqhkiG9w0BAQsF
AAOCAQEA3TPXEfNjWDjdGBX7CVW+dla5cEilaUcne8IkCJLxWh9KEik3JHRRHGJo
uM2VcGfl96S8TihRzZvoroed6ti6WqEBmtzw3Wodatg+VyOeph4EYpr/1wXKtx8/
wApIvJSwtmVi4MFU5aMqrSDE6ea73Mj2tcMyo5jMd6jmeWUHK8so/joWUoHOUgwu
X4Po1QYz+3dszkDqMp4fklxBwXRsW10KXzPMTZ+sOPAveyxindmjkW8lGy+QsRlG
PfZ+G6Z6h7mjem0Y+iWlkYcV4PIWL1iwBi8saCbGS5jN2p8M+X+Q7UNKEkROb3N6
KOqkqm57TH2H3eDJAkSnh6/DNFu0Qg==
-----END CERTIFICATE-----
`)

	demoKey = []byte(`-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCxxusCmXvPldM4
zNrudxrUlS92Xj+1/gqknPQsCTnn/Y1CxfMI3oAm7pdZwAiS3mhTknlSIpZZkp7h
0x/LigcPf2JOUo9+33zNLuYkjIQRUf62r/GDvQfte4GWAH04abh8G0es4axE8KjA
ixBeyfUIBQPtSVRxxkYc+iez2l55O66n9ugvnuRb1IFFhBgLqL7O18lmJtmBleHa
q0HMTfYlhIllTCHkF0r6j6wjxqvGRxaESVtV8APuB7EeZkt5uALEgcjhf8OTQHGM
XabCq2N9yvUkfLluHi/+5MHKuygztUG+2xMcLmG5vpecaEFBUQpT6aXWAT0Mplen
371tSEafAgMBAAECggEAG9x/8l1ZkRP7EXjRivPxqYVj9doZhA03X8sVXV1ozNno
7KEXULmGPhPAdplo/pOKqCZZiyzOgWVAL0YmQoD0UFJ3dqzrvkeKSKHkAbBf9lLy
Z3E1mZ7jgi2MBpU3CsNO3WxtFEQd+oP4/owM2b4u/73Besu2R0p6rInr9PzxN3CG
9KjN+L1819glnFjIFrpVUnvzRHx9yXqKBTaWlkdGU5YR1pLNu1NGCHd8X5A5rO+g
ngw31XGY2J4PFU3CWdM8G2Ok9UZeoaP5DlmtPOZa2KdNhyaloiUgJJZemm0lvmQe
ZvFhIjwkCNZtdlVtwv7e9Wg3WSKB0xIIzpqJ5hggqQKBgQDovMuRpYwfzgjMnP3H
9cIo4rYEtKNefefbuQ2bprC4gqmRZSOoc98jG0Vr7TbAAaG+tqUfxka194GORXz6
M1klc06peCfzNPEKeM9k6vjUmMhuG+7Cb1p2/J+jyxLBly6nx9I0leNJLMOeBoaG
LSzSpbo2YzggBZ6kluP49mr0PQKBgQDDi9BfZh7kOYatpGb/KvOR8GMgbgYg0C1z
YbtDrGJ6gV8u0Y7huC0ntEhCx3VlMk5q1C0Fgh3YMG0RyjW4RgHMRzYOHan3sva6
imcPp7P11S3XgOLJ1ksH5vb/uG3/lbmNZsbd/GEF1TmgteM5T5h93Gp82VqMk0i0
q6gVTcNoCwKBgH0DpIl4ljsjTgCyt3MoZIHXvZPrf/GqyddxoIiiUjzaGsF5xVyf
2RUfefvOMOGUPtCVhT77H1JxP4svckFnQZRnSrKnUzZFktBgMm7v9HcmFktm/6o7
bCmL6yhtVYbdcXc4U4ZhmPPecrk4ohwCuhfwY4UBaM1yl1OrpYs+uha1AoGBAKSm
Pud0xyIHxfzDZu4Hrlr9e5m8ynAqmXqDmfgbWDgqYoi7boFkYLEDvaxs4c45k4mj
6dyveI2mCNBn4N+uIqlsdHliRNEJ4XGkQ68M2BxwSU8heNPWGgsnUGZ1VjlUSo5j
MaOMc+1DYmiNfiutd4rSYt+I7BNdmGR3/OkjNxuNAoGAHlsggnBJ4jH/M6lO/Fer
RoYZ7Edfe8LyNXd2eWA++iAsZsL+Rhac0wjh9gV27+YSLoXYbKFa8Ng/H2u9hoFe
h451JV37Zcp9+/EWUj1L4kuiDC/jjbA6GqvigBh7ROcuq8FMhAkp6f/RhxYlZ0oR
OuMokOpVzrQgx0SuJNiMah8=
-----END PRIVATE KEY-----
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
