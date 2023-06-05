package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/onlysumitg/GoQhttp/env"
	"github.com/onlysumitg/GoQhttp/internal/models"
	bolt "go.etcd.io/bbolt"
)

func main() {
	log.Println("Initializing....")

	err := os.MkdirAll("./db", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll("./env", os.ModePerm)
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

	envPort := env.GetEnvVariable("PORT", "")

	port, err := strconv.Atoi(envPort)
	if err == nil {
		params.port = port
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
	go ListenToWsChannel()

	addr, hostUrl := params.getHttpAddress()

	log.Printf("GoQHttp is live at %s  :: %s \n", addr, hostUrl)

	// this is short cut to create http.Server and  server.ListenAndServe()
	// err := http.ListenAndServe(params.addr, routes)

	server := &http.Server{
		Addr:     addr,
		Handler:  routes,
		ErrorLog: app.errorLog,
	}

	//  --------------------------------------- Data clean up job----------------------------

	go app.clearLogsSchedular(db)

	go app.refreshSchedule()
	//--------------------------------------- Create super user ----------------------------

	go app.CreateSuperUser(params.superuseremail, params.superuserpwd)

	shutDownChan := make(chan bool)

	// --------------------- SINGAL HANDLER -------------------

	cleanUpFunc := func() {
		log.Println("Shutting down Server")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer func() {

			cancel()
			shutDownChan <- true
		}()

		err := server.Shutdown(ctx)

		if err != nil {
			log.Printf("Server Shutdown Failed:%+v\n", err)
		}

		log.Println("Server Shutdown Completed")
	}

	go initSignals(cleanUpFunc)

	// ---------------------LOAD SERVER -------------------

	// profiling server
	debugMe(*params)

	// go openbrowser(url)
	if params.https {

		// Construct a tls.config
		//tlsConfig := app.getCertificateToUse()
		server.TLSConfig = app.getCertificateToUse()
		err = server.ListenAndServeTLS("", "")

	} else {
		err = server.ListenAndServe()

	}
	if err != nil {
		//log.Fatal(err)
	}

	<-shutDownChan
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
	s := gocron.NewScheduler(time.Local)

	s.Every(1).Day().At("21:30").Do(func() {
		models.DailyDataCleanup(db)
	})
	s.StartAsync()

	//s.Jobs()

	if app.testMode {
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
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in refreshSchedule", r)
		}
	}()

	s := gocron.NewScheduler(time.Local)

	interval1 := env.GetEnvVariable("REFRESH_EVERY", "")
	if interval1 != "" {
		//s.Every("5m").Do(func(){ ... })
		//s.Every(interval1).Do(app.RefreshStoredProces)
		s.Every(interval1).Do(app.ProcessPromotions)
	}

	interval2 := env.GetEnvVariable("REFRESH_AT", "")
	if interval1 != "" {
		//s.Every(1).Day().At("10:30;08:00").Do(func(){ ... })
		//s.Every(1).Day().At(interval2).Do(app.RefreshStoredProces)
		s.Every(1).Day().At(interval2).Do(app.ProcessPromotions)

	}

	s.StartAsync()

}
