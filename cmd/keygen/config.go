package main

import (
	"log"
	"os"

	mail "github.com/xhit/go-simple-mail/v2"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger

	EmailServer *mail.SMTPServer
}

func baseAppConfig(params parameters) *application {

	//--------------------------------------- Setup loggers ----------------------------
	infoLog := log.New(os.Stderr, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	//---------------------------------------  final app config ----------------------------
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,

		EmailServer: SetupMailServer(),
	}

	//app.CreateHttpPathPermissions()
	return app

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
