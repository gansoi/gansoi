package ssh

import (
	"errors"
	"net"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/gansoi/gansoi/plugins"
)

var (
	hostKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEA0E8EpAVCyesVefTh1Dflu+nxF7RvtUePA3ty89i8AEXWchi0
bpION33HYestvpe1Hufv/umZv31gd3Ovptm5c6EbcujNB06gqwckC8KZQaBWPNVE
isLE/sKiHEDDpsXfIXatdCGRhd5DWI7NbejebV3e7CJ3TPDFlViJ3bSJHHuKlywH
+rbFb5i+1FEYeE6daZF3Xr6EluqBmIca+oiGfjb1KYCOBUiYrDLUggKqGHTEonOK
g/OKAqOT16dzddRUWLzMczjAabPITKH1/42yTk5I9hEIl95txcSG8Kk5ao5n+0O8
ApUXvWtrKtsTWXx5WcW0sGyCMURI2w02D0vFlQIDAQABAoIBAGkhuEHftdman/gx
M1ib7YJti5RfKJHhT7h+MYIIPLoWhSiId2fmpu4yuNIek7PBdVPn0yRgWYxfhrRC
2r/m2sYU5XmVLQUnFce7juGwkRCiD2Qcbr9plWrOaGrB9PzSmM6Wyuv1lTBvAKZN
TDjQcpHX8o5Sxss6KT46tToDx79rYEIOlqKF8Jw3sRWyCMzWvzY+VZLlLsLOy4aq
T8vVQKwEq0FlWb/l48Kaf7KI0eFEpg3T2CTTftOic0bgizcYWtsOqcCFve7W8ezr
+Y5E8HhBJDuaiFeDnmT9yhmUSsDoR/jIPdWqdhtaVdTDeyDlDKfj/3TGcc1a6gAv
nYr1fHECgYEA0QpRhBBNlnaxWgnPt36ZvcT6hZ9g8qyHhbxh3/l5eUetMOrATuJr
BSocSOQLQPU99z90RUviyvBWWhePS1DKWV5g41QDRaoGZmMrw5J+tu6IJ+I0wuQP
zkZn2IDWn7iT4rUpmlUzsw/vnG3BxOGxWE83plgoVYGf8Kbswy6d7WMCgYEA/xqf
sBXnYQwYm8N0G+ojPfekAqwuOIl5EgvWytK0fhxWIfEpuwZDH9dsr7z7veMb/z4U
TqgEFjpSX4qBdpEO6zGEJqy59wttWnOtz6ZAXw2Iy6/XaYGPm1A1xeCuQH44/TNm
tDEEDbsSOmvKVLgARg+Ynjx7Ow15IXmeAUzTjqcCgYAk8HrZKHxdc1oBvCwdk9yd
ITrX9AMQvxYvtstg5dfma5kaRNF43x/kSL24z7uBXhT2JRtpx3ArRm8r+m/S8lLf
mgLrCQSPAe2hmDK6m5+SufILgjiqm9yDKgPdyV6j2N0hObxBTn8VTGKeNPMB1JQf
M9h0p61w9rgJsaWzFADKWwKBgHhIYGURXuC2aGABX5zN4EIWouGTt9N2lnx40pUd
lncj+0TTxj5l7H51/8w5CmX1UL0f00lnFU45v59G2lZCfUtONKZbqO+gFragdqLD
G/T1l7uatLINod20izY7A05rdYmFk/aAag3jV+vt6R/bBl4Ceo6Zq+0jJCsxgdvR
EqtfAoGAIruRQ95aH+WXr2KVtV6v0itw+WlPOgPSNVW7M/KToGzZqWt8cVJ8UhF3
OBKhVLAorKhSaJKwgBL+lQtUrFAWVc+9WJi31jyor9mLl0NlNSDI6ryoyJrv4KwO
iSzceXtv5J/U8hgRd2CmnnPnR2lr/HEhA3ke+aIpLIKUqcCf17U=
-----END RSA PRIVATE KEY-----
`)
)

func TestAgent(t *testing.T) {
	a := plugins.GetAgent("ssh")
	_ = a.(*SSH)
}

func TestDefaultPort(t *testing.T) {
	cases := map[string]string{
		"hello":     "hello:22",
		"hello:22":  "hello:22",
		"hello::22": "hello::22",
	}

	for input, expected := range cases {
		output := defaultPort(input)
		if output != expected {
			t.Fatalf("defaultPort() did not return what we expected, got %s, expected %s", output, expected)
		}
	}
}

func TestCheckFail(t *testing.T) {
	a := SSH{
		// This should fail fast.
		Address: "127.0.0.99:0",
	}

	result := plugins.NewAgentResult()
	err := a.Check(result)
	if err == nil {
		t.Fatalf("Check() did not fail")
	}
}

func acceptOneSSH(listener net.Listener, config *ssh.ServerConfig) {
	go func(config ssh.ServerConfig) {
		conn, _ := listener.Accept()
		ssh.NewServerConn(conn, &config)
		// Give the client some time.
		time.Sleep(time.Millisecond * 50)
		conn.Close()
	}(*config)
}

func TestCheck(t *testing.T) {
	config := &ssh.ServerConfig{
		NoClientAuth: false,
	}
	config.SetDefaults()
	config.PasswordCallback = func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
		return nil, errors.New("Password rejected")
	}

	key, _ := ssh.ParsePrivateKey(hostKey)
	config.AddHostKey(key)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() failed: %s", err.Error())
	}
	defer listener.Close()

	acceptOneSSH(listener, config)

	a := SSH{
		Address: listener.Addr().String(),
	}

	result := plugins.NewAgentResult()
	err = a.Check(result)
	if err != nil {
		t.Fatalf("Check() failed: %s", err.Error())
	}

	if result["FingerprintSHA256"] != "SHA256:fcUrNWbMpMoorKj/lRlydfbGNtTpDUH+0lRHDxJgut0" {
		t.Fatalf("SHA256 looks wrong, got %s", result["FingerprintSHA256"])
	}

	if result["FingerprintMD5"] != "29:a6:c7:ff:16:cb:a2:81:17:a3:4b:a7:95:70:32:e6" {
		t.Fatalf("MD5 looks wrong, got %s", result["FingerprintMD5"])
	}

	config.NoClientAuth = true

	acceptOneSSH(listener, config)

	//	config.PasswordCallback = nil
	err = a.Check(result)
	if err != nil {
		t.Fatalf("Check() failed: %s", err.Error())
	}
}

func TestBanner(t *testing.T) {
	config := &ssh.ServerConfig{
		NoClientAuth: false,
		BannerCallback: func(_ ssh.ConnMetadata) string {
			return "this is a banner"
		},
	}
	config.SetDefaults()
	config.PasswordCallback = func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
		return nil, errors.New("Password rejected")
	}

	key, _ := ssh.ParsePrivateKey(hostKey)
	config.AddHostKey(key)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() failed: %s", err.Error())
	}
	defer listener.Close()

	acceptOneSSH(listener, config)

	a := SSH{
		Address: listener.Addr().String(),
	}

	result := plugins.NewAgentResult()
	err = a.Check(result)
	if err != nil {
		t.Fatalf("Check() failed: %s", err.Error())
	}

	if result["Banner"] != "this is a banner" {
		t.Fatalf("Banner looks wrong, got %s", result["Banner"])
	}
}

var _ plugins.Agent = (*SSH)(nil)
