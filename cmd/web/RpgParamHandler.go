package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"sort"
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
func (app *application) RpgParamHandlers(router *chi.Mux) {
	router.Route("/pgmfields", func(r chi.Router) {
		//r.With(paginate).Get("/", listArticles)
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(CheckLicMiddleware)

		r.Get("/", app.rpgParamList) // list

		r.Get("/dtfields", app.dtfields) // list

		r.Get("/{id}", app.rpgPramView) // one single
		r.Get("/add", app.rpgParamAdd)
		r.Post("/add", app.rpgParamAddPost)

		r.Get("/update/{id}", app.rpgParamUpdate)
		r.Post("/update/{id}", app.rpgParamAddPost)

		//r.Get("/delete/{id}", app.SPDelete)
		//r.Post("/delete", app.SPDeleteConfirm)

		r.Get("/usageds/{id}", app.rpgParamUsageDS)
		r.Get("/usagepgm/{id}", app.rpgParamUsagePGM)

	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) rpgParamList(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)
	data.RpgParams = app.RpgParamModel.List()

	nextUrl := r.URL.Query().Get("next") //filters=["color", "price", "brand"]
	data.Next = nextUrl
	app.render(w, r, http.StatusOK, "rpg_param_list.tmpl", data)

}

// ------------------------------------------------------
// data type fields: length / decimap position
// ------------------------------------------------------
func (app *application) dtfields(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	datatype := r.URL.Query().Get("datatype")
	id := r.URL.Query().Get("id")
	param := &rpg.Param{}
	if id != "" {
		p, err := app.RpgParamModel.Get(id)
		if err == nil {
			param = p
		}
	}
	data.Form = param

	templateToUse := "empty_empty.tmpl"
	if rpg.DataTypeNeedLength(datatype) {
		templateToUse = "empty_rpg_param_length.tmpl"
	}

	if rpg.DataTypeNeedDecimalValue(datatype) {
		templateToUse = "empty_rpg_param_length_decimal.tmpl"
	}

	app.render(w, r, http.StatusOK, templateToUse, data)
}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) rpgPramView(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	id := chi.URLParam(r, "id")

	param, err := app.RpgParamModel.Get(id)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.RpgParam = param
	app.render(w, r, http.StatusOK, "rpg_param_view.tmpl", data)

}

// ------------------------------------------------------

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) rpgParamAdd(w http.ResponseWriter, r *http.Request) {

	param := rpg.Param{}

	data := app.newTemplateData(r)
	data.RpgParamDatatypes = make([]string, 0)
	for k := range rpg.DataTypeMap {
		data.RpgParamDatatypes = append(data.RpgParamDatatypes, k)
	}

	sort.Strings(data.RpgParamDatatypes)
	data.Form = param
	app.render(w, r, http.StatusOK, "rpg_param_add.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) rpgParamAddPost(w http.ResponseWriter, r *http.Request) {

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

	rpgParam.IsValid()

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
		app.render(w, r, http.StatusUnprocessableEntity, "rpg_param_add.tmpl", data)
		return
	}

	logAction := "RPG Param Created"
	if rpgParam.ID != "" {
		logAction = "RPG Param Modified"
	}

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
// Delete saved query
// ------------------------------------------------------
func (app *application) rpgParamDelete(w http.ResponseWriter, r *http.Request) {

	spId := chi.URLParam(r, "spId")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	data := app.newTemplateData(r)
	data.StoredProc = sP

	app.render(w, r, http.StatusOK, "rpg_param_delete.tmpl", data)

}

// ------------------------------------------------------
// Delete saved query confirm
// ------------------------------------------------------
func (app *application) rpgParamDeleteConfirm(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("001 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	spId := r.PostForm.Get("spId")
	app.invalidateEndPointCache()

	err = app.storedProcs.Delete(spId)
	if err != nil {

		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Delete error: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	go app.deleteSPData(spId) //goroutine
	app.sessionManager.Put(r.Context(), "flash", "Endpoint deleted sucessfully")

	go func() {
		defer concurrent.Recoverer("SPMODIFIED")
		userID, _ := app.getCurrentUserID(r)

		logEvent := GetSystemLogEvent(userID, "Endpoint Deleted", fmt.Sprintf("IP %s", r.RemoteAddr), false)
		logEvent.ImpactedEndpointId = spId
		app.SystemLoggerChan <- logEvent

	}()

	http.Redirect(w, r, "/sp", http.StatusSeeOther)

}

// ------------------------------------------------------
// add new server
// ------------------------------------------------------
func (app *application) rpgParamUpdate(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	rpgParam, err := app.RpgParamModel.Get(id)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Update error: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)

	data.Form = rpgParam
	data.RpgParam = rpgParam

	data.RpgParamDatatypes = make([]string, 0)
	for k := range rpg.DataTypeMap {
		data.RpgParamDatatypes = append(data.RpgParamDatatypes, k)
	}

	sort.Strings(data.RpgParamDatatypes)

	app.render(w, r, http.StatusOK, "rpg_param_add.tmpl", data)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) rpgParamValidator(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)

	spId := chi.URLParam(r, "spId")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.StoredProc = sP

	data.ParamRegexs = app.paramRegexModel.List()
	app.render(w, r, http.StatusOK, "rpg_param_validator.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) rpgParamSaveValidator(w http.ResponseWriter, r *http.Request) {

	spId := r.FormValue("id")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	//	app.storedProcs.Save(sP)
	maped := formToMap(r)

	changed := false
	for _, p := range sP.Parameters {
		validator, found := maped[strings.ToUpper(p.Name)]
		if found {
			validatorString, ok := validator.(string)
			if ok && p.ValidatorRegex != validatorString {

				p.ValidatorRegex = validatorString

				if p.ValidatorRegex != "" {
					err := p.CheckValidatorRegex(app.paramRegexModel.Map())
					if err != nil {
						app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
						app.goBack(w, r, http.StatusSeeOther)
						return
					}
				}

				changed = true

			}
		}
	}

	if changed {
		app.invalidateEndPointCache()
		app.storedProcs.Save(sP)
		app.sessionManager.Put(r.Context(), "flash", "Done")

	} else {
		app.sessionManager.Put(r.Context(), "flash", "Nothing changed")

	}

	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
// add new server
// ------------------------------------------------------
func (app *application) rpgParamUsageDS(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	rpgParam, err := app.RpgParamModel.Get(id)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Update error: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)

	data.RpgParams = make([]*rpg.Param, 0)
	rpgParams := app.RpgParamModel.List()

	for _, p := range rpgParams {
		if p.DsHasField(rpgParam.ID) {
			data.RpgParams = append(data.RpgParams, p)
		}
	}

	nextUrl := r.URL.Query().Get("next") //filters=["color", "price", "brand"]
	data.Next = nextUrl
	app.render(w, r, http.StatusOK, "rpg_param_list.tmpl", data)

}

// ------------------------------------------------------
// add new server
// ------------------------------------------------------
func (app *application) rpgParamUsagePGM(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	rpgParam, err := app.RpgParamModel.Get(id)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Update error: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)

	data.RpgEndPoints = make([]*rpg.RpgEndPoint, 0, 10)

	storedPs := app.RpgEndpointModel.List()

	for _, sp := range storedPs {
		if sp.IsUsingField(app.RpgParamModel, rpgParam.ID) {
			data.RpgEndPoints = append(data.RpgEndPoints, sp)
		}
	}

	app.render(w, r, http.StatusOK, "rpg_endpoint_list.tmpl", data)

}
