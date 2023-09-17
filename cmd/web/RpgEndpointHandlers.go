package main

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5"
	"github.com/onlysumitg/GoQhttp/internal/rpg"
	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/utils/concurrent"
	"github.com/onlysumitg/GoQhttp/utils/stringutils"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RpgEndpointHandlers(router *chi.Mux) {
	router.Route("/rpgendpoint", func(r chi.Router) {
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

		// r.Post("/assignserver", app.RpgEndpointAssignServer)
		// r.Post("/deleteassignserver", app.RpgEndpointRemoveAssignServer)

		r.Get("/logs/{spId}", app.RpgEndpointLogs)

		r.Get("/paramalias/{spId}", app.RpgEndpointParamAlias)
		//	r.Post("/saveparamalias", app.RpgEndpointsaveparamalias)

		r.Get("/paramplacement/{spId}", app.RpgEndpointParamPos)
		r.Post("/saveparamplacement", app.RpgEndpointSaveParamPos)

		r.Get("/paramvalid/{spId}", app.RpgEndpointParamValidator)
		r.Post("/saveparamvalid", app.RpgEndpointsaveparamValidator)

		r.Get("/help", app.RpgEndpointHelp)

	})

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

	data := app.newTemplateData(r)
	data.Form = storedP
	data.Servers = app.servers.List()
	app.render(w, r, http.StatusOK, "rpg_endpoint_add.tmpl", data)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) RpgEndpointRefresh(w http.ResponseWriter, r *http.Request) {

	spId := chi.URLParam(r, "spId")

	_, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}

	// // check default server
	// if sP.DefaultServerId != "" {
	// 	dServer, err := app.servers.Get(sP.DefaultServerId)
	// 	if err == nil {

	// 		err = dServer.PrepareToSave(r.Context(), sP)
	// 		if err == nil {
	// 			app.RpgEndpointModel.Save(sP)
	// 			app.sessionManager.Put(r.Context(), "flash", "Done")
	// 		}
	// 	}
	// } else {
	// 	err = errors.New("Default Server is not defined.")
	// }

	if err != nil {
		app.sessionManager.Put(r.Context(), "error", err.Error())

	}
	app.invalidateEndPointCache()

	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) RpgEndpointView(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	spId := chi.URLParam(r, "spId")

	sP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.RpgEndPoint = sP
	data.Servers = app.servers.List()
	//data.RpgEndpointCallLog, _ = app.spCallLogModel.Get(sP.ID)
	app.render(w, r, http.StatusOK, "rpg_endpoint_view.tmpl", data)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) RpgEndpointLogs(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	spId := chi.URLParam(r, "spId")

	sP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.RpgEndPoint = sP
	//data.RpgEndpointCallLog, _ = app.spCallLogModel.Get(sP.ID)
	app.render(w, r, http.StatusOK, "rpg_endpoint_logs.tmpl", data)

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

	sP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.RpgEndPoint = sP
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

	var sP rpg.RpgEndPoint
	err = app.formDecoder.Decode(&sP, r.PostForm)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("002 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}
	sP.CheckField(validator.NotBlank(sP.EndPointName), "endpointname", "This field cannot be blank")

	//sP.CheckField(!app.RpgEndpointModel.Duplicate(&sP), "endpointname", "Endpoint with name and method already exists")
	sP.CheckField(validator.NotBlank(sP.DefaultServerId), "serverid", "This field cannot be blank")

	sP.SetNameSpace()

	if sP.Valid() {

		sP.EndPointName = stringutils.RemoveSpecialChars(stringutils.RemoveMultipleSpaces(sP.EndPointName))

		sP.CheckField(!app.RpgEndpointModel.Duplicate(&sP), "endpointname", "Endpoint with name and method already exists in this namespace.")
	}
	// assign default server

	server, err := app.servers.Get(sP.DefaultServerId)
	if err != nil {
		sP.CheckField(false, "serverid", "Server not found")

		sP.Validator.AddNonFieldError("Please select a valid server")
	} else {
		sP.DefaultServerId = server.ID
		sP.AddAllowedServer(server.ID)
	}

	logBeforeImage := ""

	// Check RpgEndpoint details from iBMI
	if sP.Valid() {

		if sP.ID != "" {
			//orginalSp, err := app.RpgEndpointModel.Get(sP.ID)
			if err == nil {
				//logBeforeImage = orginalSp.LogImage()
				//sP.Parameters = orginalSp.Parameters

			}
		}

		//err = server.PrepareToSave(r.Context(), &sP)

		if err != nil {
			sP.CheckField(false, "name", err.Error())
		}

	}

	if !sP.Valid() {
		data := app.newTemplateData(r)
		data.Form = sP
		app.sessionManager.Put(r.Context(), "error", "Please fix error(s) and resubmit")
		data.Servers = app.servers.List()

		app.render(w, r, http.StatusUnprocessableEntity, "rpg_endpoint_add.tmpl", data)
		return
	}

	logAction := "Endpoint Created"
	if sP.ID != "" {
		logAction = "Endpoint Modified"
	}

	// finally save
	id, err := app.RpgEndpointModel.Save(&sP)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	app.invalidateEndPointCache()
	app.sessionManager.Put(r.Context(), "flash", fmt.Sprintf("Endpoint %s added sucessfully", sP.EndPointName))

	go func() {
		defer concurrent.Recoverer("RpgEndpointMODIFIED")
		defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
		userID, _ := app.getCurrentUserID(r)

		logEvent := GetSystemLogEvent(userID, logAction, fmt.Sprintf(" %s %s,IP %s", sP.EndPointName, sP.HttpMethod, r.RemoteAddr), false)
		logEvent.ImpactedEndpointId = id
		logEvent.BeforeUpdate = logBeforeImage
		//logEvent.AfterUpdate = sP.LogImage()
		app.SystemLoggerChan <- logEvent

	}()

	//http.Redirect(w, r, fmt.Sprintf("/savesql/%s", id), http.StatusSeeOther)
	http.Redirect(w, r, fmt.Sprintf("/sp/%s", id), http.StatusSeeOther)

}

// ------------------------------------------------------
// Delete saved query
// ------------------------------------------------------
func (app *application) RpgEndpointDelete(w http.ResponseWriter, r *http.Request) {

	spId := chi.URLParam(r, "spId")

	sP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	data := app.newTemplateData(r)
	data.RpgEndPoint = sP

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

	http.Redirect(w, r, "/sp", http.StatusSeeOther)

}

// ------------------------------------------------------
// add new server
// ------------------------------------------------------
func (app *application) RpgEndpointUpdate(w http.ResponseWriter, r *http.Request) {

	spId := chi.URLParam(r, "spId")

	sP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Update error: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)

	data.Form = sP
	data.Servers = app.servers.List()
	data.RpgEndPoint = sP

	app.render(w, r, http.StatusOK, "rpg_endpoint_add.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RpgEndpointCall(w http.ResponseWriter, r *http.Request) {

	spId := r.FormValue("id")

	sP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	if sP.DefaultServerId != "" {

		//dServer, err := app.servers.Get(sP.DefaultServerId)
		if err == nil {

			//_, err = dServer.DummyCall(sP, formToMap(r), app.GetParamValidatorRegex())

		}
	} else {
		err = errors.New("default server is not defined")
	}

	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error call Stored proc: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	app.RpgEndpointModel.Save(sP)
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

	sP, err := app.RpgEndpointModel.Get(spId)
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
	sP.AddAllowedServer(server.ID)
	app.RpgEndpointModel.Save(sP)
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

	sP, err := app.RpgEndpointModel.Get(spId)
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
	if server.ID == sP.DefaultServerId {
		app.sessionManager.Put(r.Context(), "error", "Can not remove default server")
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	sP.DeleteAllowedServer(server.ID)

	app.RpgEndpointModel.Save(sP)
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

	sP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.RpgEndPoint = sP
	//data.ParamPlacements = sP.AvailableParamterPostions()
	app.render(w, r, http.StatusOK, "rpg_endpoint_param_placement.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RpgEndpointSaveParamPos(w http.ResponseWriter, r *http.Request) {

	// if !app.features.AllowParamPlacement {
	// 	app.goBack(w, r, http.StatusNotFound)
	// 	return
	// }

	// spId := r.FormValue("id")

	// sP, err := app.RpgEndpointModel.Get(spId)
	// if err != nil {
	// 	app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
	// 	app.goBack(w, r, http.StatusSeeOther)
	// 	return
	// }

	// //	app.RpgEndpointModel.Save(sP)
	// formMap := formToMap(r)

	// changed := false
	// for _, p := range sP.Parameters {
	// 	placement, found := formMap[strings.ToUpper(p.Name)]
	// 	if found {
	// 		placementString, ok := placement.(string)
	// 		if ok && p.Placement != placementString {
	// 			p.Placement = strings.TrimSpace(strings.ToUpper(placementString))

	// 			changed = true

	// 		}
	// 	}
	// }

	// if changed {
	// 	app.invalidateEndPointCache()

	// 	sP.AssignAliasForPathPlacement()
	// 	app.RpgEndpointModel.Save(sP)
	// 	app.sessionManager.Put(r.Context(), "flash", "Done")
	// } else {
	// 	app.sessionManager.Put(r.Context(), "flash", "Nothing changed")

	// }

	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) RpgEndpointParamValidator(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)

	spId := chi.URLParam(r, "spId")

	sP, err := app.RpgEndpointModel.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.RpgEndPoint = sP

	data.ParamRegexs = app.paramRegexModel.List()
	app.render(w, r, http.StatusOK, "rpg_endpoint_param_validator.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RpgEndpointsaveparamValidator(w http.ResponseWriter, r *http.Request) {

	// spId := r.FormValue("id")

	// sP, err := app.RpgEndpointModel.Get(spId)
	// if err != nil {
	// 	app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
	// 	app.goBack(w, r, http.StatusSeeOther)
	// 	return
	// }

	// //	app.RpgEndpointModel.Save(sP)
	// maped := formToMap(r)

	// changed := false
	// for _, p := range sP.Parameters {
	// 	validator, found := maped[strings.ToUpper(p.Name)]
	// 	if found {
	// 		validatorString, ok := validator.(string)
	// 		if ok && p.ValidatorRegex != validatorString {

	// 			p.ValidatorRegex = validatorString

	// 			if p.ValidatorRegex != "" {
	// 				err := p.CheckValidatorRegex(app.paramRegexModel.Map())
	// 				if err != nil {
	// 					app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
	// 					app.goBack(w, r, http.StatusSeeOther)
	// 					return
	// 				}
	// 			}

	// 			changed = true

	// 		}
	// 	}
	// }

	// if changed {
	// 	app.invalidateEndPointCache()
	// 	app.RpgEndpointModel.Save(sP)
	// 	app.sessionManager.Put(r.Context(), "flash", "Done")

	// } else {
	// 	app.sessionManager.Put(r.Context(), "flash", "Nothing changed")

	// }

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
