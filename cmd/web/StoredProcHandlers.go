package main

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
	"github.com/zerobit-tech/GoQhttp/internal/validator"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) StoredProcHandlers(router *chi.Mux) {
	router.Route("/sp", func(r chi.Router) {
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(CheckLicMiddleware)

		//r.Use(app.CurrentServerMiddleware)
		//r.With(paginate).Get("/", listArticles)
		r.Get("/", app.SPList)
		r.Get("/find", app.FindSP)
		r.Post("/call", app.SPCall)
		r.Get("/{spId}", app.SPView)

		r.Get("/add", app.SPAdd)
		r.Post("/add", app.SPAddPost)

		r.Get("/update/{spId}", app.SPUpdate)
		r.Post("/update/{spId}", app.SPAddPost)

		r.Get("/delete/{spId}", app.SPDelete)
		r.Post("/delete", app.SPDeleteConfirm)

		// r.Get("/run/{spId}", app.SPRun)
		// r.Get("/run", app.SPRun)

		// r.Post("/build", app.SPBuild)

		r.Get("/refresh/{spId}", app.SpRefresh)

		r.Post("/assignserver", app.AssignServer)
		r.Post("/deleteassignserver", app.RemoveAssignServer)

		r.Get("/logs/{spId}", app.SpLogs)

		r.Get("/paramalias/{spId}", app.SpParamAlias)
		r.Post("/saveparamalias", app.SPsaveparamalias)

		r.Get("/paramplacement/{spId}", app.SpParamPos)
		r.Post("/saveparamplacement", app.SpSaveParamPos)

		r.Get("/paramvalid/{spId}", app.spParamValidator)
		r.Post("/saveparamvalid", app.spsaveparamValidator)

		r.Get("/help", app.SPHelp)

	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SPHelp(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	app.render(w, r, http.StatusOK, "sp_help_inbuilt_param.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SPList(w http.ResponseWriter, r *http.Request) {

	serverID := r.URL.Query().Get("server")
	loadSpecialparam := r.URL.Query().Get("loadspecial")

	loadSpecial := false
	if loadSpecialparam == "Y" {
		loadSpecial = true
	}

	_, err := app.servers.Get(serverID)
	if err != nil {
		serverID = ""

	}

	data := app.newTemplateData(r)
	data.StoredProcs = make([]*storedProc.StoredProc, 0, 10)

	storedPs := app.storedProcs.List(loadSpecial)

	if serverID != "" {
		for _, s := range storedPs {
			if s == nil || s.DefaultServer == nil {
				continue
			}
			allowed := false
			if s.DefaultServer.ID == serverID {
				allowed = true
			} else {

				for _, als := range s.AllowedOnServers {
					if als.ID == serverID {
						allowed = true
					}

				}
			}
			if allowed {
				data.StoredProcs = append(data.StoredProcs, s)
			}
		}
	} else {
		data.StoredProcs = storedPs
	}

	nextUrl := r.URL.Query().Get("next") //filters=["color", "price", "brand"]
	data.Next = nextUrl
	app.render(w, r, http.StatusOK, "sp_list.tmpl", data)

}

// ------------------------------------------------------

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SPAdd(w http.ResponseWriter, r *http.Request) {

	storedP := storedProc.StoredProc{}

	serverID := r.URL.Query().Get("server")

	server, err := app.servers.Get(serverID)
	if err == nil {
		storedP.DefaultServer = &storedProc.ServerRecord{ID: server.ID}

	}

	libName := r.URL.Query().Get("lib")
	storedP.Lib = libName
	spName := r.URL.Query().Get("sp")
	storedP.Name = spName

	isSPecific := r.URL.Query().Get("spcific")

	if strings.EqualFold(isSPecific, "Y") {

		storedP.UseSpecificName = true
	}

	data := app.newTemplateData(r)
	data.Form = storedP
	data.Servers = app.servers.List()
	app.render(w, r, http.StatusOK, "sp_add.tmpl", data)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) SpRefresh(w http.ResponseWriter, r *http.Request) {

	spId := chi.URLParam(r, "spId")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}

	// check default server
	if sP.DefaultServer != nil && sP.DefaultServer.ID != "" {
		dServer, err := app.servers.Get(sP.DefaultServer.ID)
		if err == nil {

			err = dServer.PrepareToSave(r.Context(), sP)
			if err == nil {
				app.storedProcs.Save(sP)
				app.sessionManager.Put(r.Context(), "flash", "Done")
			}
		}
	} else {
		err = errors.New("Default Server is not defined.")
	}

	if err != nil {
		app.sessionManager.Put(r.Context(), "error", err.Error())

	}
	app.invalidateEndPointCache()

	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) SPView(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	spId := chi.URLParam(r, "spId")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.StoredProc = sP
	data.Servers = app.servers.List()
	//data.SPCallLog, _ = app.spCallLogModel.Get(sP.ID)
	app.render(w, r, http.StatusOK, "sp_view.tmpl", data)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) SpLogs(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	spId := chi.URLParam(r, "spId")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.StoredProc = sP
	data.SPCallLog, _ = app.spCallLogModel.Get(sP.ID)
	app.render(w, r, http.StatusOK, "sp_logs.tmpl", data)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) SpParamAlias(w http.ResponseWriter, r *http.Request) {

	if !app.features.AllowParameterAlias {
		//app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusNotFound)
		return
	}

	data := app.newTemplateData(r)

	spId := chi.URLParam(r, "spId")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", err)
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	if sP.IsSpecial {
		app.sessionManager.Put(r.Context(), "error", "Not allowed")
		app.goBack(w, r, http.StatusBadRequest)
		return
	}
	data.StoredProc = sP
	app.render(w, r, http.StatusOK, "sp_param_alias.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SPAddPost(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("001 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	var sP storedProc.StoredProc
	err = app.formDecoder.Decode(&sP, r.PostForm)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("002 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}
	sP.CheckField(validator.NotBlank(sP.EndPointName), "endpointname", "This field cannot be blank")

	sP.CheckField(validator.NotBlank(sP.Name), "name", "This field cannot be blank")
	sP.CheckField(validator.NotBlank(sP.Lib), "lib", "This field cannot be blank")
	//sP.CheckField(!app.storedProcs.Duplicate(&sP), "endpointname", "Endpoint with name and method already exists")
	sP.CheckField(validator.NotBlank(sP.DefaultServerId), "serverid", "This field cannot be blank")

	sP.SetNameSpace()

	if sP.Valid() {

		sP.EndPointName = stringutils.RemoveSpecialChars(stringutils.RemoveMultipleSpaces(sP.EndPointName))

		sP.CheckField(!app.storedProcs.Duplicate(&sP), "endpointname", "Endpoint with name and method already exists in this namespace.")
		sP.CheckField(!app.RpgEndpointModel.DuplicateByName(sP.EndPointName, sP.HttpMethod, sP.Namespace), "endpointname", "Endpoint with name and method already exists in this namespace.")

	}
	// assign default server

	server, err := app.servers.Get(sP.DefaultServerId)
	if err != nil {
		sP.CheckField(false, "serverid", "Server not found")

		sP.Validator.AddNonFieldError("Please select a valid server")
	} else {
		srcd := &storedProc.ServerRecord{ID: server.ID, Name: server.Name}
		sP.DefaultServer = srcd
		sP.AddAllowedServer(server.ID, server.Name)
	}

	logBeforeImage := ""

	// Check SP details from iBMI
	if sP.Valid() {

		if sP.ID != "" {
			orginalSp, err := app.storedProcs.Get(sP.ID)
			if err == nil {
				logBeforeImage = orginalSp.LogImage()
				sP.Parameters = orginalSp.Parameters

			}
		}

		err = server.PrepareToSave(r.Context(), &sP)

		if err != nil {
			sP.CheckField(false, "name", err.Error())
		}

	}

	if !sP.Valid() {
		data := app.newTemplateData(r)
		data.Form = sP
		app.sessionManager.Put(r.Context(), "error", "Please fix error(s) and resubmit")
		data.Servers = app.servers.List()

		app.render(w, r, http.StatusUnprocessableEntity, "sp_add.tmpl", data)
		return
	}

	logAction := "Endpoint Created"
	if sP.ID != "" {
		logAction = "Endpoint Modified"
	}

	// finally save
	id, err := app.storedProcs.Save(&sP)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	app.invalidateEndPointCache()
	app.sessionManager.Put(r.Context(), "flash", fmt.Sprintf("Endpoint %s added sucessfully", sP.Name))

	go func() {
		defer concurrent.Recoverer("SPMODIFIED")
		defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
		userID, _ := app.getCurrentUserID(r)

		logEvent := GetSystemLogEvent(userID, logAction, fmt.Sprintf(" %s %s,IP %s", sP.EndPointName, sP.HttpMethod, r.RemoteAddr), false)
		logEvent.ImpactedEndpointId = id
		logEvent.BeforeUpdate = logBeforeImage
		logEvent.AfterUpdate = sP.LogImage()
		app.SystemLoggerChan <- logEvent

	}()

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
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
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

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	if sP.IsSpecial {
		app.sessionManager.Put(r.Context(), "error", "Not allowed to delete!")
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

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
func (app *application) SPUpdate(w http.ResponseWriter, r *http.Request) {

	spId := chi.URLParam(r, "spId")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Update error: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)

	data.Form = sP
	data.Servers = app.servers.List()
	data.StoredProc = sP

	app.render(w, r, http.StatusOK, "sp_add.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SPCall(w http.ResponseWriter, r *http.Request) {

	spId := r.FormValue("id")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	if sP.DefaultServer != nil && sP.DefaultServer.ID != "" {

		dServer, err := app.servers.Get(sP.DefaultServer.ID)
		if err == nil {

			_, err = dServer.DummyCall(sP, formToMap(r), app.GetParamValidatorRegex())

		}
	} else {
		err = errors.New("default server is not defined")
	}

	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error call Stored proc: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	app.storedProcs.Save(sP)
	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SPsaveparamalias(w http.ResponseWriter, r *http.Request) {
	if !app.features.AllowParameterAlias {
		//app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusNotFound)
		return
	}

	spId := r.FormValue("id")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	if sP.IsSpecial {
		app.sessionManager.Put(r.Context(), "error", "Not allowed")
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	//	app.storedProcs.Save(sP)
	aliasMap := formToMap(r)

	changed := false
	for _, p := range sP.Parameters {
		alias, found := aliasMap[strings.ToUpper(p.Name)]
		if found {
			aliasString, ok := alias.(string)
			if ok && p.Alias != aliasString && p.Placement != "PATH" {
				p.Alias = strings.TrimSpace(strings.ToUpper(aliasString))
				changed = true

			}
		}
	}

	if changed {
		err = sP.ValidateAlias()
		if err != nil {
			app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
			app.goBack(w, r, http.StatusSeeOther)
			return
		}
		app.invalidateEndPointCache()

		sP.AssignAliasForPathPlacement()
		app.storedProcs.Save(sP)
		app.sessionManager.Put(r.Context(), "flash", "Done")

	} else {
		app.sessionManager.Put(r.Context(), "flash", "Nothing changed")

	}

	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
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

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) AssignServer(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	spId := r.Form.Get("spid")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Endpoint not found: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	serverid := r.Form.Get("serverid")
	server, err := app.servers.Get(serverid)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Server not found %s", serverid))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}
	sP.AddAllowedServer(server.ID, server.Name)
	app.storedProcs.Save(sP)
	app.invalidateEndPointCache()

	app.sessionManager.Put(r.Context(), "flase", "Done")
	app.goBack(w, r, http.StatusSeeOther)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RemoveAssignServer(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	spId := r.Form.Get("spid")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Endpoint not found: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	serverid := r.Form.Get("serverid")
	server, err := app.servers.Get(serverid)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Server not found %s", serverid))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}
	if server.ID == sP.DefaultServer.ID {
		app.sessionManager.Put(r.Context(), "error", "Can not remove default server")
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	sP.DeleteAllowedServer(server.ID)

	app.storedProcs.Save(sP)
	app.invalidateEndPointCache()

	app.sessionManager.Put(r.Context(), "flase", "Done")
	app.goBack(w, r, http.StatusSeeOther)
}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) ReloadSpTemplate(w http.ResponseWriter, r *http.Request) {

	app.LoadSPTemplates()

	app.sessionManager.Put(r.Context(), "flash", "Done")

	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) SpParamPos(w http.ResponseWriter, r *http.Request) {

	if !app.features.AllowParamPlacement {
		app.goBack(w, r, http.StatusNotFound)
		return
	}

	data := app.newTemplateData(r)

	spId := chi.URLParam(r, "spId")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}

	if sP.IsSpecial {
		app.sessionManager.Put(r.Context(), "error", "Not allowed")
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	data.StoredProc = sP
	data.ParamPlacements = sP.AvailableParamterPostions()
	app.render(w, r, http.StatusOK, "sp_param_placement.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SpSaveParamPos(w http.ResponseWriter, r *http.Request) {

	if !app.features.AllowParamPlacement {
		app.goBack(w, r, http.StatusNotFound)
		return
	}

	spId := r.FormValue("id")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	if sP.IsSpecial {
		app.sessionManager.Put(r.Context(), "error", "Not allowed")
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	//	app.storedProcs.Save(sP)
	formMap := formToMap(r)

	changed := false
	for _, p := range sP.Parameters {
		placement, found := formMap[strings.ToUpper(p.Name)]
		if found {
			placementString, ok := placement.(string)
			if ok && p.Placement != placementString {
				p.Placement = strings.TrimSpace(strings.ToUpper(placementString))

				changed = true

			}
		}
	}

	if changed {
		app.invalidateEndPointCache()

		sP.AssignAliasForPathPlacement()
		app.storedProcs.Save(sP)
		app.sessionManager.Put(r.Context(), "flash", "Done")
	} else {
		app.sessionManager.Put(r.Context(), "flash", "Nothing changed")

	}

	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) spParamValidator(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)

	spId := chi.URLParam(r, "spId")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}

	if sP.IsSpecial {
		app.sessionManager.Put(r.Context(), "error", "Not allowed")
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	data.StoredProc = sP

	data.ParamRegexs = app.paramRegexModel.List()
	app.render(w, r, http.StatusOK, "sp_param_validator.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) spsaveparamValidator(w http.ResponseWriter, r *http.Request) {

	spId := r.FormValue("id")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}
	if sP.IsSpecial {
		app.sessionManager.Put(r.Context(), "error", "Not allowed")
		app.goBack(w, r, http.StatusBadRequest)
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
// run promotions
// ------------------------------------------------------
func (app *application) FindSP(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)
	serverID := r.URL.Query().Get("serverid")

	server, err := app.servers.Get(serverID)
	if err == nil {

		libName := r.URL.Query().Get("lib")

		spName := r.URL.Query().Get("sp")
		if strings.TrimSpace(spName) != "" {
			sps, err := server.SearchSP(libName, spName)
			if err != nil {
				app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
			}
			data.StoredProcs = sps
		}
		data.Server = server
	}
	data.Servers = app.servers.List()
	app.render(w, r, http.StatusOK, "server_search_sp.tmpl", data)

}
