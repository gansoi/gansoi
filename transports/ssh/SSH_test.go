package ssh

import (
	"testing"

	"github.com/gansoi/gansoi/transports"

	"golang.org/x/crypto/ssh"
)

func TestGenerateKey(t *testing.T) {
	key := generateKey()
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		t.Fatalf("generateKey() did not generate a parsable key: %s", err.Error())
	}
	pub := signer.PublicKey()
	if pub.Type() != "ssh-rsa" {
		t.Fatalf("Key type is wrong, format is '%s'", pub.Type())
	}
}

var _ transports.Transport = (*SSH)(nil)
