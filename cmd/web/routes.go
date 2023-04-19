package main

// import "net/http"

// func (app *application) testRoutes() *http.ServeMux {
// 	// http.HandleFunc(xx,yy) => this also use a pre built default ServerMux -->  var DefaultServeMux = &defaultServeMux
// 	mux := http.NewServeMux()

// 	// Test handlers
// 	mux.HandleFunc("/helloworld", app.helloworld)  // app route
// 	mux.HandleFunc("/template", templates)   // independent route

// 	// file downloader
// 	mux.HandleFunc("/download", downloadFileHandler)

// 	// static files => http://127.0.0.1:4000/static/
// 	fileServer := http.FileServer(http.Dir("./ui/static/"))

// 	// http.StripPrefix is a middle ware
// 	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

// 	return mux
// }

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/onlysumitg/GoQhttp/env"
	"github.com/onlysumitg/GoQhttp/ui" // New import
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func addMiddleWares(app *application, router *chi.Mux) {

	//fmt.Println(">app.redirectToHttps>>>>>>", app.redirectToHttps)
	// session middleware
	if app.redirectToHttps {
		router.Use(app.RedirectToHTTPS)
	}
	router.Use(app.sessionManager.LoadAndSave)

	// A good base middleware stack : inbuilt in chi
	router.Use(RequestID) //(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	//router.Use(middleware.Recoverer)
	router.Use(middleware.SetHeader("X-Frame-Options", "DENY"))

	router.Use(middleware.Heartbeat("/ping"))

	// CSRF
	// router.Use(noSurf)

	//router.Use(app.MustHasPathsPermission)

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func addStaticFiles(router *chi.Mux) {
	// Take the ui.Files embedded filesystem and convert it to a http.FS type so
	// that it satisfies the http.FileSystem interface. We then pass that to the
	// http.FileServer() function to create the file server handler.
	fileServer := http.FileServer(http.FS(ui.Files))

	// Our static files are contained in the "static" folder of the ui.Files
	// embedded filesystem. So, for example, our CSS stylesheet is located at
	// "static/css/main.css". This means that we now longer need to strip the
	// prefix from the request URL -- any requests that start with /static/ can
	// just be passed directly to the file server and the corresponding static
	// file will be served (so long as it exists).
	// router.Handle("/static/*", http.StripPrefix("/static", fileServer))
	router.Handle("/static/*", fileServer)
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) routes() *chi.Mux {

	allowedOrigins := env.GetEnvVariable("ALLOWEDORIGINS", "https://*,http://*")

	allowedOriginList := strings.Split(allowedOrigins, ",")

	router := chi.NewRouter()

	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	router.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: allowedOriginList, // []string{"https://*", "http://*"},

		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},

		ExposedHeaders: []string{"Link"},

		AllowCredentials: false,

		MaxAge: 300, // Maximum value not ignored by any of major browsers
	}))

	addMiddleWares(app, router)

	addStaticFiles(router)

	router.Get("/", app.langingPage)
	router.Get("/help", app.helpPage)
	router.Get("/testmode", app.testModePage)

	app.APIHandlers(router)
	app.APILogHandlers(router)
	app.ServerHandlers(router)
	app.StoredProcHandlers(router)

	//app.WsHandlers(router)

	app.UserHandlers(router)
	app.UsersHandlers(router)
	// app.RbacHandlers(router)

	return router // standard.Then(router)
}
