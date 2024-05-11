package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zerobit-tech/GoQhttp/internal/models"
	"github.com/zerobit-tech/GoQhttp/utils/templateutil"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SPtemplateHandler(router *chi.Mux) {

	router.Route("/t", func(r chi.Router) {
		// CSRF
		r.Use(app.sessionManager.LoadAndSave)
		r.Use(app.RequireAuthenticationForTemplatedAPI)
		r.Use(noSurf)
		r.Use(CheckLicMiddleware)
		r.Get("/render/{name}", app.renderSPTemplate)

		// logGroup := r.Group(nil)
		// logGroup.Use(app.RequireSuperAdmin)
		// logGroup.Get("/clear", app.clearapilogs)

	})

}

func (app *application) renderSPTemplate(w http.ResponseWriter, r *http.Request) {

	templateName := fmt.Sprintf("%s.html", chi.URLParam(r, "name"))

	app.spRender(w, r, http.StatusOK, templateName, nil, nil)

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) LoadSPTemplates() {
	tCache, err := templateutil.NewTemplateCache()

	defer app.cacheMutext.Unlock()
	app.cacheMutext.Lock()

	app.storedProcsTemplateCache = tCache

	app.storedProcsTemplates = make([]string, 0)
	if err != nil {
		log.Println("ERROR::", err)
		return
	}

	for k := range tCache {
		app.storedProcsTemplates = append(app.storedProcsTemplates, k)
	}

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) spRender(w http.ResponseWriter, r *http.Request, status int, page string, data any, headers http.Header) {

	// Write the template to the buffer, instead of straight to the
	// http.ResponseWriter. If there's an error, call our serverError() helper
	// and then return.
	app.cacheMutext.RLock()
	buf, err := templateutil.TemplateToBuffer(app.storedProcsTemplateCache, page, data)
	app.cacheMutext.RUnlock()
	if err != nil {
		app.serverError500(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	for key, value := range headers {
		w.Header()[key] = value
	}

	// If the template is written to the buffer without any errors, we are safe
	// to go ahead and write the HTTP status code to http.ResponseWriter.
	w.WriteHeader(status)
	// Write the contents of the buffer to the http.ResponseWriter. Note: this
	// is another time where we pass our http.ResponseWriter to a function that
	// takes an io.Writer.

	buf.WriteTo(w)
}

// ------------------------------------------------------
func (app *application) RequireAuthenticationForTemplatedAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the user is not authenticated, redirect them to the login page and
		// return from the middleware chain so that no subsequent handlers in
		// the chain are executed.
		goToUrl := fmt.Sprintf("/user/login?next=%s", r.URL.RequestURI())

		if !app.isAuthenticated(r, false) {
			app.sessionManager.Put(r.Context(), "error", "Login required")

			http.Redirect(w, r, goToUrl, http.StatusSeeOther)
			return
		}

		user, err := app.GetUser(r)
		if !user.HasVerified {
			app.sessionManager.Put(r.Context(), "error", "Please verify your email")

			http.Redirect(w, r, goToUrl, http.StatusSeeOther)
			return
		}

		userId := ""
		if err == nil {
			userId = user.ID
		}
		ctx := context.WithValue(r.Context(), models.ContextUserKey, userId)
		// Otherwise set the "Cache-Control: no-store" header so that pages
		// require authentication are not stored in the users browser cache (or
		// other intermediary cache).
		w.Header().Add("Cache-Control", "no-store")
		// And call the next handler in the chain.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
