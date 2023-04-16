package main

import (
	"log"

	"github.com/onlysumitg/GoQhttp/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
// func (app *application) SampleEmail() {
// 	e := &models.EmailRequest{
// 		To:       []string{"onlysumitg@gmail.com"},
// 		Subject:  "Test email 32",
// 		Body:     " this is test email2",
// 		Template: "email_verify_email.tmpl",
// 	}

// 	app.SendEmail(e)
// }

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SendEmail(r *models.EmailRequest) {

	if r == nil {
		return
	}

	smtpClient, err := app.EmailServer.Connect()
	if err != nil {
		log.Fatal(err)
	}

	// Create email
	email := mail.NewMSG()
	email.SetFrom("support@zerobit.tech")
	email.AddTo(r.To...)
	//email.AddCc("another_you@example.com")
	email.SetSubject(r.Subject)

	if r.Template != "" {
		tBody, err := app.templateToString(r.Template, r.Data)

		if err == nil && tBody != "" {
			r.Body = tBody
		}
	}

	email.SetBody(mail.TextHTML, r.Body)

	//email.AddAttachment("super_cool_file.png")

	// Send email
	err = email.Send(smtpClient)
	if err != nil {
		log.Fatal(err)
	}
}


func (a *application)SendNotificationsToAdmins(r *models.EmailRequest){
	emails := make([]string,0)
	for _,u:= range a.users.List(){
		if u.IsSuperUser {
			emails = append(emails, u.Email)
		}
	}

	r.To = emails

	a.SendEmail(r)
}