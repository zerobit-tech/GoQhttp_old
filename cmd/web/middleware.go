package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"sync/atomic"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
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
		//u := r.URL
		//log.Println("starte", u.String(), "::", r.URL.Scheme, r.TLS, r.Host, r.RequestURI, "::", r.Header.Get(xForwardedProtoHeader))
		if r.Header.Get(xForwardedProtoHeader) != "https" {

			//log.Println(":::::::: REDIRECTING :::::::::")
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
		response := &models.StoredProcResponse{ReferenceId: middleware.GetReqID(r.Context())}

		token := r.Header.Get("Authorization")
		if token == "" {
			token = r.Header.Get("Authentication")

		}

		if token == "" {
			response.Status = http.StatusUnauthorized
			response.Message = http.StatusText(http.StatusUnauthorized)
			app.writeJSONAPI(w, response, nil)
			return

		}

		user, err := app.users.GetByToken(token)
		if err != nil {
			response.Status = http.StatusUnauthorized
			response.Message = http.StatusText(http.StatusUnauthorized)
			app.writeJSONAPI(w, response, nil)
			return
		}

		ctx := context.WithValue(r.Context(), models.ContextUserKey, user.ID)
		ctx = context.WithValue(ctx, models.ContextUserName, user.Name)

		// Otherwise set the "Cache-Control: no-store" header so that pages
		// require authentication are not stored in the users browser cache (or
		// other intermediary cache).
		w.Header().Add("Cache-Control", "no-store")
		// And call the next handler in the chain.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

//	------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) LogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := middleware.GetReqID(r.Context())

		requestBody := ""
		x, err := httputil.DumpRequest(r, true)
		if err == nil {
			requestBody = string(x)
		} else {
			requestBody = "Error :" + err.Error()
		}
		go func() {

			buf := bytes.NewBufferString("")

			models.RequestLog.SetOutput(buf)
			models.RequestLog.Println("\n\n" + requestBody)

			models.SaveLogs(app.LogDB, 998, requestId, buf.String(), app.testMode)
		}()

		rec := httptest.NewRecorder()

		next.ServeHTTP(rec, r)

		// After processing ==> log response
		responseBody := ""
		y, err := httputil.DumpResponse(rec.Result(), true)
		if err == nil {

			responseBody = string(y)
			go func() {

				buf := bytes.NewBufferString("")

				models.ResponseLog.SetOutput(buf)
				models.ResponseLog.Println("\n\n" + responseBody)

				models.SaveLogs(app.LogDB, 999, requestId, buf.String(), app.testMode)
			}()
		}

		// this copies the recorded response to the response writer
		for k, v := range rec.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(rec.Code)
		rec.Body.WriteTo(w)

		if rec.Code >= 400 {
			go func() {
				email := &models.EmailRequest{
					Subject:  fmt.Sprintf("%d %s", rec.Code, requestId),
					Body:     fmt.Sprintf("<h3>Request</h3><br><pre>%s</pre><br><br><br><br><h3>Response</h3><br><pre>%s</pre>", requestBody, responseBody),
					Template: "",
					Data:     nil,
				}

				app.SendNotificationsToAdmins(email)
			}()
		}
	})
}

var reqid uint64
var prefix string = uuid.NewString()

func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get(middleware.RequestIDHeader)
		if requestID == "" {
			myid := atomic.AddUint64(&reqid, 1)
			requestID = fmt.Sprintf("%s-%06d", prefix, myid)
		}
		ctx = context.WithValue(ctx, middleware.RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
