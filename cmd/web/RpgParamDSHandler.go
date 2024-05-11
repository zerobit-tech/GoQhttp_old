package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zerobit-tech/GoQhttp/internal/rpg"
	"github.com/zerobit-tech/GoQhttp/internal/validator"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RpgParamDSHandlers(router *chi.Mux) {
	router.Route("/pgmfieldsds", func(r chi.Router) {
		//r.With(paginate).Get("/", listArticles)
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(CheckLicMiddleware)

		r.Get("/add", app.rpgParamDSAdd)
		r.Post("/add", app.rpgParamDSAddPost)

		r.Get("/update/{id}", app.rpgParamDSUpdate)
		r.Post("/update/{id}", app.rpgParamDSAddPost)

		r.Get("/fieldrow", app.rpgParamDSAdd_Field_row)

	})

}

// ------------------------------------------------------

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) rpgParamDSAdd_Field_row(w http.ResponseWriter, r *http.Request) {

	currentindex := r.URL.Query().Get("index")

	currentindexI, err := strconv.Atoi(currentindex)
	if err != nil {
		currentindexI = 0
	}

	currentindexI += 1

	dsField := &rpg.DSField{
		ParamID:   "",
		NameToUse: "",
		Dim:       0,
	}

	data := app.newTemplateData(r)

	data.DsField = dsField
	data.Index = currentindexI

	data.RpgParams = app.RpgParamModel.List()

	app.render(w, r, http.StatusOK, "empty_rpg_param_ds_add_field_row.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) rpgParamDSAdd(w http.ResponseWriter, r *http.Request) {

	param := rpg.Param{}

	data := app.newTemplateData(r)

	param.Init()

	data.Index = len(param.DsFields) - 1

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

	rpgParam.IsDs = true
	rpgParam.CheckField(validator.NotBlank(rpgParam.Name), "name", "This field cannot be blank")

	if rpgParam.Valid() {
		rpgParam.Name = stringutils.RemoveSpecialChars(stringutils.RemoveMultipleSpaces(rpgParam.Name))
		rpgParam.CheckField(!app.RpgParamModel.DuplicateName(&rpgParam), "name", "Duplicate name")
	}
	// assign default server

	logBeforeImage := ""

	duplicateNameError := false
	// Check SP details from iBMI
	if rpgParam.Valid() {

		if rpgParam.ID != "" {
			orginalSp, err := app.RpgParamModel.Get(rpgParam.ID)
			if err == nil {
				logBeforeImage = orginalSp.LogImage()

			}
		}

		for _, f := range rpgParam.DsFields {
			if strings.TrimSpace(f.NameToUse) == "" && f.ParamID != "" {
				param, err := app.RpgParamModel.Get(f.ParamID)
				if err == nil {
					f.Param = param
				}

			}
		}

		rpgParam.AssignDSFieldNames()
		duplicateNameError = rpgParam.ValidateFields()

	}

	if !rpgParam.Valid() || duplicateNameError {
		data := app.newTemplateData(r)
		data.Form = rpgParam
		app.sessionManager.Put(r.Context(), "error", "Please fix error(s) and resubmit")
		data.Servers = app.servers.List()
		data.RpgParamDatatypes = make([]string, 0)
		for k := range rpg.DataTypeMap {
			data.RpgParamDatatypes = append(data.RpgParamDatatypes, k)
		}
		data.Index = len(rpgParam.DsFields) - 1
		data.RpgParams = app.RpgParamModel.List()

		sort.Strings(data.RpgParamDatatypes)
		app.render(w, r, http.StatusUnprocessableEntity, "rpg_param_ds_add.tmpl", data)
		return
	}

	logAction := "RPG Param Created"
	if rpgParam.ID != "" {
		logAction = "RPG Param Modified"
	}

	rpgParam.IsDs = true
	rpgParam.FilterOutInvalidParams()

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
	http.Redirect(w, r, fmt.Sprintf("/pgmfields/%s", id), http.StatusSeeOther)

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
	data.Index = len(rpgParam.DsFields) - 1

	app.render(w, r, http.StatusOK, "rpg_param_ds_add.tmpl", data)

}
