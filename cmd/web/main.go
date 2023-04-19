package main

import (
	"flag"
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

	err := os.MkdirAll("./db", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Initializing....")

	// go run ./cmd/web -port=4002 -host="localhost"
	// go run ./cmd/web -h  ==> help text
	// default value for addr => ":4000"

	// using single var
	// addr := flag.String("addr", ":4000", "HTTP work addess")
	// fmt.Printf("\nStarting servers at port %s", *addr)
	// err := http.ListenAndServe(*addr, getTestRoutes())

	//using struct

	//--------------------------------------- Setup CLI paramters ----------------------------
	var params parameters
	flag.StringVar(&params.host, "host", "", "Http Host Name")
	flag.IntVar(&params.port, "port", 4081, "Port")

	flag.StringVar(&params.superuseremail, "superuseremail", "admin2@example.com", "Super User email")
	flag.StringVar(&params.superuserpwd, "superuserpwd", "adminpass", "Super User password")

	flag.BoolVar(&params.https, "https", true, "Use http or https")
	flag.BoolVar(&params.useletsencrypt, "useletsencrypt", false, "Use let's encrypt ssl certificate")

	flag.BoolVar(&params.testmode, "testmode", false, "Enable test mode")
	flag.StringVar(&params.domain, "domain", "0.0.0.0", "Domain name")

	flag.BoolVar(&params.redirectToHttps, "redirecttohttps", false, "Redirect to https")

	flag.Parse()

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
	app := baseAppConfig(params, db, userdb, logdb)
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

	//--------------------------------------- Create super user ----------------------------

	go app.CreateSuperUser(params.superuseremail, params.superuserpwd)

	// go openbrowser(url)
	if params.https {

		// Construct a tls.config
		//tlsConfig := app.getCertificateToUse()
		server.TLSConfig = app.getCertificateToUse()
		err = server.ListenAndServeTLS("", "")

	} else {
		err = server.ListenAndServe()

	}
	log.Fatal(err)

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
