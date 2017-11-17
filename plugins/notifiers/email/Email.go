package email

import (
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/gansoi/gansoi/plugins"
)

type (
	// Email will send an email.
	Email struct {
		SMTP     string `json:"smtp" description:"The address of a SMTP server to use (host:port)"`
		Username string `json:"username" description:"Leave empty for no authentication"`
		Password string `json:"password"`
		From     string `json:"from" description:"A from: email address"`
		To       string `json:"to" description:"A to: email address"`
	}
)

func init() {
	plugins.RegisterNotifier("email", Email{})
}

func newEmail2(e *Email, text string) []byte {
	header := make(map[string]string)
	header["From"] = e.From
	header["To"] = e.To
	header["Date"] = time.Now().Format(time.RFC1123Z)
	header["Subject"] = text
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "7bit"
	header["Message-ID"] = "<5465465465.84222.54111@gansoimail.com>"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	message += "\r\n" + text

	return []byte(message)
}

// Notify implement plugins.Notifier
func (e *Email) Notify(text string) error {
	conn, err := net.Dial("tcp", e.SMTP)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, e.SMTP)
	if err != nil {
		return err
	}
	defer c.Close()

	err = c.Hello("test.gansoi-dev.com")
	if err != nil {
		return err
	}

	if e.Username != "" {
		auth := unencryptedAuth{smtp.PlainAuth(
			"",
			e.Username,
			e.Password,
			e.SMTP,
		)}
		err = c.Auth(auth)
		if err != nil {
			return err
		}
	}

	err = c.Mail(e.From)
	if err != nil {
		return err
	}

	err = c.Rcpt(e.To)
	if err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	w.Write(newEmail2(e, text))
	w.Close()
	return c.Quit()
}
