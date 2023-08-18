package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) DocHandlers(router *chi.Mux) {

	router.Route("/docs", func(r chi.Router) {
		// CSRF
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(noSurf)
		r.Use(CheckLicMiddleware)
		r.Get("/", app.docList)

		r.Get("/envvar", app.docEnvVar)
		r.Get("/faq", app.docFAQ)
		r.Get("/prtable", app.docPromotionTable)
		r.Get("/uttable", app.docUserTokenTable)
		r.Get("/datetime", app.docDateTimeFormat)

	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) docList(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)

	app.render(w, r, http.StatusOK, "doc_list.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) docEnvVar(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)

	app.render(w, r, http.StatusOK, "doc_envvar.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) docFAQ(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)

	app.render(w, r, http.StatusOK, "doc_faq.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) docPromotionTable(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)

	app.render(w, r, http.StatusOK, "doc_promotion_table.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) docUserTokenTable(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)

	app.render(w, r, http.StatusOK, "doc_user_token_table.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) docDateTimeFormat(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)

	app.render(w, r, http.StatusOK, "doc_date_time_formats.tmpl", data)

}
