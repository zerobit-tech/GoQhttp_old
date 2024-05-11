package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SystemHandler(router *chi.Mux) {

	router.Route("/sys", func(r chi.Router) {
		// CSRF
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(noSurf)
		r.Use(CheckLicMiddleware)
		r.Get("/reloadtemplates", app.reloadSpTemplates)
		r.Get("/invalidatecachee", app.invalidatecache)

		// logGroup := r.Group(nil)
		// logGroup.Use(app.RequireSuperAdmin)
		// logGroup.Get("/clear", app.clearapilogs)

	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------

func (app *application) reloadSpTemplates(w http.ResponseWriter, r *http.Request) {

	app.LoadSPTemplates()
	app.sessionManager.Put(r.Context(), "flash", "Done")
	app.goBack(w, r, http.StatusSeeOther)
}

// ------------------------------------------------------
//
// ------------------------------------------------------

func (app *application) invalidatecache(w http.ResponseWriter, r *http.Request) {
	app.LoadSPTemplates()
	app.invalidateEndPointCache()

	app.deleteRPGDrivers()
	app.createRPGDrivers()

	app.sessionManager.Put(r.Context(), "flash", "Done")
	app.goBack(w, r, http.StatusSeeOther)
}
