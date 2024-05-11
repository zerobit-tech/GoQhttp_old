package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/zerobit-tech/GoQhttp/cliparams"
	"github.com/zerobit-tech/GoQhttp/env"
	"github.com/zerobit-tech/GoQhttp/internal/models"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/acme/autocert"
)

var FeatureSet string = "ALL"
var Version string = "v1.0.0"

func main() {

	//validateSetup()

	gocron.SetPanicHandler(func(jobName string, _ interface{}) {
		fmt.Printf("Panic in job: %s", jobName)
		fmt.Println("Recovering")
		// if r := recover(); r != nil {
		// 	log.Println("Recovered in refreshSchedule", r)
		// }
	})

	today := time.Now().Local().Format(stringutils.ISODateFormat0)
	log.Println("Initializing....")

	createInitialFolders()

	//--------------------------------------- Setup CLI paramters ----------------------------
	params := &cliparams.Parameters{}
	params.Load()

	if params.ValidateSetup {
		validateSetup()
	}

	params.Featureset = FeatureSet

	// --------------------------------------- Setup database ----------------------------
	db, err := bolt.Open("db/internal.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// --------------------------------------- Setup database ----------------------------
	userdb, err := bolt.Open("db/user.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer userdb.Close()
	// --------------------------------------- Setup database ----------------------------
	logdb, err := bolt.Open(fmt.Sprintf("db/log_%s.db", today), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer logdb.Close()

	// --------------------------------------- Setup database ----------------------------
	systemlogdb, err := bolt.Open(fmt.Sprintf("db/systemlog_%s.db", today), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer logdb.Close()

	// --------------------------------------- Setup app config and dependency injection ----------------------------
	app := baseAppConfig(*params, db, userdb, logdb, systemlogdb, Version)
	routes := app.routes()
	app.batches()

	//--------------------------------------- Setup websockets ----------------------------
	go concurrent.RecoverAndRestart(10, "ListenToWsChannel", app.ListenToWsChannel) //goroutine
	go concurrent.RecoverAndRestart(10, "SendToWsChannel", app.SendDataTOWebSocket) //goroutine
	go concurrent.RecoverAndRestart(10, "CaptureGraphData", app.CaptureGraphData)   //goroutine

	go concurrent.RecoverAndRestart(10, "spCallLogModel:AddLogid", app.spCallLogModel.AddLogid) //goroutine
	go concurrent.RecoverAndRestart(10, "systemlogger", app.SystemLogger)                       //goroutine

	addr, hostUrl := params.GetHttpAddress()

	// this is short cut to create http.Server and  server.ListenAndServe()
	// err := http.ListenAndServe(params.addr, routes)

	app.mainAppServer = &http.Server{
		Addr:     addr,
		Handler:  routes,
		ErrorLog: app.errorLog,
	}

	//  --------------------------------------- Data clean up job----------------------------
	go app.clearLogsSchedular(db) //goroutine
	go app.promotionsSchedule()   //goroutine
	go app.pingServerSchedule()   //goroutine

	//--------------------------------------- Create super user ----------------------------
	go app.CreateSuperUser(params.Superuseremail, params.Superuserpwd) //goroutine
	// --------------------- SINGAL HANDLER -------------------
	go initSignals(app.CleanupAndShutDown) //goroutine
	// ---------------------LOAD SERVER -------------------

	// profiling server
	debugMe(*params)

	log.Println(qhttpTextArt)
	log.Printf("QHttp is live at  %s \n", hostUrl)

	// go openbrowser(url)
	if params.Https {

		// Construct a tls.config
		//tlsConfig := app.getCertificateToUse()
		var m *autocert.Manager
		app.mainAppServer.TLSConfig, m = app.getCertificateAndManager()

		// lets encrypt need port 80 to run verification
		if app.useletsencrypt {
			go concurrent.RecoverAndRestart(10, "http server", func() { http.ListenAndServe(":http", m.HTTPHandler(nil)) })
		}

		err = app.mainAppServer.ListenAndServeTLS("", "")

	} else {
		err = app.mainAppServer.ListenAndServe()

	}
	if err != nil {
		log.Println(err)
	}

	<-app.shutDownChan
	// mux := http.NewServeMux()
	// mux.Handle("/", http.HandlerFunc(home))

}

func createInitialFolders() {
	err := os.MkdirAll("./db", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("./env", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	myfile, e := os.OpenFile("./env/.env", os.O_RDWR|os.O_CREATE, 0666)
	if e != nil {
		log.Fatal(e)
	}
	myfile.Close()

	err = os.MkdirAll("./lic", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("./cert", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll("./templates", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

}

// -----------------------------------------------------------------
//  TO AUTO redict http to https
// -----------------------------------------------------------------
// func redirect(w http.ResponseWriter, req *http.Request) {
// 	http.Redirect(w, req,
// 		"https://"+req.Host+req.URL.String(),
// 		http.StatusMovedPermanently)
// }

// go func() {
//     if err := http.ListenAndServe(":80", http.HandlerFunc(redirectToTls)); err != nil {
//         log.Fatalf("ListenAndServe error: %v", err)
//     }
// }()

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) clearLogsSchedular(db *bolt.DB) {

	defer concurrent.Recoverer("clearLogsSchedular")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	s := gocron.NewScheduler(time.Local)

	s.Every(1).Day().At("21:30").Do(func() {
		models.DailyDataCleanup(db)
	})
	s.StartAsync()

	//s.Jobs()

	if app.debugMode {
		t := gocron.NewScheduler(time.Local)

		t.Every(1).Day().At("21:30").Do(func() {
			models.DailyDataCleanup_TESTMODE(db)
		})
		t.StartAsync()

	}
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) promotionsSchedule() {

	if !app.features.AllowPromotion {
		return
	}

	defer concurrent.Recoverer("Recovered in refreshSchedule")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	//return

	s := gocron.NewScheduler(time.Local)

	interval1 := env.GetEnvVariable("PROMOTE_EVERY", "")
	if interval1 != "" {
		//s.Every("5m").Do(func(){ ... })
		//s.Every(interval1).Do(app.RefreshStoredProces)

		if app.features.AllowPromotion {
			s.Every(interval1).Do(app.ProcessPromotions)
		}
	}

	interval2 := env.GetEnvVariable("PROMOTE_AT", "")
	if interval2 != "" {
		//s.Every(1).Day().At("10:30;08:00").Do(func(){ ... })
		//s.Every(1).Day().At(interval2).Do(app.RefreshStoredProces)
		if app.features.AllowPromotion {
			s.Every(1).Day().At(interval2).Do(app.ProcessPromotions)
		}
	}

	s.StartAsync()

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) pingServerSchedule() {
	defer concurrent.Recoverer("PingServer")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
	//ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	//return
	pingServerEvery := env.GetEnvVariable("PING_SERVER_EVERY", "20s")
	if pingServerEvery == "0" {
		return
	}

	app.ServerPingScheduler = gocron.NewScheduler(time.Local)

	//s.WithDistributedLocker()
	app.ServerPingScheduler.Every(pingServerEvery).Do(app.PingServers)
	//s.SingletonMode()

	app.ServerPingScheduler.StartAsync()

}

// // -----------------------------------------------------------------
// //
// // -----------------------------------------------------------------
// func (app *application) PingServersSchedular() {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			log.Println("Recovered in PingServersSchedular", r)
// 		}
// 	}()

// 	//return

// 	servers := app.servers.List()

// 	scheduedTime := len(servers) / 5 // 5 seconds per server

// 	s := gocron.NewScheduler(time.Local)

// 	interval1 := fmt.Sprintf("%ds", scheduedTime)
// 	if scheduedTime > 0 {

// 		s.Every(interval1).Do(app.ProcessPromotions)
// 	}

// 	s.StartAsync()

// }
