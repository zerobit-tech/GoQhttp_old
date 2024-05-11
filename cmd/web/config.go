package main

import (
	"context"
	"crypto/tls"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-co-op/gocron"
	"github.com/go-playground/form/v4"
	"github.com/zerobit-tech/GoQhttp/cliparams"
	"github.com/zerobit-tech/GoQhttp/env"
	"github.com/zerobit-tech/GoQhttp/featureflags"
	"github.com/zerobit-tech/GoQhttp/session"

	"github.com/zerobit-tech/GoQhttp/internal/dbserver"
	"github.com/zerobit-tech/GoQhttp/internal/endpoints"
	"github.com/zerobit-tech/GoQhttp/internal/iwebsocket"
	"github.com/zerobit-tech/GoQhttp/internal/models"
	"github.com/zerobit-tech/GoQhttp/internal/rpg"
	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
	"github.com/zerobit-tech/GoQhttp/logger"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"

	mail "github.com/xhit/go-simple-mail/v2"
	bolt "go.etcd.io/bbolt"
)

// -------------------------------------------------------------------------
//
// -------------------------------------------------------------------------
type application struct {
	tlsCertificate *tls.Certificate
	tlsMutex       sync.Mutex

	version       string
	endPointMutex sync.Mutex
	requestMutex  sync.Mutex
	mainAppServer *http.Server
	graphMutex    sync.Mutex

	cacheMutext sync.RWMutex

	invalidEndPointCache bool
	endPointCache        map[string]*storedProc.StoredProc

	errorLog *log.Logger
	infoLog  *log.Logger

	DB          *bolt.DB
	LogDB       *bolt.DB
	SystemLogDB *bolt.DB

	UserDB      *bolt.DB
	EmailServer *mail.SMTPServer

	templateCache map[string]*template.Template

	storedProcsTemplateCache map[string]*template.Template
	storedProcsTemplates     []string

	maxAllowedEndPoints        int
	maxAllowedEndPointsPerUser int

	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	users          *models.UserModel

	servers        *models.ServerModel
	storedProcs    *models.StoredProcModel
	spCallLogModel *models.SPCallLogModel

	paramRegexModel *models.ParamRegexModel

	InProduction bool
	hostURL      string
	domain       string

	useHttps       bool
	useletsencrypt bool

	debugMode bool

	//redirectToHttps bool

	ToWSChan  chan iwebsocket.WsServerPayload
	WSClients concurrent.MapInterface

	GraphData100 []*GraphStruc
	GraphData200 []*GraphStruc
	GraphData300 []*GraphStruc
	GraphData400 []*GraphStruc
	GraphData500 []*GraphStruc
	GraphStats   *GraphStats
	GraphStream  chan *GraphStruc

	Done chan any

	hasClosedGraphChan bool

	shutDownChan chan int // 1= restrt app  2= shutdown app
	// shutDownContextX context.Context
	// shutDownStart   context.CancelFunc

	features *featureflags.Features

	SystemLoggerChan chan *SystemLogEvent

	ServerPingScheduler *gocron.Scheduler

	// ------------ RPG ------------
	RpgParamModel    *rpg.RpgParamModel
	RpgEndpointModel *rpg.RpgEndpointModel

	// ----- Generic endpoint ----------------
	Endpoint  *endpoints.Endpoint
	Endpoints []*endpoints.Endpoint
}

