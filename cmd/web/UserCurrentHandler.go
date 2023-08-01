package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) CurrentUsersHandlers(router *chi.Mux) {
	router.Route("/currentuser", func(r chi.Router) {
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(CheckLicMiddleware)

		// CSRF
		r.Use(noSurf)

		r.Get("/", app.CurrentUserView)

	})

}

// ------------------------------------------------------
// Delete servet
// ------------------------------------------------------
func (app *application) CurrentUserView(w http.ResponseWriter, r *http.Request) {

	user, err := app.GetUser(r)
	if err != nil {
		app.clientError(w, http.StatusNotFound, err)
		return
	}

	data := app.newTemplateData(r)
	data.User = user

	app.render(w, r, http.StatusOK, "user_view.tmpl", data)

}
