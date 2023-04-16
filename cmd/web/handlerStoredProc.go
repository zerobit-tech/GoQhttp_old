package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/onlysumitg/GoQhttp/internal/models"
	"github.com/onlysumitg/GoQhttp/internal/validator"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) StoredProcHandlers(router *chi.Mux) {
	router.Route("/sp", func(r chi.Router) {
		r.Use(app.RequireAuthentication)
		r.Use(app.CurrentServerMiddleware)
		//r.With(paginate).Get("/", listArticles)
		r.Get("/", app.SPList)
		r.Post("/call", app.SPCall)
		r.Get("/{spId}", app.SPView)

		r.Get("/add", app.SPAdd)
		r.Post("/add", app.SPAddPost)

		r.Get("/update/{spId}", app.SPUpdate)

		r.Get("/delete/{spId}", app.SPDelete)
		r.Post("/delete", app.SPDeleteConfirm)

		r.Get("/run/{spId}", app.SPRun)
		r.Get("/run", app.SPRun)

		r.Post("/build", app.SPBuild)

	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SPList(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.StoredProcs = app.storedProcs.List()
	nextUrl := r.URL.Query().Get("next") //filters=["color", "price", "brand"]
	data.Next = nextUrl
	app.render(w, r, http.StatusOK, "sp_list.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SPRun(w http.ResponseWriter, r *http.Request) {
	// data := app.newTemplateData(r)

	// savesQueries := app.savedQueries.List()
	// data.SavesQueries = savesQueries
	// data.SavesQueriesByCategory = make(map[string][]*models.StoredProc)

	// //spId := chi.URLParam(r, "spId")

	// for _, savesQuery := range savesQueries {
	// 	savesQuery.PopulateFields()

	// 	queryList, found := data.SavesQueriesByCategory[savesQuery.Category]
	// 	if !found {
	// 		queryList = make([]*models.SP, 0)
	// 	}
	// 	queryList = append(queryList, savesQuery)
	// 	data.SavesQueriesByCategory[savesQuery.Category] = queryList

	// }

	// nextUrl := r.URL.Query().Get("next") //filters=["color", "price", "brand"]
	// data.Next = nextUrl
	// app.render(w, r, http.StatusOK, "sp_run.tmpl", data)

}

// ------------------------------------------------------
func (app *application) SPBuild(w http.ResponseWriter, r *http.Request) {

	// formMap := map[string]string{}
	// err := json.NewDecoder(r.Body).Decode(&formMap)
	// if err != nil {
	// 	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
	// 	return
	// }
	// log.Println("><<>>>>>>", formMap)
	// savedQueeryId, found := formMap["sPid"]
	// if savedQueeryId == "" || !found {
	// 	app.serverError(w, r, errors.New("sPid is required"))
	// 	return
	// }
	// sP, err := app.savedQueries.Get(savedQueeryId)
	// log.Println("sP>>>", sP, err)
	// if err != nil {
	// 	app.serverError(w, r, err)
	// 	return
	// }

	// sqlToRun, fieldError := sP.ReplaceFields(formMap)
	// if len(fieldError) > 0 {
	// 	// if has error field -> return blank sql to run
	// 	sqlToRun = ""

	// }

	// sPBuild := models.SPBuild{SqlToRun: sqlToRun, FieldErrors: fieldError}

	// app.writeJSON(w, http.StatusOK, sPBuild, nil)

	// // need to return a json

}

// ------------------------------------------------------
func (app *application) SPRunAsJson(w http.ResponseWriter, r *http.Request) {

	// currentServerID := app.sessionManager.GetString(r.Context(), "currentserver")
	// currentServer, err := app.servers.Get(currentServerID)
	// if err != nil {
	// 	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
	// 	return
	// }

	// if err := r.ParseForm(); err != nil {
	// 	// handle error
	// }
	// savedQueeryId := r.PostForm.Get("sPid")
	// if savedQueeryId != "" {
	// 	app.serverError500(w, r, errors.New("sPid is required"))
	// 	return
	// }
	// sP, err := app.savedQueries.Get(savedQueeryId)
	// if err != nil {
	// 	app.serverError(w, r, err)
	// 	return
	// }

	// fieldMap := make(map[string]string)
	// for key, values := range r.PostForm {
	// 	fieldMap[key] = values[0]
	// }

	// sqlToRun, fieldError := sP.ReplaceFields(fieldMap)
	// if len(fieldError) == 0 {
	// 	// No error
	// 	// run the sql
	// }

	// sessionID := app.sessionManager.Token(r.Context())

	// currentTabId, lastTabid := getTabIds(r)

	// queryResults := models.ProcessSQLStatements(sqlToRun, currentServer, sessionID, currentTabId, lastTabid)
	// app.writeJSON(w, http.StatusOK, queryResults, nil)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SPAdd(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = models.StoredProc{}
	app.render(w, r, http.StatusOK, "sp_add.tmpl", data)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) SPView(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	spId := chi.URLParam(r, "spId")
	log.Println("spId >>>", spId, data.CSRFToken)
	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.StoredProc = sP
	app.render(w, r, http.StatusOK, "sp_view.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SPAddPost(w http.ResponseWriter, r *http.Request) {
	currentServerID := app.sessionManager.GetString(r.Context(), "currentserver")
	currentServer, err := app.servers.Get(currentServerID)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	err = r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("001 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	var sP models.StoredProc
	err = app.formDecoder.Decode(&sP, r.PostForm)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("002 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	sP.CheckField(validator.NotBlank(sP.Name), "name", "This field cannot be blank")
	sP.CheckField(validator.NotBlank(sP.Lib), "lib", "This field cannot be blank")

	if sP.Valid() {
		err = sP.PreapreToSave(*currentServer)

		if err != nil {
			sP.CheckField(false, "name", err.Error())
		}

	}

	if !sP.Valid() {
		data := app.newTemplateData(r)
		data.Form = sP
		app.sessionManager.Put(r.Context(), "error", "Please fix error(s) and resubmit")

		app.render(w, r, http.StatusUnprocessableEntity, "sp_add.tmpl", data)
		return
	}

	_, err = app.storedProcs.Save(&sP)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}

	id, err := app.storedProcs.Save(&sP)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	fmt.Println("id", id)
	app.sessionManager.Put(r.Context(), "flash", fmt.Sprintf("SP %s added sucessfully", sP.Name))

	//http.Redirect(w, r, fmt.Sprintf("/savesql/%s", id), http.StatusSeeOther)
	http.Redirect(w, r, fmt.Sprintf("/sp/%s", id), http.StatusSeeOther)

}

// ------------------------------------------------------
// Delete saved query
// ------------------------------------------------------
func (app *application) SPDelete(w http.ResponseWriter, r *http.Request) {

	spId := chi.URLParam(r, "spId")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error deleting query: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	data := app.newTemplateData(r)
	data.StoredProc = sP

	app.render(w, r, http.StatusOK, "sp_delete.tmpl", data)

}

// ------------------------------------------------------
// Delete saved query confirm
// ------------------------------------------------------
func (app *application) SPDeleteConfirm(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("001 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	spId := r.PostForm.Get("spId")

	err = app.storedProcs.Delete(spId)
	if err != nil {

		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error deleting query: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Query deleted sucessfully")

	http.Redirect(w, r, "/sp", http.StatusSeeOther)

}

// ------------------------------------------------------
// add new server
// ------------------------------------------------------
func (app *application) SPUpdate(w http.ResponseWriter, r *http.Request) {

	spId := chi.URLParam(r, "spId")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error updating server: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)

	data.Form = sP

	app.render(w, r, http.StatusOK, "sp_add.tmpl", data)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) SPCall(w http.ResponseWriter, r *http.Request) {

	currentServerID := app.sessionManager.GetString(r.Context(), "currentserver")
	currentServer, err := app.servers.Get(currentServerID)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	spId := r.FormValue("id")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	_, err = sP.DummyCall(*currentServer, formToMap(r))
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error call Stored proc: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	app.storedProcs.Save(sP)
	app.goBack(w, r, http.StatusSeeOther)

}

func formToMap(r *http.Request) map[string]any {
	formMap := make(map[string]any)
	err := r.ParseForm()
	if err == nil {

		for key := range r.PostForm {

			formMap[strings.ToUpper(key)] = r.FormValue(key)
		}

	}

	return formMap

}
