package main

import (
	"context"
	"log"
	"net/http"

	"github.com/justinas/nosurf" // New import
	"github.com/onlysumitg/GoQhttp/internal/models"
)

// Create a NoSurf middleware function which uses a customized CSRF cookie with
// the Secure, Path and HttpOnly attributes set.
func noSurf(next http.Handler) http.Handler {

	defaultFailureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(" :::::::::::: CSRF FAILED ::::::::::::::::", nosurf.Reason(r))
		http.Error(w, http.StatusText(400), 400)
	})

	csrfHandler := nosurf.New(next)
	// csrfHandler.SetBaseCookie(http.Cookie{
	// 	HttpOnly: true,
	// 	//Path:     "/",
	// 	//Secure: true,
	// })
	csrfHandler.SetFailureHandler(defaultFailureHandler)
	return csrfHandler
}

const (
	xForwardedProtoHeader = "x-forwarded-proto"
)

func (app *application) RedirectToHTTPS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//host, _, _ := net.SplitHostPort(r.Host)
		u := r.URL
		log.Println("starte", u.String(), "::", r.URL.Scheme, r.TLS, r.Host, r.RequestURI, "::", r.Header.Get(xForwardedProtoHeader))
		if r.Header.Get(xForwardedProtoHeader) != "https" {

			log.Println(":::::::: REDIRECTING :::::::::")
			sslUrl := "https://" + r.Host + r.RequestURI
			http.Redirect(w, r, sslUrl, http.StatusMovedPermanently)
			return
		}

		//log.Println(":::::::: NOT REDIRECTING :::::::::")

		next.ServeHTTP(w, r)
	})
}

// ------------------------------------------------------
//
//	middleware
//
// ------------------------------------------------------
func (app *application) RequireTokenAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token := r.Header.Get("Authorization")
		if token == "" {
			token = r.Header.Get("Authentication")

		}

		if token == "" {
			app.UnauthorizedErrorJSON(w, r)
			return
		}

		user, err := app.users.GetByToken(token)
		if err != nil {
			app.UnauthorizedErrorJSON(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), models.ContextUserKey, user.ID)
		// Otherwise set the "Cache-Control: no-store" header so that pages
		// require authentication are not stored in the users browser cache (or
		// other intermediary cache).
		w.Header().Add("Cache-Control", "no-store")
		// And call the next handler in the chain.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
