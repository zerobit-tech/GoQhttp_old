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
	"github.com/go-playground/form"
	"github.com/onlysumitg/GoQhttp/env"
	"github.com/onlysumitg/GoQhttp/featureflags"

	"github.com/onlysumitg/GoQhttp/internal/dbserver"
	"github.com/onlysumitg/GoQhttp/internal/iwebsocket"
	"github.com/onlysumitg/GoQhttp/internal/models"
	"github.com/onlysumitg/GoQhttp/internal/storedProc"
	"github.com/onlysumitg/GoQhttp/logger"
	"github.com/onlysumitg/GoQhttp/utils/concurrent"

	mail "github.com/xhit/go-simple-mail/v2"
	bolt "go.etcd.io/bbolt"

	_ "github.com/onlysumitg/GoQhttp/internal/ibmiServer"
	//_ "github.com/onlysumitg/GoQhttp/internal/mssqlServer"
	//_ "github.com/onlysumitg/GoQhttp/internal/mysqlServer"
)

type application struct {
	tlsCertificate *tls.Certificate
	tlsMutex       sync.Mutex

	version         string
	endPointMutex   sync.Mutex
	requestMutex    sync.Mutex
	mainAppServer   *http.Server
	graphMutex      sync.Mutex
	requestLogMutex sync.Mutex

	invalidEndPointCache bool
	endPointCache        map[string]*storedProc.StoredProc

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

	//--------------------------------------- Setup shutdown  ----------------------------

	//shutDownctx, startShutdown := context.WithCancel(context.Background())
	//---------------------------------------  final app config ----------------------------
	app := &application{
		version:       "1.2.0",
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

		spCallLogModel:             &models.SPCallLogModel{DB: logdb, DataChan: make(chan models.SPCallLogEntry)},
		useHttps:                   true,
		maxAllowedEndPoints:        -1,
		maxAllowedEndPointsPerUser: -1,

		//redirectToHttps: params.redirectToHttps,
		domain:         params.domain,
		useletsencrypt: params.useletsencrypt,
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
	}

	if app.debugMode {
		app.maxAllowedEndPoints = 50
		app.maxAllowedEndPointsPerUser = 2

	}

	appFeatures, ok := featureflags.FeatureSetMap[strings.ToUpper(params.featureset)]
	if !ok {
		log.Fatalln("Feature Set is not defined!")
	}

	app.features = appFeatures

	//goroutine
	//go models.SaveLogs(app.LogDB)
	go logger.StartLogging(app.LogDB)

	//app.CreateHttpPathPermissions()
	return app

}

func (app *application) CleanupAndShutDown() {
	log.Println("Closing channels...")
	// if app.shutDownStart != nil {

	// 	//fmt.Println("Starting shut down>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>><<<<<<<<<<<<<<<<<<<<<<<    <<<<<<<<<<<<<<")
	// 	app.shutDownStart()
	// }

	close(app.Done)

	close(app.ToWSChan)

	// close(app.GraphChan)  // closed in TimeTook middleware

	log.Println("Closing database connections...")
	dbserver.CloseConnections()

	log.Println("Shutting down Server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

// 	log.Println("Server Shutdown Completed")
// }
