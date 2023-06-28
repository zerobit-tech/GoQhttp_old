package main

import (
	"log"
	"os"

	"github.com/onlysumitg/GoQhttp/internal/models"

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

		EmailServer: models.SetupMailServer(),
	}

	//app.CreateHttpPathPermissions()
	return app

}
