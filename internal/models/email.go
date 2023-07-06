package models

import (
	mail "github.com/xhit/go-simple-mail/v2"
)

type EmailRequest struct {
	To       []string
	Subject  string
	Body     string
	Template string
	Data     any
}

type AccountEmailTemplateData struct {
	User          *User
	VerficationId string
	VerficationLink string
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func SetupMailServer() *mail.SMTPServer {
	server := mail.NewSMTPClient()
	server.Host = "smtp.zerobit.tech"
	server.Port = 587 // SMTP Port 	465 (25 or 587 for non-SSL)
	server.Username = "qhttp@zerobit.tech"
	server.Password = "Zer0#2023"
	server.Encryption = mail.EncryptionTLS
	return server
}
