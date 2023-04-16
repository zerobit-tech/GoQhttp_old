package main

import (
	"net/http"
)

func (app *application) langingPage(w http.ResponseWriter, r *http.Request) {
	//http.Redirect(w, r, "/endpoints", http.StatusSeeOther)

	data := app.newTemplateData(r)

	app.render(w, r, http.StatusOK, "public_index.tmpl", data)

}

func (app *application) appLangingPage() string {
	return "/sp"

}
func (app *application) helpPage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	app.render(w, r, http.StatusOK, "help.tmpl", data)

}

func (app *application) testModePage(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	app.render(w, r, http.StatusOK, "testmode.tmpl", data)

}
