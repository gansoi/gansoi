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
MIIFaDCCBFCgAwIBAgISBMvqVkf2GIVZsBiPpHDZiNHbMA0GCSqGSIb3DQEBCwUA
MEoxCzAJBgNVBAYTAlVTMRYwFAYDVQQKEw1MZXQncyBFbmNyeXB0MSMwIQYDVQQD
ExpMZXQncyBFbmNyeXB0IEF1dGhvcml0eSBYMzAeFw0xOTExMTQxMTE0MzFaFw0y
MDAyMTIxMTE0MzFaMBkxFzAVBgNVBAMTDmdhbnNvaS1kZXYuY29tMIIBIjANBgkq
hkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwZ3HdZQA97b6UKaYjwWNjevZaPd9nu1t
n5A36BrFA1adqHuLvbi/piSEgsuh/IUd4dr8EED2sEIZn4m7STfgeZg0AAy+MbIB
FDXko+3DNEEaoRZaZVpWeTIgPkWd1HiQTCygxdH4vlIRSo2eb+LeElbW1R3IOaTU
BQX8On7fMn30psZ/QmgmsjffhcrrFkBX2W0CUAsGNyeE/PJ/CEdyy+SvbmMsPONP
T0h0doGYRFRbP/TsvGQM+wJIKt3PKAQl1BDd6Uz4UIWhhBJrwzyiY7xllo8jaPv0
7tXoW4hmPhriG2EnTYzmgIlZRt1mDJ9SCrv+uxekwdGS0p1EapwqAwIDAQABo4IC
dzCCAnMwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEF
BQcDAjAMBgNVHRMBAf8EAjAAMB0GA1UdDgQWBBTrAFFX/s+KTglAgHCaiV1zJ58t
AzAfBgNVHSMEGDAWgBSoSmpjBH3duubRObemRWXv86jsoTBvBggrBgEFBQcBAQRj
MGEwLgYIKwYBBQUHMAGGImh0dHA6Ly9vY3NwLmludC14My5sZXRzZW5jcnlwdC5v
cmcwLwYIKwYBBQUHMAKGI2h0dHA6Ly9jZXJ0LmludC14My5sZXRzZW5jcnlwdC5v
cmcvMCsGA1UdEQQkMCKCECouZ2Fuc29pLWRldi5jb22CDmdhbnNvaS1kZXYuY29t
MEwGA1UdIARFMEMwCAYGZ4EMAQIBMDcGCysGAQQBgt8TAQEBMCgwJgYIKwYBBQUH
AgEWGmh0dHA6Ly9jcHMubGV0c2VuY3J5cHQub3JnMIIBBgYKKwYBBAHWeQIEAgSB
9wSB9ADyAHcAB7dcG+V9aP/xsMYdIxXHuuZXfFeUt2ruvGE6GmnTohwAAAFuadYk
dAAABAMASDBGAiEAlbJjYw+P9gVy7BEiAAKVeeO9C//s/MIh4QUWNV5gHlsCIQDx
w2EZD8nGHoOU0HNVw3hw9nOrL+/wQjvJWqqSGaDk8AB3APCVpFnyANGCQBAtL5OI
jq1L/h1H45nh0DSmsKiqjrJzAAABbmnWJjwAAAQDAEgwRgIhAIMg+4rnDdX2FZe0
YRg/wxcCvTo3LLlGxMZmrYLmw1uVAiEAw6cSvOUlI6J2WTXuEjJKqs4XU82d1qNA
q/xxIxkmcWswDQYJKoZIhvcNAQELBQADggEBAIc8FPIlpLvJKNjN1ZuD3pYVZmWB
q8rO7BJ4kybdvMsx27v/x56IHrFXww84PAt97iVOoDmSi6D//pRlTF6syTnYdYVe
2yr7QYSO+EmuL0ATpijQLMHh5kCr+YsyVnkQqxOoHvOl7rMhhhsaGCoMHSbV/vgD
i30nN4tZ6pkuCKgdJvRl2L4Ipg0LsDlTYvzL3g1g5gEi6GjjrXMarboUqM2AU2Uy
8PsEGA424js+04RE1D0/Vl8DS87D6jcb/BrOqdKUcE/O2gEfR2IPuvGPPBoRI4eD
FDgxJnolI3ZZjskczrHkQWkEhX9yXZ1KUY+zYnR+ZyVFbHPT9bgtmMerDvk=
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
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDBncd1lAD3tvpQ
ppiPBY2N69lo932e7W2fkDfoGsUDVp2oe4u9uL+mJISCy6H8hR3h2vwQQPawQhmf
ibtJN+B5mDQADL4xsgEUNeSj7cM0QRqhFlplWlZ5MiA+RZ3UeJBMLKDF0fi+UhFK
jZ5v4t4SVtbVHcg5pNQFBfw6ft8yffSmxn9CaCayN9+FyusWQFfZbQJQCwY3J4T8
8n8IR3LL5K9uYyw8409PSHR2gZhEVFs/9Oy8ZAz7Akgq3c8oBCXUEN3pTPhQhaGE
EmvDPKJjvGWWjyNo+/Tu1ehbiGY+GuIbYSdNjOaAiVlG3WYMn1IKu/67F6TB0ZLS
nURqnCoDAgMBAAECggEALiw/XeXe2TRUu5kPNsRfPqIpJeQnnBgJFY1iF8MSirpE
POLBr4v4gFmjFEIVrk/ckXkUtZeYfO42iBpjhJGDwduvQvfG/7jwku5LSWCxNav0
+seG5TbG+n4evFsDyaC64L+f061AQbr2jq35Q7t5tCvrbcV9c2nAejY0MYsCFVsT
eYwqIrcoKa2YdPPwSovNIMLawdJ69gk9SRcYGNC9PuVhKx55W8oCuERN24qXpcIv
ubByKo+pyV9FuFmQXjDiUEruaWjZJN+cPoIMiGes+Wt2jX/uQGGOFJCfpNvUL+gF
YoG8iYPttW1lKmBcTjrMAD8A+veVRYMr60+3X9FesQKBgQD5LSgNL8jh1huQRybD
Cd1TyWE7rMJWCFxjBACx3o7Y3GWs9HwfIIklmaRr/XFYaxL902MQiM1/h4VxfM3h
LgvYrvb79fHJE6YGJwDgQXytyU+DpmZc/dO2RzbrcC//6ThJ9dLG2NCWN7Zb9gdG
lNZ15sChiRoJw1B0qx/6/iwBGQKBgQDG6x7aIrf89IU8g3qlYPRG/2u4aLjkuIL5
cYKUV8inC5wqKAJlzZeP/zO7KanGNNqbQolP8OzRvOrYB33G9rj8MWoqNnXCh5HE
o4NmeyU1kGCZIWfK0dQ5o/1VDRiqR4QLmWzH9LQSJmiTmq8xfg5wBElxYCwnJ5Ph
yWjmK2obewKBgQC0Z4485eDiKVsxfWJCCa1V0VJ1myHkmg8RnT9HiGQR/JgcvbHz
82F4eKHDke5zXqqUXWy29uUZtSvXO83vf6ASFLM7Pxj8RR6KC8kllCRJypuoLFym
bRvQOAU6NrJ57QfU/IbLcwSqDdIZCKrB3lbba+MB0Tqp4OAqaA8ycL25WQKBgG5F
q43Lz2loeHAqhxQlImY1dJb/XvhUnS8BUR1BaNfJ5OGt2l5rUcke9aOdHiPl6wZF
JW/upKpE63+k74IcBkKwdiU+mZukohd7ef2W6PK/vf1F+o8CMX9evLKolvLWwInZ
nOdgbW7eYfzptJNgSUqp5bj3UqLGGwIBMKjgimr/AoGABj0mQj88NOud9/5iKViF
sdBPna3NO5/DW1MZKJdRofFtbiIfBDoKxNLQzXvENDYR+KL5I785lMoUVRIr5hx+
WRlSOenwI4iwx0HDOcl6ixxfmi/zfrNtP2PSBAzt19TKTDzcecXAI2RqxgyXgOG5
b9H9RN2xe9+HpIC6vL+juBU=
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
