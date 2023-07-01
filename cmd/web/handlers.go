package main

import (
	"net/http"
)

func (app *application) langingPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)

	// data := app.newTemplateData(r)

	// app.render(w, r, http.StatusOK, "public_index.tmpl", data)

}

func (app *application) appLangingPage() string {
	return "/dashboard"

}
 
 