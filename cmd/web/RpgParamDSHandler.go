package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/onlysumitg/GoQhttp/internal/rpg"
	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/utils/concurrent"
	"github.com/onlysumitg/GoQhttp/utils/stringutils"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RpgParamDSHandlers(router *chi.Mux) {
	router.Route("/rpgparamds", func(r chi.Router) {
		//r.With(paginate).Get("/", listArticles)
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(CheckLicMiddleware)

		r.Get("/add", app.rpgParamDSAdd)
		r.Post("/add", app.rpgParamDSAddPost)

		r.Get("/update/{id}", app.rpgParamDSUpdate)
		r.Post("/update/{id}", app.rpgParamDSAddPost)

	})

}

// ------------------------------------------------------

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) rpgParamDSAdd(w http.ResponseWriter, r *http.Request) {

	param := rpg.Param{}

	data := app.newTemplateData(r)

	param.DsFields = make([]string, 20)
	data.Form = param

	data.RpgParams = app.RpgParamModel.List()

	app.render(w, r, http.StatusOK, "rpg_param_ds_add.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) rpgParamDSAddPost(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("001 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	var rpgParam rpg.Param
	err = app.formDecoder.Decode(&rpgParam, r.PostForm)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("002 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}
	rpgParam.CheckField(validator.NotBlank(rpgParam.Name), "name", "This field cannot be blank")

	if rpgParam.Valid() {
		rpgParam.Name = stringutils.RemoveSpecialChars(stringutils.RemoveMultipleSpaces(rpgParam.Name))
		rpgParam.CheckField(!app.RpgParamModel.DuplicateName(&rpgParam), "name", "Duplicate name")
	}
	// assign default server

	logBeforeImage := ""

	// Check SP details from iBMI
	if rpgParam.Valid() {

		if rpgParam.ID != "" {
			orginalSp, err := app.RpgParamModel.Get(rpgParam.ID)
			if err == nil {
				logBeforeImage = orginalSp.LogImage()

			}
		}

	}

	if !rpgParam.Valid() {
		data := app.newTemplateData(r)
		data.Form = rpgParam
		app.sessionManager.Put(r.Context(), "error", "Please fix error(s) and resubmit")
		data.Servers = app.servers.List()
		data.RpgParamDatatypes = make([]string, 0)
		for k := range rpg.DataTypeMap {
			data.RpgParamDatatypes = append(data.RpgParamDatatypes, k)
		}

		sort.Strings(data.RpgParamDatatypes)
		app.render(w, r, http.StatusUnprocessableEntity, "rpg_param_ds_add.tmpl", data)
		return
	}

	logAction := "RPG Param Created"
	if rpgParam.ID != "" {
		logAction = "RPG Param Modified"
	}

	rpgParam.IsDs = true
	// finally save
	id, err := app.RpgParamModel.Save(&rpgParam)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", fmt.Sprintf("Param %s added sucessfully", rpgParam.Name))

	go func() {
		defer concurrent.Recoverer("RPGPARAMCREATED")
		defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
		userID, _ := app.getCurrentUserID(r)

		logEvent := GetSystemLogEvent(userID, logAction, fmt.Sprintf(" %s,IP %s", rpgParam.Name, r.RemoteAddr), false)
		logEvent.ImpactedEndpointId = id
		logEvent.BeforeUpdate = logBeforeImage
		logEvent.AfterUpdate = rpgParam.LogImage()
		app.SystemLoggerChan <- logEvent

	}()

	//http.Redirect(w, r, fmt.Sprintf("/savesql/%s", id), http.StatusSeeOther)
	http.Redirect(w, r, fmt.Sprintf("/rpgparam/%s", id), http.StatusSeeOther)

}

// ------------------------------------------------------
// add new server
// ------------------------------------------------------
func (app *application) rpgParamDSUpdate(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	rpgParam, err := app.RpgParamModel.Get(id)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Update error: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)

	data.Form = rpgParam
	data.RpgParams = app.RpgParamModel.List()

	app.render(w, r, http.StatusOK, "rpg_param_ds_add.tmpl", data)

}
