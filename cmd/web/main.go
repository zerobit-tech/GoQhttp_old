package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/onlysumitg/GoQhttp/env"
	"github.com/onlysumitg/GoQhttp/internal/models"
	"github.com/onlysumitg/GoQhttp/utils/concurrent"
	bolt "go.etcd.io/bbolt"
)

func main() {

	//validateSetup()

	gocron.SetPanicHandler(func(jobName string, _ interface{}) {
		fmt.Printf("Panic in job: %s", jobName)
		fmt.Println("Recovering")
		// if r := recover(); r != nil {
		// 	log.Println("Recovered in refreshSchedule", r)
		// }
	})

	log.Println("Initializing....")

	err := os.MkdirAll("./db", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("./env", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	myfile, e := os.Create("./env/.env")
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

	// go run ./cmd/web -port=4002 -host="localhost"
	// go run ./cmd/web -h  ==> help text
	// default value for addr => ":4000"

	// using single var
	// addr := flag.String("addr", ":4000", "HTTP work addess")
	// fmt.Printf("\nStarting servers at port %s", *addr)
	// err := http.ListenAndServe(*addr, getTestRoutes())

	//using struct

	//--------------------------------------- Setup CLI paramters ----------------------------
	params := &parameters{}
	params.Load()

	if params.validateSetup {
		validateSetup()
	}

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
	logdb, err := bolt.Open("db/log.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer logdb.Close()

	// --------------------------------------- Setup app config and dependency injection ----------------------------
	app := baseAppConfig(*params, db, userdb, logdb)
	routes := app.routes()
	app.batches()

	//--------------------------------------- Setup websockets ----------------------------
	go concurrent.RecoverAndRestart(10, "ListenToWsChannel", app.ListenToWsChannel) //goroutine
	go concurrent.RecoverAndRestart(10, "SendToWsChannel", app.SendToWsChannel)     //goroutine
	go concurrent.RecoverAndRestart(10, "CaptureGraphData", app.CaptureGraphData)   //goroutine

	go concurrent.RecoverAndRestart(10, "spCallLogModel:AddLogid", app.spCallLogModel.AddLogid) //goroutine

	go concurrent.RecoverAndRestart(10, "spCallLogModel:AddLogid", app.spCallLogModel.AddLogid)


	go models.SaveLogs(app.LogDB)
	
	addr, hostUrl := params.getHttpAddress()

	// this is short cut to create http.Server and  server.ListenAndServe()
	// err := http.ListenAndServe(params.addr, routes)

	app.mainAppServer = &http.Server{
		Addr:     addr,
		Handler:  routes,
		ErrorLog: app.errorLog,
	}

	//  --------------------------------------- Data clean up job----------------------------

	go app.clearLogsSchedular(db) //goroutine

	go app.refreshSchedule() //goroutine

	go app.pingServerSchedule() //goroutine
	//--------------------------------------- Create super user ----------------------------

	go app.CreateSuperUser(params.superuseremail, params.superuserpwd) //goroutine

	// --------------------- SINGAL HANDLER -------------------

	go initSignals(app.CleanupAndShutDown) //goroutine

	// ---------------------LOAD SERVER -------------------

	// profiling server
	debugMe(*params)

	log.Printf("QHttp is live at  %s \n", hostUrl)

	// go openbrowser(url)
	//if params.https {

	// Construct a tls.config
	//tlsConfig := app.getCertificateToUse()
	app.mainAppServer.TLSConfig = app.getCertificateToUse()
	err = app.mainAppServer.ListenAndServeTLS("", "")

	// } else {
	// 	err = server.ListenAndServe()

	// }
	if err != nil {
		log.Println(err)
	}

	<-app.shutDownChan
	// mux := http.NewServeMux()
	// mux.Handle("/", http.HandlerFunc(home))

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
func (app *application) refreshSchedule() {

	defer concurrent.Recoverer("Recovered in refreshSchedule")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	//return

	s := gocron.NewScheduler(time.Local)

	interval1 := env.GetEnvVariable("REFRESH_EVERY", "")
	if interval1 != "" {
		//s.Every("5m").Do(func(){ ... })
		//s.Every(interval1).Do(app.RefreshStoredProces)

		if app.features.Promotion {
			s.Every(interval1).Do(app.ProcessPromotions)
		}
	}

	interval2 := env.GetEnvVariable("REFRESH_AT", "")
	if interval2 != "" {
		//s.Every(1).Day().At("10:30;08:00").Do(func(){ ... })
		//s.Every(1).Day().At(interval2).Do(app.RefreshStoredProces)
		if app.features.Promotion {
			s.Every(1).Day().At(interval2).Do(app.ProcessPromotions)
		}
	}

	s.StartAsync()

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) pingServerSchedule() {

	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	//return

	s := gocron.NewScheduler(time.Local)
	s.Every("20s").Do(app.PingServers)
	s.StartAsync()

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
