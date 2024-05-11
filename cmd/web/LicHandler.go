package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zerobit-tech/GoQhttp/lic"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) LicHandlers(router *chi.Mux) {

	router.Route("/license", func(r chi.Router) {
		// CSRF
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(CheckLicMiddlewareNoRedirect)
		r.Use(noSurf)
		r.Get("/", app.LicList)

	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) LicList(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)
	data.LicenseEntries = lic.GetLicFileWithStatus()

	app.render(w, r, http.StatusOK, "lic.tmpl", data)

}
