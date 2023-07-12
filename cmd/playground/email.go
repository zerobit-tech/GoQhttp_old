package main

import (
	"bytes"
	"html/template"
	"log"

	mail "github.com/xhit/go-simple-mail/v2"
)

type EmailRequest struct {
	To       []string
	Subject  string
	Body     string
	Template string
	Data     any
}

func SetupMailServer() *mail.SMTPServer {
	server := mail.NewSMTPClient()
	server.Host = "smtp.zerobit.tech"
	server.Port = 587 // SMTP Port 	465 (25 or 587 for non-SSL)
	server.Username = "qhttp@zerobit.tech"
	server.Password = "Zer0#2023"
	server.Encryption = mail.EncryptionTLS
	return server
}

func (r *EmailRequest) Send(server *mail.SMTPServer) {
	smtpClient, err := server.Connect()
	if err != nil {
		log.Fatal(err)
	}

	// Create email
	email := mail.NewMSG()
	email.SetFrom("qhttp@zerobit.tech")
	email.AddTo(r.To...)
	//email.AddCc("another_you@example.com")
	email.SetSubject(r.Subject)

	r.ParseTemplate()

	email.SetBody(mail.TextHTML, r.Body)

	//email.AddAttachment("super_cool_file.png")

	// Send email
	err = email.Send(smtpClient)
	if err != nil {
		log.Fatal(err)
	}
}

func (r *EmailRequest) ParseTemplate() error {

	t, err := template.ParseFiles(r.Template)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, r.Data); err != nil {
		return err
	}
	r.Body = buf.String()
	return nil
}
