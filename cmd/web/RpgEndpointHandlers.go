package main

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
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
func (app *application) RpgEndpointHandlers(router *chi.Mux) {
	router.Route("/pgmendpoints", func(r chi.Router) {
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(CheckLicMiddleware)

		//r.Use(app.CurrentServerMiddleware)
		//r.With(paginate).Get("/", listArticles)
		r.Get("/", app.RpgEndpointList)
		r.Get("/find", app.FindRpgEndpoint)
		r.Post("/call", app.RpgEndpointCall)
		r.Get("/{spId}", app.RpgEndpointView)

		r.Get("/add", app.RpgEndpointAdd)
		r.Post("/add", app.RpgEndpointAddPost)

		r.Get("/update/{spId}", app.RpgEndpointUpdate)
		r.Post("/update/{spId}", app.RpgEndpointAddPost)

		r.Get("/delete/{spId}", app.RpgEndpointDelete)
		r.Post("/delete", app.RpgEndpointDeleteConfirm)

		// r.Get("/run/{spId}", app.RpgEndpointRun)
		// r.Get("/run", app.RpgEndpointRun)

		// r.Post("/build", app.RpgEndpointBuild)

		r.Get("/refresh/{spId}", app.RpgEndpointRefresh)

		r.Post("/assignserver", app.RpgEndpointAssignServer)
		r.Post("/deleteassignserver", app.RpgEndpointRemoveAssignServer)

		r.Get("/logs/{spId}", app.RpgEndpointLogs)

		//r.Get("/paramalias/{spId}", app.RpgEndpointParamAlias)
		//	r.Post("/saveparamalias", app.RpgEndpointsaveparamalias)

		r.Get("/paramplacement/{spId}", app.RpgEndpointParamPos)
		r.Post("/saveparamplacement", app.RpgEndpointSaveParamPos)

		r.Get("/help", app.RpgEndpointHelp)
		r.Get("/fieldrow", app.rpgParam_Field_row)

	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) rpgParam_Field_row(w http.ResponseWriter, r *http.Request) {

	currentindex := r.URL.Query().Get("index")

	currentindexI, err := strconv.Atoi(currentindex)
	if err != nil {
		currentindexI = 0
	}

	currentindexI += 1

	data := app.newTemplateData(r)
	programParam := &rpg.ProgramParams{
		NameToUse: "",
		Dim:       0,
	}

	data.ProgramParam = programParam
	data.Index = currentindexI

	data.RpgParams = app.RpgParamModel.List()

	app.render(w, r, http.StatusOK, "empty_rpg_endpoint_add_param_row.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RpgEndpointHelp(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	app.render(w, r, http.StatusOK, "rpg_endpoint_help_inbuilt_param.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RpgEndpointList(w http.ResponseWriter, r *http.Request) {

	serverID := r.URL.Query().Get("server")

	_, err := app.servers.Get(serverID)
	if err != nil {
		serverID = ""

	}

	data := app.newTemplateData(r)
	data.RpgEndPoints = make([]*rpg.RpgEndPoint, 0, 10)

	storedPs := app.RpgEndpointModel.List()

	if serverID != "" {
		for _, s := range storedPs {
			if s == nil {
				continue
			}
			allowed := false
			if s.DefaultServerId == serverID {
				allowed = true
			} else {

				for _, als := range s.AllowedOnServers {
					if als == serverID {
						allowed = true
					}

				}
			}
			if allowed {
				data.RpgEndPoints = append(data.RpgEndPoints, s)
			}
		}
	} else {
		data.RpgEndPoints = storedPs
	}

	nextUrl := r.URL.Query().Get("next") //filters=["color", "price", "brand"]
	data.Next = nextUrl
	app.render(w, r, http.StatusOK, "rpg_endpoint_list.tmpl", data)

}

// ------------------------------------------------------

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RpgEndpointAdd(w http.ResponseWriter, r *http.Request) {

	storedP := rpg.RpgEndPoint{}

	serverID := r.URL.Query().Get("server")

	server, err := app.servers.Get(serverID)
	if err == nil {
		storedP.DefaultServerId = server.ID

	}

	storedP.Init()

	data := app.newTemplateData(r)

	data.Form = storedP
	data.Servers = app.servers.List()
	data.RpgParams = app.RpgParamModel.List()

	data.Index = len(storedP.Parameters) - 1

	//data.RpgPrograms = app.RpgProgramModel.List()
	app.render(w, r, http.StatusOK, "rpg_endpoint_add.tmpl", data)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) RpgEndpointRefresh(w http.ResponseWriter, r *http.Request) {

	spId := chi.URLParam(r, "spId")

	rpgEndPoint, err := app.RpgEndpointModel.Get(spId)

	if err != nil {
		app.sessionManager.Put(r.Context(), "error", err.Error())

	}

	rpgEndPoint.Refresh()

	app.RpgEndpointModel.Save(rpgEndPoint)

	app.invalidateEndPointCache()

	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) RpgEndpointView(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	spId := chi.URLParam(r, "spId")

	pgmP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.RpgEndPoint = pgmP
	data.Servers = app.servers.List()
	//data.RpgEndpointCallLog, _ = app.spCallLogModel.Get(pgmP.ID)
	app.render(w, r, http.StatusOK, "rpg_endpoint_view.tmpl", data)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) RpgEndpointLogs(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	spId := chi.URLParam(r, "spId")

	pgmP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.RpgEndPoint = pgmP
	data.SPCallLog, _ = app.spCallLogModel.Get(pgmP.ID)
	app.render(w, r, http.StatusOK, "sp_logs.tmpl", data)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) RpgEndpointParamAlias(w http.ResponseWriter, r *http.Request) {

	if !app.features.AllowParameterAlias {
		//app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusNotFound)
		return
	}

	data := app.newTemplateData(r)

	spId := chi.URLParam(r, "spId")

	pgmP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.RpgEndPoint = pgmP
	app.render(w, r, http.StatusOK, "rpg_endpoint_param_alias.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RpgEndpointAddPost(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("001 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	var pgmP rpg.RpgEndPoint
	err = app.formDecoder.Decode(&pgmP, r.PostForm)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("002 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}
	pgmP.CheckField(validator.NotBlank(pgmP.EndPointName), "endpointname", "This field cannot be blank")
	pgmP.CheckField(validator.NotBlank(pgmP.Name), "name", "This field cannot be blank")
	pgmP.CheckField(validator.NotBlank(pgmP.Lib), "lib", "This field cannot be blank")

	//pgmP.CheckField(!app.RpgEndpointModel.Duplicate(&pgmP), "endpointname", "Endpoint with name and method already exists")
	pgmP.CheckField(validator.NotBlank(pgmP.DefaultServerId), "serverid", "This field cannot be blank")

	pgmP.SetNameSpace()

	if pgmP.Valid() {

		pgmP.EndPointName = stringutils.RemoveSpecialChars(stringutils.RemoveMultipleSpaces(pgmP.EndPointName))

		pgmP.CheckField(!app.RpgEndpointModel.Duplicate(&pgmP), "endpointname", "Endpoint with name and method already exists in this namespace.")
		pgmP.CheckField(!app.storedProcs.DuplicateByName(pgmP.Name, pgmP.HttpMethod, pgmP.Namespace), "endpointname", "Endpoint with name and method already exists in this namespace.")

	}
	// assign default server

	server, err := app.servers.Get(pgmP.DefaultServerId)
	if err != nil {
		pgmP.CheckField(false, "serverid", "Server not found")

		pgmP.Validator.AddNonFieldError("Please select a valid server")
	} else {
		_, err := app.GetRPGDriver(server)
		if err != nil {
			pgmP.CheckField(false, "serverid", "Program Drivers not available")
			pgmP.Validator.AddNonFieldError("Program Drivers are not available for this server")

		} else {

			pgmP.DefaultServerId = server.ID

			pgmP.AddAllowedServer(server.ID)
		}
	}

	logBeforeImage := ""

	// Check RpgEndpoint details from iBMI
	if pgmP.Valid() {

		if pgmP.ID != "" {
			//orginalSp, err := app.RpgEndpointModel.Get(pgmP.ID)
			if err == nil {
				//logBeforeImage = orginalSp.LogImage()
				//pgmP.Parameters = orginalSp.Parameters

			}
		}

		//err = server.PrepareToSave(r.Context(), &pgmP)

		if err != nil {
			pgmP.CheckField(false, "name", err.Error())
		}

	}

	pgmP.AssignParamObjects(app.RpgParamModel)
	pgmP.AssignParamNames()
	parametersError := pgmP.ValidateParams()

	if !pgmP.Valid() || parametersError {
		pgmP.FilterOutInvalidParams()
		data := app.newTemplateData(r)
		data.Form = pgmP
		app.sessionManager.Put(r.Context(), "error", "Please fix error(s) and resubmit")
		data.Servers = app.servers.List()

		data.Index = len(pgmP.Parameters) - 1
		data.RpgParams = app.RpgParamModel.List()
		app.render(w, r, http.StatusUnprocessableEntity, "rpg_endpoint_add.tmpl", data)
		return
	}

	logAction := "Endpoint Created"
	if pgmP.ID != "" {
		logAction = "Endpoint Modified"
	}

	pgmP.FilterOutInvalidParams()
	pgmP.Refresh()
	// finally save
	id, err := app.RpgEndpointModel.Save(&pgmP)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	app.invalidateEndPointCache()
	app.sessionManager.Put(r.Context(), "flash", fmt.Sprintf("Endpoint %s added sucessfully", pgmP.EndPointName))

	go func() {
		defer concurrent.Recoverer("RpgEndpointMODIFIED")
		defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
		userID, _ := app.getCurrentUserID(r)

		logEvent := GetSystemLogEvent(userID, logAction, fmt.Sprintf(" %s %s,IP %s", pgmP.EndPointName, pgmP.HttpMethod, r.RemoteAddr), false)
		logEvent.ImpactedEndpointId = id
		logEvent.BeforeUpdate = logBeforeImage
		//logEvent.AfterUpdate = pgmP.LogImage()
		app.SystemLoggerChan <- logEvent

	}()

	//http.Redirect(w, r, fmt.Sprintf("/savesql/%s", id), http.StatusSeeOther)
	http.Redirect(w, r, fmt.Sprintf("/pgmendpoints/%s", id), http.StatusSeeOther)

}

// ------------------------------------------------------
// Delete saved query
// ------------------------------------------------------
func (app *application) RpgEndpointDelete(w http.ResponseWriter, r *http.Request) {

	spId := chi.URLParam(r, "spId")

	pgmP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	data := app.newTemplateData(r)
	data.RpgEndPoint = pgmP

	app.render(w, r, http.StatusOK, "rpg_endpoint_delete.tmpl", data)

}

// ------------------------------------------------------
// Delete saved query confirm
// ------------------------------------------------------
func (app *application) RpgEndpointDeleteConfirm(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("001 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	spId := r.PostForm.Get("spId")
	app.invalidateEndPointCache()

	err = app.RpgEndpointModel.Delete(spId)
	if err != nil {

		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Delete error: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	go app.deleteRpgEndpointData(spId) //goroutine
	app.sessionManager.Put(r.Context(), "flash", "Endpoint deleted sucessfully")

	go func() {
		defer concurrent.Recoverer("RpgEndpointMODIFIED")
		userID, _ := app.getCurrentUserID(r)

		logEvent := GetSystemLogEvent(userID, "Endpoint Deleted", fmt.Sprintf("IP %s", r.RemoteAddr), false)
		logEvent.ImpactedEndpointId = spId
		app.SystemLoggerChan <- logEvent

	}()

	http.Redirect(w, r, "/pgmendpoints", http.StatusSeeOther)

}

// ------------------------------------------------------
// add new server
// ------------------------------------------------------
func (app *application) RpgEndpointUpdate(w http.ResponseWriter, r *http.Request) {

	spId := chi.URLParam(r, "spId")

	pgmP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Update error: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)

	data.Form = pgmP
	data.Servers = app.servers.List()
	data.RpgEndPoint = pgmP

	data.RpgParams = app.RpgParamModel.List()

	data.Index = len(pgmP.Parameters) - 1

	app.render(w, r, http.StatusOK, "rpg_endpoint_add.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RpgEndpointCall(w http.ResponseWriter, r *http.Request) {

	spId := r.FormValue("id")

	pgmP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	if pgmP.DefaultServerId != "" {

		//dServer, err := app.servers.Get(pgmP.DefaultServerId)
		if err == nil {

			//_, err = dServer.DummyCall(pgmP, formToMap(r), app.GetParamValidatorRegex())

		}
	} else {
		err = errors.New("default server is not defined")
	}

	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error call Stored proc: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	app.RpgEndpointModel.Save(pgmP)
	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RpgEndpointAssignServer(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	spId := r.Form.Get("spid")

	pgmP, err := app.RpgEndpointModel.Get(spId)
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
	pgmP.AddAllowedServer(server.ID)
	app.RpgEndpointModel.Save(pgmP)
	app.invalidateEndPointCache()

	app.sessionManager.Put(r.Context(), "flase", "Done")
	app.goBack(w, r, http.StatusSeeOther)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RpgEndpointRemoveAssignServer(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	spId := r.Form.Get("spid")

	pgmP, err := app.RpgEndpointModel.Get(spId)
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
	if server.ID == pgmP.DefaultServerId {
		app.sessionManager.Put(r.Context(), "error", "Can not remove default server")
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	pgmP.DeleteAllowedServer(server.ID)

	app.RpgEndpointModel.Save(pgmP)
	app.invalidateEndPointCache()

	app.sessionManager.Put(r.Context(), "flase", "Done")
	app.goBack(w, r, http.StatusSeeOther)
}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) ReloadRpgEndpointTemplate(w http.ResponseWriter, r *http.Request) {

	app.LoadSPTemplates()

	app.sessionManager.Put(r.Context(), "flash", "Done")

	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) RpgEndpointParamPos(w http.ResponseWriter, r *http.Request) {

	if !app.features.AllowParamPlacement {
		app.goBack(w, r, http.StatusNotFound)
		return
	}

	data := app.newTemplateData(r)

	spId := chi.URLParam(r, "spId")

	pgmP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.RpgEndPoint = pgmP
	data.ParamPlacements = pgmP.AvailableParamterPostions()
	app.render(w, r, http.StatusOK, "rpg_endpoint_param_placement.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RpgEndpointSaveParamPos(w http.ResponseWriter, r *http.Request) {

	if !app.features.AllowParamPlacement {
		app.goBack(w, r, http.StatusNotFound)
		return
	}

	spId := r.FormValue("id")

	pgmP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	//	app.RpgEndpointModel.Save(pgmP)
	formMap := formToMap(r)

	changed := false
	for _, p := range pgmP.Parameters {
		placement, found := formMap[p.GetNameToUse()]
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

		pgmP.AssignAliasForPathPlacement()
		pgmP.Refresh()
		app.RpgEndpointModel.Save(pgmP)
		app.sessionManager.Put(r.Context(), "flash", "Done")
	} else {
		app.sessionManager.Put(r.Context(), "flash", "Nothing changed")

	}

	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
// run promotions
// ------------------------------------------------------
func (app *application) FindRpgEndpoint(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)
	// serverID := r.URL.Query().Get("serverid")

	// server, err := app.servers.Get(serverID)
	// if err == nil {

	// 	libName := r.URL.Query().Get("lib")

	// 	spName := r.URL.Query().Get("sp")
	// 	if strings.TrimSpace(spName) != "" {
	// 		sps, err := server.SearchRpgEndpoint(libName, spName)
	// 		if err != nil {
	// 			app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
	// 		}
	// 		data.RpgEndPoints = sps
	// 	}
	// 	data.Server = server
	// }
	// data.Servers = app.servers.List()
	app.render(w, r, http.StatusOK, "server_search_sp.tmpl", data)

}
