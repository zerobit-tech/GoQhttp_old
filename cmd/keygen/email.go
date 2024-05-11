package main

import (
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"time"

	"github.com/jprobinson/eazye"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"

	mail "github.com/xhit/go-simple-mail/v2"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
// func (app *application) SampleEmail() {
// 	e := &models.EmailRequest{
// 		To:       []string{"support@zerobit.tech"},
// 		Subject:  "Test email 32",
// 		Body:     " this is test email2",
// 		Template: "email_verify_email.tmpl",
// 	}

//		app.SendEmail(e)
//	}
type EmailRequest struct {
	To       []string
	Subject  string
	Body     string
	Template string
	Data     any
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SendEmail(r *EmailRequest) {

	if r == nil {
		return
	}

	smtpClient, err := app.EmailServer.Connect()
	if err != nil {
		log.Println(err)
		return
	}

	// Create email
	email := mail.NewMSG()
	email.SetFrom("qhttp@zerobit.tech")
	email.AddTo(r.To...)
	//email.AddCc("another_you@example.com")
	email.SetSubject(r.Subject)

	email.SetBody(mail.TextHTML, r.Body)

	//email.AddAttachment("super_cool_file.png")

	// Send email
	err = email.Send(smtpClient)
	if err != nil {
		log.Println(err)
	}
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func ReadEmails(waitC chan<- int) {
	defer concurrent.Recoverer("GetByEmail")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	defer func() {
		waitC <- 1
	}()

	for {
		time.Sleep(10 * time.Second)
		log.Println("Checking mail box")
		mailBox := eazye.MailboxInfo{
			Host:               "smtp.zerobit.tech",
			TLS:                true,
			InsecureSkipVerify: true,
			User:               "qhttp@zerobit.tech",
			Pwd:                "Zer0#2023",
			Folder:             "inbox",
			ReadOnly:           false,
		}

		emails, errx := eazye.GetUnread(mailBox, true, false)
		if errx != nil {
			fmt.Println("eazye", errx)
		}

		for _, email := range emails {
			fmt.Println(email.To, " : : ", email.From, " :: ", email.Subject)

			if strings.EqualFold(strings.ToUpper(strings.TrimSpace(email.Subject)), "QHTTP LIC") {

				params := &parameters{
					client: email.From.Name,
					email:  email.From.Address,
					days:   35,
				}

				processLicRequest(params)
			}

		}

	}

}
