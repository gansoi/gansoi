package email

import (
	"net/smtp"
)

type unencryptedAuth struct {
	smtp.Auth
}

func (a unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	s := *server

	s.TLS = true

	return a.Auth.Start(&s)
}