// -------------------------------------------------------------------------
//
// -------------------------------------------------------------------------
func baseAppConfig(params cliparams.Parameters, db *bolt.DB, userdb *bolt.DB, logdb *bolt.DB, systemlogdb *bolt.DB, version string) *application {

	//--------------------------------------- Setup loggers ----------------------------
	infoLog := log.New(os.Stderr, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	//--------------------------------------- Setup form decoder ----------------------------
	formDecoder := form.NewDecoder()

	_, hostUrl := params.GetHttpAddress()

	//--------------------------------------- Setup shutdown  ----------------------------

	//shutDownctx, startShutdown := context.WithCancel(context.Background())
	//---------------------------------------  final app config ----------------------------
	app := &application{
		version:  version,
		errorLog: errorLog,
		infoLog:  infoLog,

		DB:          db,
		LogDB:       logdb,
		UserDB:      userdb,
		SystemLogDB: systemlogdb,

		EmailServer: models.SetupMailServer(),

		sessionManager: session.GetSessionManager(db),
		formDecoder:    formDecoder,
		users:          &models.UserModel{DB: userdb},

		hostURL: hostUrl,

		servers:     &models.ServerModel{DB: db},
		storedProcs: &models.StoredProcModel{DB: db},

		paramRegexModel: &models.ParamRegexModel{DB: db},

		spCallLogModel:             &models.SPCallLogModel{DB: logdb, DataChan: make(chan models.SPCallLogEntry)},
		useHttps:                   true,
		maxAllowedEndPoints:        -1,
		maxAllowedEndPointsPerUser: -1,

		//redirectToHttps: params.redirectToHttps,
		domain:         params.Domain,
		useletsencrypt: params.Useletsencrypt,
		ToWSChan:       make(chan iwebsocket.WsServerPayload),
		WSClients:      concurrent.NewSuperEfficientSyncMap(0),

		GraphData100: make([]*GraphStruc, 0, 200),
		GraphData200: make([]*GraphStruc, 0, 500),
		GraphData300: make([]*GraphStruc, 0, 200),
		GraphData400: make([]*GraphStruc, 0, 200),
		GraphData500: make([]*GraphStruc, 0, 500),
		GraphStats:   &GraphStats{},
		GraphStream:  make(chan *GraphStruc),
		Done:         make(chan any),

		shutDownChan: make(chan int), // 1= restrt app  2= shutdown app
		// shutDownContext: shutDownctx,
		// shutDownStart:   startShutdown,

		debugMode: env.IsInDebugMode(),

		SystemLoggerChan: make(chan *SystemLogEvent),

		//------------------RPG
		RpgParamModel:    &rpg.RpgParamModel{DB: db},
		RpgEndpointModel: &rpg.RpgEndpointModel{DB: db},
	}

	//--------------------------------------- Setup template cache ----------------------------
	templateCache, err := app.newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}
	app.templateCache = templateCache

	if app.debugMode {
		app.maxAllowedEndPoints = 50
		app.maxAllowedEndPointsPerUser = 2

	}

	appFeatures, ok := featureflags.FeatureSetMap[strings.ToUpper(params.Featureset)]
	if !ok {
		log.Fatalln("Feature Set is not defined!")
	}

	app.features = appFeatures

	app.LoadSPTemplates()

	//goroutine
	//go models.SaveLogs(app.LogDB)
	go logger.StartLogging(app.LogDB)

	go app.LoadDefaultParamValidatorRegex()
	//app.CreateHttpPathPermissions()

	app.onLoad()

	return app

}

// -------------------------------------------------------------------------
//
// -------------------------------------------------------------------------
func (app *application) CleanupAndShutDown() {

	if app.ServerPingScheduler != nil {
		log.Println("Stoping server pings...")
		go app.ServerPingScheduler.Clear()
		go app.ServerPingScheduler.Stop()

	}

	log.Println("Closing channels...")
	// if app.shutDownStart != nil {

	// 	//fmt.Println("Starting shut down>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>><<<<<<<<<<<<<<<<<<<<<<<    <<<<<<<<<<<<<<")
	// 	app.shutDownStart()
	// }

	close(app.Done)
	close(app.ToWSChan)
	close(app.SystemLoggerChan)

	// close(app.GraphChan)  // closed in TimeTook middleware

	log.Println("Closing database connections...")
	dbserver.CloseConnections()

	log.Println("Shutting down Server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {

		cancel()
		app.shutDownChan <- 2
	}()

	err := app.mainAppServer.Shutdown(ctx)

	if err != nil {
		log.Printf("Forced Server Shutdown:%+v\n", err)
	}

	log.Println("Server Shutdown Completed")

}

// func (app *application) cleanUpFunc() {
// 	log.Println("Closing channels..")

// 	close(app.GraphChan)
// 	close(app.ToWSChan)

// 	log.Println("Shutting down Server")
// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer func() {

// 		cancel()
// 		shutDownChan <- true
// 	}()

// 	err := server.Shutdown(ctx)

// 	if err != nil {
// 		log.Printf("Forced Server Shutdown:%+v\n", err)
// 	}

//		log.Println("Server Shutdown Completed")
//	}
//
// -------------------------------------------------------------------------
//
// -------------------------------------------------------------------------
func (app *application) allowHtmlTemplates() bool {
	if env.AllowHtmlTemplates() {
		return true
	}

	return app.features.AllowHtmlTemplates

}
