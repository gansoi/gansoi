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
MIIFZjCCBE6gAwIBAgISA3a3XdIcCh3auoN5IKfy4vxJMA0GCSqGSIb3DQEBCwUA
MEoxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MSMwIQYDVQQD
ExpMZXQncyBFbmNyeXB0IEF1dGhvcml0eSBYMzAeFw0xOTA0MjYwMDQ1MTRaFw0x
OTA3MjUwMDQ1MTRaMBkxFzAVBgNVBAMTDmdhbnNvaS1kZXYuY29tMIIBIjANBgkq
hkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAp3A9I13miwdZd6Vw4uYbBaT4OtRhwk5H
LJpb2PGAGoFEhfYpOj7PFQ614FRr0QwboE1CDyBWSgKNH6dGRCgRCWttZ65dcyeQ
gJ5mbpNbZe5Pmp2Eg/M9awZ05ptxEqx9s9PERLmmlpAVxerGhmMMNnmaJidMbLwA
gtl2rkJxjohW7pjRsRRi2X2DTLa8VIX5By8Ewr4ZShHlKqUajhBGcpoI14NIFMO7
h4o25FLKhPG2dPx6Km8aYjBRL5KXjsqHwK9c773rAyJpmPZU2z/Fwt5wMj5Mdvp7
V0FqwG9sIjzsC5MmKQOtP+Q1/ZPOh7e3QF4HUZ8jmer+pv/jFGVuXQIDAQABo4IC
dTCCAnEwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEF
BQcDAjAMBgNVHRMBAf8EAjAAMB0GA1UdDgQWBBRD0wFMFRDTi37iJtkCiZL9zgMA
9jAfBgNVHSMEGDAWgBSoSmpjBH3duubRObemRWXv86jsoTBvBggrBgEFBQcBAQRj
MGEwLgYIKwYBBQUHMAGGImh0dHA6Ly9vY3NwLmludC14My5sZXRzZW5jcnlwdC5v
cmcwLwYIKwYBBQUHMAKGI2h0dHA6Ly9jZXJ0LmludC14My5sZXRzZW5jcnlwdC5v
cmcvMCsGA1UdEQQkMCKCECouZ2Fuc29pLWRldi5jb22CDmdhbnNvaS1kZXYuY29t
MEwGA1UdIARFMEMwCAYGZ4EMAQIBMDcGCysGAQQBgt8TAQEBMCgwJgYIKwYBBQUH
AgEWGmh0dHA6Ly9jcHMubGV0c2VuY3J5cHQub3JnMIIBBAYKKwYBBAHWeQIEAgSB
9QSB8gDwAHYAdH7agzGtMxCRIZzOJU9CcMK//V5CIAjGNzV55hB7zFYAAAFqV1Fq
zQAABAMARzBFAiAWb44adGHbIOF2jusnA+ur3/QCToViJXe92ykBQZQ0KwIhALiy
Q02DXbK+EkuPNpVZmpApJ6Y/X+R5CSQSQl8kKVmLAHYAKTxRllTIOWW6qlD8WAfU
t2+/WHopctykwwz05UVH9HgAAAFqV1FqvwAABAMARzBFAiAHg+jW6vzvMzYtEtZF
lOOoClUIreahJV+xRlwVjENTcAIhAIyA+O8D/x5XrsjCpKCnwGoozcgc6tffmngv
PhQvYTEHMA0GCSqGSIb3DQEBCwUAA4IBAQBJECvp3B5sAU6LGg3Kh4fAPnAVYQQx
wTWFpxnnDlpishi+37q0tOubefIgpAKP7RBTimyG4g5fwrA3kYw9fjI63MUM6Tcj
r9bx/73NDizdx1rCLdjWLiB6BsqBypbX4NABVRWA9WlMKsJCPmmvR+xmvw5Ud4OW
SlFrulXkT/KGGh+loWyrwHnpcEBWCswZQQSM2Bcd4YW06UiUDngWKjRfeJCGRZGn
feRd6uYZwsmrsAoEGU50Okyxick4BlsSaaTflmYxGPcD8pBcYDnLcfB+G+ZegvDC
hnuaItwGyYDqdUz3NOGCdium91eVbW0eXjekYjnV1YHHO4XacXOdpn5t
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
MIIEwAIBADANBgkqhkiG9w0BAQEFAASCBKowggSmAgEAAoIBAQCncD0jXeaLB1l3
pXDi5hsFpPg61GHCTkcsmlvY8YAagUSF9ik6Ps8VDrXgVGvRDBugTUIPIFZKAo0f
p0ZEKBEJa21nrl1zJ5CAnmZuk1tl7k+anYSD8z1rBnTmm3ESrH2z08REuaaWkBXF
6saGYww2eZomJ0xsvACC2XauQnGOiFbumNGxFGLZfYNMtrxUhfkHLwTCvhlKEeUq
pRqOEEZymgjXg0gUw7uHijbkUsqE8bZ0/HoqbxpiMFEvkpeOyofAr1zvvesDImmY
9lTbP8XC3nAyPkx2+ntXQWrAb2wiPOwLkyYpA60/5DX9k86Ht7dAXgdRnyOZ6v6m
/+MUZW5dAgMBAAECggEBAIqZXPJuowzIi6V/nB8mDI4yjYvdDAMUWtQv/gFpfvnx
sTAWoO/m1/H9WE4Wc5z6oc+ixCDTSro+vGTMSVzXsaqTc1+UtSRCRtpAxFeZwNly
WBCLfQrgiSGTmabeWs3zXKZDkD63Uo9lO7C185mAWbaqGBCnDRsg0Gd/7T64H1m0
aUbUlN8anAHTOg7FuAOh9ZnvUjzuSRImkd7QhMfaxnxlp8wEub248eJ2/tdZWo/9
+feP1pRIT99h7fAzZt1xti7oqeYXocNJHgHcQhxKUCzh6qOLuaEeX3KdBtClMwZf
BWtQEfp1vlclhTan7kHyYVB8VLzLkZ4/rFkCMDhCb10CgYEAz2Wih+kK7Y7QtMJW
dpiIAFibhsxdSIjbM29T3FauvuPzgVBPCN9WJ0/lb5K+1BirAy1w5cugPWcXPviu
VYl4qqkGP974Xv2ykzro6ogizgsbZVJbaLFeGxpFr3ESkpec1IVjXcdwXo4oVWsJ
zZk9ypCUIoGR3UWGqiBq3ce2wDMCgYEAzq1eor6/arWueSCBvFbtADf5b3R/+sZ0
Eu50gogAjFKU+VBR41GPMufC5f2L0Wt3eJ64So9oy66434qYZmQEf2m6p6dY3WLk
Yqh4L+fi76bz0yj7STnc6PiDbb2u1sJjq8nLDgHY9K4jgR20s44UIVPZtL/RLmfA
pRDpaylORy8CgYEAomzYk1Rh6JaWwVoJ2SuwJ4U9mJ5o2CvGcEvaA5AnnvxrcnTw
0OQIVxVtfKoSpY0EMaLe7jlBL5WTUQ/1iKv3QsDJUjBPmCVcO45BZ5xilbeBJyMO
z7lJGDTFKpeFkiUHADHPFlzhlkwDLpJ7xPcku4XrXfLXx9Fxm79zsml8qJ0CgYEA
hiHESnDrP/D7c8ciC76Kf9rT+n6sVG+Zg3PYj6J+YJNz9b4n9xTGj+4T8jm0FUze
s5ym1FTwWUhi2UlUkRcWSC6xIf+cz8lPmavmnIitIoXCtmdhdmukMFqgyYcSQnTU
RPho3QVb7ahaiIBj9ygdbmfB4emfc7OINlicdJzRHacCgYEAqIVrMvdiCwySGcAV
/C+bu8NXdulpkKpM3AwxDWeZ4gm/riIIQjloMS0B3/m3IqNVm04m8EGuol/XW+Ja
Of+GS9cjy5HKJ3+gbp1DB/HOZDiRYiMlDTMBelarKXiGb4G/fT/0NqDNLpCnlk7A
cGPgWP/5ZNeLaIJgxnkg3nDs3eo=
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
