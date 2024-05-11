package main

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5"
	"github.com/zerobit-tech/GoQhttp/internal/ibmiServer"
	"github.com/zerobit-tech/GoQhttp/internal/models"
	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
	"github.com/zerobit-tech/GoQhttp/internal/validator"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) ServerListMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		next.ServeHTTP(w, r)
	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) CurrentServerMiddlewareXX(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		currentServerID := app.sessionManager.GetString(r.Context(), "currentserver")
		server, err2 := app.servers.Get(currentServerID)

		continueToNext := true
		message := "Please select a server"
		if err2 != nil {
			continueToNext = false
		} else if server.OnHold {
			continueToNext = false
			message = "Server is on hold. Please select a differnt server"
			app.sessionManager.Remove(r.Context(), "currentserver")

		}

		if continueToNext {
			next.ServeHTTP(w, r)
		} else {
			app.sessionManager.Put(r.Context(), "warning", message)

			goToUrl := fmt.Sprintf("/servers?next=%s", r.URL.RequestURI())

			reponseMap := make(map[string]string)
			reponseMap["redirectTo"] = goToUrl
			switch r.Header.Get("Accept") {

			case "application/json":
				app.writeJSON(w, http.StatusSeeOther, reponseMap, nil)

			default:
				http.Redirect(w, r, goToUrl, http.StatusSeeOther)
			}
		}

	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) ServerHandlers(router *chi.Mux) {
	router.Route("/servers", func(r chi.Router) {
		//r.With(paginate).Get("/", listArticles)
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(CheckLicMiddleware)

		r.Get("/", app.ServerList)
		r.Get("/{serverid}", app.ServerView)
		r.Get("/select/{serverid}", app.ServerSelect)
		r.Get("/listprom/{serverid}", app.ListPromotion)
		r.Get("/liblist/{serverid}", app.GetLibList)
		//	r.Get("/findsp/{serverid}", app.FindSP)

		superadmingroup := r.Group(nil)

		superadmingroup.Use(app.RequireSuperAdmin)
		superadmingroup.Get("/add", app.ServerAdd)
		superadmingroup.Post("/add", app.ServerAddPost)

		superadmingroup.Get("/update/{serverid}", app.ServerUpdate)
		superadmingroup.Post("/update", app.ServerUpdatePost)

		superadmingroup.Get("/delete/{serverid}", app.ServerDelete)
		superadmingroup.Post("/delete", app.ServerDeleteConfirm)

		superadmingroup.Get("/runpromotions/{serverid}", app.RunPromotion)

		superadmingroup.Get("/clearcache/{serverid}", app.ClearCache)

		superadmingroup.Get("/syncusertoken/{serverid}", app.SyncUserTokens)

		superadmingroup.Get("/promotiontable/{serverid}", app.showPromotionTable)
		superadmingroup.Get("/crtpromotiontable/{serverid}", app.createPromotionTable)

	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) ServerList(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Servers = app.servers.List()
	nextUrl := r.URL.Query().Get("next") //filters=["color", "price", "brand"]
	if nextUrl == "" {
		nextUrl = "/sp"
	}
	data.Next = nextUrl
	app.render(w, r, http.StatusOK, "server_list.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) ServerSelect(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverid")
	server, err := app.servers.Get(serverID)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	if server.OnHold {
		app.sessionManager.Put(r.Context(), "warning", "Server is on hold. Please select a differnt server")

	} else {
		app.sessionManager.Put(r.Context(), "currentserver", server.ID)
		app.sessionManager.Put(r.Context(), "flash", fmt.Sprintf("Selected server: %s", server.Name))

		nextUrl := r.URL.Query().Get("next") //filters=["color", "price", "brand"]

		if nextUrl != "" {
			http.Redirect(w, r, nextUrl, http.StatusSeeOther)
			return
		}
	}

	app.goBack(w, r, http.StatusSeeOther)

	//http.Redirect(w, r, "/query", http.StatusSeeOther)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) ServerView(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	serverID := chi.URLParam(r, "serverid")
	//log.Println("serverID >>>", serverID)
	server, err := app.servers.Get(serverID)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.Server = server

	data.Users = make([]*models.User, 0, 10)

	for _, u := range app.users.List() {
		if u.ServerId == serverID {
			data.Users = append(data.Users, u)
		}
	}

	app.render(w, r, http.StatusOK, "server_view.tmpl", data)

}

// ------------------------------------------------------
// Delete servet
// ------------------------------------------------------
func (app *application) ServerDelete(w http.ResponseWriter, r *http.Request) {

	serverID := chi.URLParam(r, "serverid")

	server, err := app.servers.Get(serverID)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error deleting server: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)
	data.Server = server

	data.StoredProcs = make([]*storedProc.StoredProc, 0, 10)
	for _, s := range app.storedProcs.List(false) {
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

	data.Users = make([]*models.User, 0, 10)

	for _, u := range app.users.List() {
		if u.ServerId == serverID {
			data.Users = append(data.Users, u)
		}
	}

	data.AllowServerDelete = true
	if len(data.Users) > 0 || len(data.StoredProcs) > 0 {
		data.AllowServerDelete = false
	}

	app.render(w, r, http.StatusOK, "server_delete.tmpl", data)

}

// ------------------------------------------------------
// Delete servet
// ------------------------------------------------------
func (app *application) ServerDeleteConfirm(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("001 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	serverID := r.PostForm.Get("serverid")

	err = app.servers.Delete(serverID)
	if err != nil {

		//log.Println("ServerDeleteConfirm  002 >>>>>>", err.Error())
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error deleting server: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	go func() {
		defer concurrent.Recoverer("Server ADD log")
		userID, _ := app.getCurrentUserID(r)

		logEvent := GetSystemLogEvent(userID, "Server Deleted", fmt.Sprintf("IP %s", r.RemoteAddr), false)
		logEvent.ImpactedServerId = serverID
		app.SystemLoggerChan <- logEvent

	}()

	app.sessionManager.Put(r.Context(), "flash", "Server deleted sucessfully")

	http.Redirect(w, r, "/servers", http.StatusSeeOther)

}

// ------------------------------------------------------
// run promotions
// ------------------------------------------------------
func (app *application) RunPromotion(w http.ResponseWriter, r *http.Request) {

	if !app.features.AllowPromotion {
		//app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusNotFound)
		return
	}

	serverID := chi.URLParam(r, "serverid")

	server, err := app.servers.Get(serverID)
	if err != nil {

		//log.Println("ServerDeleteConfirm  002 >>>>>>", err.Error())
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}
	go app.ProcessPromotion(server) //goroutine
	app.sessionManager.Put(r.Context(), "flash", "Queued. Please wait.")

	go func() {
		defer concurrent.Recoverer("Manual Promotion log")
		userID, _ := app.getCurrentUserID(r)
		logEvent := GetSystemLogEvent(userID, "Manual Promotion", fmt.Sprintf("Server %s,IP %s", server.Name, r.RemoteAddr), false)
		logEvent.ImpactedServerId = serverID
		app.SystemLoggerChan <- logEvent

	}()

	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
// run promotions
// ------------------------------------------------------
func (app *application) ClearCache(w http.ResponseWriter, r *http.Request) {

	serverID := chi.URLParam(r, "serverid")

	server, err := app.servers.Get(serverID)
	if err != nil {

		//log.Println("ServerDeleteConfirm  002 >>>>>>", err.Error())
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}
	err = server.ClearCache()
	if err != nil {

		//log.Println("ServerDeleteConfirm  002 >>>>>>", err.Error())
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Cache cleared")

	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
// run promotions
// ------------------------------------------------------
func (app *application) ListPromotion(w http.ResponseWriter, r *http.Request) {
	if !app.features.AllowPromotion {
		//app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusNotFound)
		return
	}

	data := app.newTemplateData(r)
	serverID := chi.URLParam(r, "serverid")

	server, err := app.servers.Get(serverID)
	if err != nil {

		//log.Println("ServerDeleteConfirm  002 >>>>>>", err.Error())
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}
	promotions, err := server.ListPromotion(false)

	if err != nil {

		//log.Println("ServerDeleteConfirm  002 >>>>>>", err.Error())
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}
	data.Server = server
	data.Promotions = promotions

	app.render(w, r, http.StatusOK, "server_promotion_table.tmpl", data)

}

// ------------------------------------------------------
// add new server
// ------------------------------------------------------
func (app *application) ServerAdd(w http.ResponseWriter, r *http.Request) {

	if app.features.MaxAllowedServers > 0 {
		if len(app.servers.List()) >= app.features.MaxAllowedServers {
			app.sessionManager.Put(r.Context(), "error", "Limit reached: Can not add more servers")
			app.goBack(w, r, http.StatusSeeOther)
			return
		}
	}

	data := app.newTemplateData(r)

	// set form initial values
	data.Form = ibmiServer.Server{
		ConnectionsOpen:   20,
		ConnectionsIdle:   20,
		ConnectionMaxAge:  600,
		ConnectionIdleAge: 3600,
		LibList:           make([]string, 20),
		//Namespace:         "V1",
	}
	app.render(w, r, http.StatusOK, "server_add.tmpl", data)

}

// ----------------------------------------------
func (app *application) ServerAddPost(w http.ResponseWriter, r *http.Request) {
	// Limit the request body size to 4096 bytes
	//r.Body = http.MaxBytesReader(w, r.Body, 4096)

	// r.ParseForm() method to parse the request body. This checks
	// that the request body is well-formed, and then stores the form data in the request’s
	// r.PostForm map.
	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("001 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	if app.features.MaxAllowedServers > 0 {
		if len(app.servers.List()) >= app.features.MaxAllowedServers {
			app.sessionManager.Put(r.Context(), "error", "Limit reached: Can not add more servers")
			app.goBack(w, r, http.StatusSeeOther)
			return
		}
	}

	// Use the r.PostForm.Get() method to retrieve the title and content
	// from the r.PostForm map.
	//	title := r.PostForm.Get("title")
	//	content := r.PostForm.Get("content")

	// the r.PostForm map is populated only for POST , PATCH and PUT requests, and contains the
	// form data from the request body.

	// In contrast, the r.Form map is populated for all requests (irrespective of their HTTP method),

	var server ibmiServer.Server
	// Call the Decode() method of the form decoder, passing in the current
	// request and *a pointer* to our snippetCreateForm struct. This will
	// essentially fill our struct with the relevant values from the HTML form.
	// If there is a problem, we return a 400 Bad Request response to the client.
	err = app.formDecoder.Decode(&server, r.PostForm)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("002 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	//fmt.Println(">>>> server lib >>", server.LibList)

	server.CheckField(validator.NotBlank(server.Name), "name", "This field cannot be blank")

	server.CheckField(validator.NotBlank(server.IP), "ip", "This field cannot be blank")
	server.CheckField(validator.NotBlank(server.UserName), "user_name", "This field cannot be blank")
	server.CheckField(validator.NotBlank(server.Password), "password", "This field cannot be blank")
	//server.CheckField(validator.NotBlank(server.WorkLib), "worklib", "This field cannot be blank")

	// Use the Valid() method to see if any of the checks failed. If they did,
	// then re-render the template passing in the form in the same way as
	// before.
	if server.Valid() {
		server.Name = stringutils.RemoveSpecialChars(stringutils.RemoveMultipleSpaces(server.Name))
		server.CheckField(!app.servers.DuplicateName(&server), "name", "Duplicate Name")
	}

	if !server.Valid() {
		data := app.newTemplateData(r)
		data.Form = &server
		app.sessionManager.Put(r.Context(), "error", "Please fix error(s) and resubmit")

		app.render(w, r, http.StatusUnprocessableEntity, "server_add.tmpl", data)
		return
	}
	server.Password, _ = stringutils.Encrypt(server.Password, server.GetSecretKey())

	id, err := app.servers.Insert(&server)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}

	//db, err := server.GetSingleConnection()
	// defer db.Close()
	// if err != nil {
	// 	server.Password = server.GetPassword()
	// 	app.servers.Delete(server.ID)
	// 	data := app.newTemplateData(r)
	// 	data.Form = server
	// 	app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Can not verify server connection %s", err.Error()))
	// 	app.render(w, r, http.StatusUnprocessableEntity, "server_add.tmpl", data)

	// 	return
	// }

	go func() {
		defer concurrent.Recoverer("Server ADD log")
		userID, _ := app.getCurrentUserID(r)
		logEvent := GetSystemLogEvent(userID, "Server Created", fmt.Sprintf("%s,IP %s", server.Name, r.RemoteAddr), false)
		logEvent.ImpactedServerId = server.ID
		logEvent.AfterUpdate = server.LogImage()

		app.SystemLoggerChan <- logEvent

	}()

	app.sessionManager.Put(r.Context(), "flash", fmt.Sprintf("Server %s added sucessfully", server.Name))

	http.Redirect(w, r, fmt.Sprintf("/servers/%s", id), http.StatusSeeOther)
}

// ------------------------------------------------------
// add new server
// ------------------------------------------------------
func (app *application) ServerUpdate(w http.ResponseWriter, r *http.Request) {

	serverID := chi.URLParam(r, "serverid")

	server, err := app.servers.Get(serverID)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error updating server: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	server.Password = ""
	data := app.newTemplateData(r)

	data.Form = server

	if len(server.LibList) == 0 {
		server.LibList = make([]string, 20)
	}
	data.Server = server
	app.render(w, r, http.StatusOK, "server_update.tmpl", data)

}

// ----------------------------------------------
func (app *application) ServerUpdatePost(w http.ResponseWriter, r *http.Request) {
	// Limit the request body size to 4096 bytes
	//r.Body = http.MaxBytesReader(w, r.Body, 4096)

	// r.ParseForm() method to parse the request body. This checks
	// that the request body is well-formed, and then stores the form data in the request’s
	// r.PostForm map.
	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("001 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	// Use the r.PostForm.Get() method to retrieve the title and content
	// from the r.PostForm map.
	//	title := r.PostForm.Get("title")
	//	content := r.PostForm.Get("content")

	// the r.PostForm map is populated only for POST , PATCH and PUT requests, and contains the
	// form data from the request body.

	// In contrast, the r.Form map is populated for all requests (irrespective of their HTTP method),

	var server ibmiServer.Server
	// Call the Decode() method of the form decoder, passing in the current
	// request and *a pointer* to our snippetCreateForm struct. This will
	// essentially fill our struct with the relevant values from the HTML form.
	// If there is a problem, we return a 400 Bad Request response to the client.
	err = app.formDecoder.Decode(&server, r.PostForm)

	//fmt.Println(">> decord form 1", server)

	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("002 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	originalServer, err := app.servers.Get(server.ID)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("003 Invalid server"))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}
	server.Name = stringutils.RemoveSpecialChars(stringutils.RemoveMultipleSpaces(server.Name))

	server.CheckField(!app.servers.DuplicateName(&server), "name", "Duplicate Name")

	server.CheckField(validator.NotBlank(server.Name), "name", "This field cannot be blank")
	server.CheckField(validator.NotBlank(server.IP), "ip", "This field cannot be blank")
	server.CheckField(validator.NotBlank(server.UserName), "user_name", "This field cannot be blank")
	//server.CheckField(validator.NotBlank(server.Password), "password", "This field cannot be blank")
	//server.CheckField(validator.NotBlank(server.WorkLib), "worklib", "This field cannot be blank")

	// Use the Valid() method to see if any of the checks failed. If they did,
	// then re-render the template passing in the form in the same way as
	// before.

	if !server.Valid() {
		data := app.newTemplateData(r)
		data.Form = server
		app.sessionManager.Put(r.Context(), "error", "Please fix error(s) and resubmit")

		app.render(w, r, http.StatusUnprocessableEntity, "server_update.tmpl", data)
		return
	}

	// if new server password is not set --> get it from original server
	if server.Password == "" {
		server.Password = originalServer.Password
	} else {
		server.Password, _ = stringutils.Encrypt(server.Password, server.GetSecretKey())

	}

	err = app.servers.Update(&server, true)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}

	go func() {
		defer concurrent.Recoverer("SERVERMODIFIED")
		defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
		userID, _ := app.getCurrentUserID(r)

		logEvent := GetSystemLogEvent(userID, "Server Modified", fmt.Sprintf(" %s,IP %s", server.Name, r.RemoteAddr), false)
		logEvent.ImpactedServerId = server.ID
		logEvent.BeforeUpdate = originalServer.LogImage()
		logEvent.AfterUpdate = server.LogImage()
		app.SystemLoggerChan <- logEvent

	}()

	app.sessionManager.Put(r.Context(), "flash", fmt.Sprintf("Server %s updated sucessfully", server.Name))

	http.Redirect(w, r, "/servers", http.StatusSeeOther)
}

// ------------------------------------------------------
// run promotions
// ------------------------------------------------------
func (app *application) SyncUserTokens(w http.ResponseWriter, r *http.Request) {
	if !app.features.AllowTokenSync {
		app.goBack(w, r, http.StatusNotFound)
		return
	}

	defer app.requestMutex.Unlock()
	// inProgress := app.requestMutex.TryLock()
	// if inProgress {
	// 	//log.Println("ServerDeleteConfirm  002 >>>>>>", err.Error())
	// 	app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Please wait. User token sync is already in progress"))
	// 	app.goBack(w, r, http.StatusSeeOther)
	// 	return
	// }

	app.requestMutex.Lock()
	serverID := chi.URLParam(r, "serverid")

	server, err := app.servers.Get(serverID)
	if err != nil {

		//log.Println("ServerDeleteConfirm  002 >>>>>>", err.Error())
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}
	err = app.SyncUserToken(server)

	if err != nil {

		//log.Println("ServerDeleteConfirm  002 >>>>>>", err.Error())
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "User token sync Completed.")

	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
// run promotions
// ------------------------------------------------------
func (app *application) GetLibList(w http.ResponseWriter, r *http.Request) {

	serverID := chi.URLParam(r, "serverid")

	server, err := app.servers.Get(serverID)
	if err != nil {

		//log.Println("ServerDeleteConfirm  002 >>>>>>", err.Error())
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}
	libList, _ := server.GetLibList()

	app.renderAnyWithoutBase(w, r, http.StatusOK, "server_lib_list.tmpl", map[string]any{"liblist": libList})

}

// ------------------------------------------------------
// run promotions
// ------------------------------------------------------
func (app *application) showPromotionTable(w http.ResponseWriter, r *http.Request) {

	serverID := chi.URLParam(r, "serverid")

	server, err := app.servers.Get(serverID)
	if err != nil {

		//log.Println("ServerDeleteConfirm  002 >>>>>>", err.Error())
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	data := app.newTemplateData(r)

	data.Form = server

	data.Server = server
	app.render(w, r, http.StatusOK, "server_crt_promotion_table.tmpl", data)

}

// ------------------------------------------------------
// run promotions
// ------------------------------------------------------
func (app *application) createPromotionTable(w http.ResponseWriter, r *http.Request) {

	serverID := chi.URLParam(r, "serverid")

	server, err := app.servers.Get(serverID)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}
	err = server.CreatePromotionTable(context.TODO())
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))

	} else {

		app.sessionManager.Put(r.Context(), "flash", "done")
	}

	app.goBack(w, r, http.StatusSeeOther)

}
