package main

import (
	"html/template"
	"log"
	"os"
	"sync"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form"
	"github.com/onlysumitg/GoQhttp/internal/models"

	mail "github.com/xhit/go-simple-mail/v2"
	bolt "go.etcd.io/bbolt"
)

type application struct {
	endPointMutex        sync.Mutex
	requestMutex         sync.Mutex
	invalidEndPointCache bool
	endPointCache        map[string]*models.StoredProc

	errorLog *log.Logger
	infoLog  *log.Logger

	DB          *bolt.DB
	LogDB       *bolt.DB
	UserDB      *bolt.DB
	EmailServer *mail.SMTPServer

	templateCache map[string]*template.Template

	maxAllowedEndPoints        int
	maxAllowedEndPointsPerUser int

	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	users          *models.UserModel

	servers        *models.ServerModel
	storedProcs    *models.StoredProcModel
	spCallLogModel *models.SPCallLogModel

	InProduction bool
	hostURL      string
	domain       string

	useHttps       bool
	useletsencrypt bool

	testMode bool

	redirectToHttps bool
}

func baseAppConfig(params parameters, db *bolt.DB, userdb *bolt.DB, logdb *bolt.DB) *application {

	//--------------------------------------- Setup loggers ----------------------------
	infoLog := log.New(os.Stderr, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	//--------------------------------------- Setup template cache ----------------------------
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}
	//--------------------------------------- Setup form decoder ----------------------------
	formDecoder := form.NewDecoder()

	_, hostUrl := params.getHttpAddress()
	//---------------------------------------  final app config ----------------------------
	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		templateCache: templateCache,

		DB:          db,
		LogDB:       logdb,
		UserDB:      userdb,
		EmailServer: models.SetupMailServer(),

		sessionManager: getSessionManager(db),
		formDecoder:    formDecoder,
		users:          &models.UserModel{DB: userdb},

		hostURL: hostUrl,

		servers:     &models.ServerModel{DB: db},
		storedProcs: &models.StoredProcModel{DB: db},

		spCallLogModel:             &models.SPCallLogModel{DB: logdb, DataChan: make(chan models.SPCallLogEntry, 5000)},
		useHttps:                   params.https,
		maxAllowedEndPoints:        -1,
		maxAllowedEndPointsPerUser: -1,
		testMode:                   params.testmode,
		redirectToHttps:            params.redirectToHttps,
		domain:                     params.domain,
		useletsencrypt:             params.useletsencrypt,
	}

	if app.testMode {
		app.maxAllowedEndPoints = 50
		app.maxAllowedEndPointsPerUser = 2

	}



	//app.CreateHttpPathPermissions()
	return app

}
