//go:generate moq -out mailer_moq.go . Mailer
package mailer

import (
	"crypto/tls"
	"time"

	"github.com/rotisserie/eris"
	mail "github.com/xhit/go-simple-mail/v2"
)

var (
	EncryptionSSLTLS   string = "SSLTLS"
	EncryptionSTARTTLS string = "STARTTLS"
)

type Mailer interface {
	SendHTMLMail(to []string, subject string, body string) error
}

type MailerImpl struct {
	From       string
	Username   string
	Password   string
	Host       string
	Port       int
	Encryption string
}

func NewMailer(from string, username string, password string, host string, port int, encryption string) Mailer {
	return &MailerImpl{
		From:       from,
		Username:   username,
		Password:   password,
		Host:       host,
		Port:       port,
		Encryption: encryption,
	}
}

func (mailer *MailerImpl) SendHTMLMail(to []string, subject string, body string) (err error) {
	server := mail.NewSMTPClient()

	server.Host = mailer.Host
	server.Port = mailer.Port
	server.Username = mailer.Username
	server.Password = mailer.Password

	switch mailer.Encryption {
	case EncryptionSSLTLS:
		server.Encryption = mail.EncryptionSSLTLS
	case EncryptionSTARTTLS:
		server.Encryption = mail.EncryptionSTARTTLS
	default:
		server.Encryption = mail.EncryptionNone
	}

	// - PLAIN (default)
	// - LOGIN
	// - CRAM-MD5
	// - None
	server.Authentication = mail.AuthPlain

	// Variable to keep alive connection
	// server.KeepAlive = false

	// Timeout for connect to SMTP Server
	server.ConnectTimeout = 10 * time.Second

	// Timeout for send the data and wait respond
	server.SendTimeout = 10 * time.Second

	// Set TLSConfig to provide custom TLS configuration. For example,
	// to skip TLS verification (useful for testing):
	server.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// SMTP client
	smtpClient, err := server.Connect()

	if err != nil {
		return eris.Wrap(err, "SendHTMLMail")
	}

	email := mail.NewMSG()
	email.SetFrom(mailer.From).SetSubject(subject).SetBody(mail.TextHTML, body)

	for _, dest := range to {
		email.AddTo(dest)
	}

	// email.AddAlternativeData(mail.TextHTML, []byte(htmlBody))

	if err := email.Send(smtpClient); err != nil {
		return eris.Wrap(err, "SendHTMLMail")
	}

	return nil
}
