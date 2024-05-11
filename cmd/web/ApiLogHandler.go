package main

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zerobit-tech/GoQhttp/internal/models"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) APILogHandlers(router *chi.Mux) {

	router.Route("/apilogs", func(r chi.Router) {
		// CSRF
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(noSurf)
		r.Use(CheckLicMiddleware)
		r.Get("/", app.apilogs)
		r.Get("/{logid}", app.apilogs)
		r.Post("/", app.apilogs)

		logGroup := r.Group(nil)
		logGroup.Use(app.RequireSuperAdmin)
		logGroup.Get("/clear", app.clearapilogs)

	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) apilogs(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	objectid := strings.TrimSpace(r.PostForm.Get("objectid"))
	logid := chi.URLParam(r, "logid")

	if objectid == "" {
		objectid = logid
	}
	logEntries := make([]string, 0)
	if objectid != "" {
		logEntries = models.GetLogs(app.LogDB, objectid)
	}

	data := app.newTemplateData(r)
	data.LogEntries = logEntries

	app.render(w, r, http.StatusOK, "api_logs.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------

func (app *application) clearapilogs(w http.ResponseWriter, r *http.Request) {
	//models.ClearLogs(app.LogDB) // TODO
	app.sessionManager.Put(r.Context(), "flash", "Api logs has been cleared")

	app.goBack(w, r, http.StatusSeeOther)
}
