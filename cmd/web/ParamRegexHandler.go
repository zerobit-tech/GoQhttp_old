package main

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/zerobit-tech/GoQhttp/internal/models"
	"github.com/zerobit-tech/GoQhttp/internal/validator"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) ParamRegexHandlers(router *chi.Mux) {
	router.Route("/pramregex", func(r chi.Router) {
		r.Use(app.sessionManager.LoadAndSave)
		r.Use(app.RequireAuthentication)
		r.Use(app.RequireSuperAdmin)
		r.Use(CheckLicMiddleware)

		// CSRF
		r.Use(noSurf)
		r.Get("/", app.paramRegexList)
		r.Get("/add", app.paramRegexAdd)
		r.Post("/add", app.paramRegexAdd)
		r.Get("/edit/{id}", app.paramRegexAdd)
		r.Post("/edit/{id}", app.paramRegexAdd)

		r.Get("/delete/{id}", app.paramRegexDelete)
		r.Post("/delete", app.paramRegexDeleteConfirm)

	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) paramRegexList(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)
	data.ParamRegexs = app.paramRegexModel.List()
	app.render(w, r, http.StatusOK, "param_regex_list.tmpl", data)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) paramRegexAdd(w http.ResponseWriter, r *http.Request) {

	paramR := &models.ParamRegex{}

	id := chi.URLParam(r, "id")
	if id != "" {
		u, err := app.paramRegexModel.Get(id)
		if err == nil {
			paramR = u

		}
	}

	if r.Method == http.MethodPost {

		err := app.decodePostForm(r, &paramR)
		if err != nil {
			app.clientError(w, http.StatusBadRequest, err)
			return
		}

		paramR.CheckField(validator.NotBlank(paramR.Name), "name", "This field cannot be blank")

		paramR.CheckField(validator.NotBlank(paramR.Regex), "regex", "This field cannot be blank")

		if paramR.Valid() {
			_, err := regexp.Compile(paramR.Regex)
			if err != nil {
				paramR.AddFieldError("regex", err.Error())
			}
		}

		if paramR.Valid() {
			app.paramRegexModel.Save(paramR)
			app.sessionManager.Put(r.Context(), "flash", "Saved sucessfully")

			http.Redirect(w, r, "/pramregex", http.StatusSeeOther)
			return
		}

	}

	data := app.newTemplateData(r)
	data.Form = paramR

	app.render(w, r, http.StatusOK, "param_regex_add.tmpl", data)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) paramRegexDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	pr, err := app.paramRegexModel.Get(id)
	if err != nil {
		app.clientError(w, http.StatusNotFound, err)
		return
	}

	data := app.newTemplateData(r)
	data.ParamRegex = pr

	app.render(w, r, http.StatusOK, "param_regex_delete.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) paramRegexDeleteConfirm(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("001 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	id := r.PostForm.Get("id")

	err = app.paramRegexModel.Delete(id)
	if err != nil {

		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("delete failed:: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Deleted sucessfully")

	http.Redirect(w, r, "/pramregex", http.StatusSeeOther)

}
