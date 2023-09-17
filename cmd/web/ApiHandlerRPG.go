 package main

// import (
// 	"fmt"
// 	"net/http"
// 	"runtime/debug"
// 	"strings"

// 	"github.com/go-chi/chi/v5/middleware"
// 	"github.com/onlysumitg/GoQhttp/internal/ibmiServer"
// 	"github.com/onlysumitg/GoQhttp/internal/models"
// 	"github.com/onlysumitg/GoQhttp/internal/rpg"
// 	"github.com/onlysumitg/GoQhttp/internal/storedProc"
// 	"github.com/onlysumitg/GoQhttp/logger"
// 	"github.com/onlysumitg/GoQhttp/utils/concurrent"
// 	"github.com/onlysumitg/GoQhttp/utils/httputils"
// 	"github.com/onlysumitg/GoQhttp/utils/xmlutils"
// )

// // ------------------------------------------------------
// //
// //	actual api call processing
// //
// // ------------------------------------------------------
// func (app *application) ProcessRPGAPICall(w http.ResponseWriter, r *http.Request, namespace string, endpointName string,
// 	pathParams []httputils.PathParam,
// 	requesyBodyFlatMap map[string]xmlutils.ValueDatatype) {

// 	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

// 	rpgPgm, sp, endPointNotfoundError := app.GetRPGEndPoint(namespace, endpointName, r.Method)

// 	requestId := middleware.GetReqID(r.Context())

// 	response := &storedProc.StoredProcResponse{ReferenceId: requestId}
// 	//log.Printf("%v: %v\n", "SeversCall001", time.Now())
// 	apiCall := &models.ApiCall{
// 		ID: requestId,

// 		RequestFlatMap: requesyBodyFlatMap,
// 		RequestHeader:  httputils.GetHeadersAsMap(r),

// 		StatusCode: http.StatusOK,

// 		Log:         make([]*logger.LogEvent, 0, 10),
// 		LogDB:       app.LogDB,
// 		HttpRequest: r,

// 		Response: response,
// 	}

// 	// log api data
// 	defer func() {
// 		go apiCall.SaveLogs(app.debugMode) //goroutine
// 	}()

// 	apiCall.Logger("INFO", fmt.Sprintf("Received call for EndPoint %s | Method %s", endpointName, strings.ToUpper(r.Method)), false)

// 	if endPointNotfoundError != nil {
// 		apiCall.Logger("INFO", fmt.Sprintf("%s endpoint %s/%s not found", r.Method, namespace, endpointName), false)
// 		response.Status = http.StatusNotImplemented
// 		response.Message = endPointNotfoundError.Error()
// 		app.writeJSONAPI(w, response, nil)
// 		return

// 	}

// 	graphStruc := GetGraphStruct(r.Context())
// 	graphStruc.Spid = rpgPgm.ID
// 	graphStruc.SpName = rpgPgm.EndPointName
// 	graphStruc.SpUrl = fmt.Sprintf("/sp/%s", rpgPgm.ID)

// 	user, found := app.getCurrentUser(r)

// 	if !found && !rpgPgm.AllowWithoutAuth {
// 		apiCall.Logger("INFO", "Unauthoerized user", false)

// 		response.Status = http.StatusUnauthorized
// 		response.Message = http.StatusText(http.StatusUnauthorized)
// 		app.writeJSONAPI(w, response, nil)
// 		return
// 	}
// 	if found {
// 		apiCall.Logger("INFO", fmt.Sprintf("Request user %s %s", user.Name, user.Email), false)
// 	} else {
// 		apiCall.Logger("INFO", "Processing request without Auth", false)
// 	}

// 	server, level := app.getServerToUseRPG(rpgPgm, user)
// 	if server == nil || level == 0 {

// 		apiCall.Logger("ERROR", "Could not find Server to use", false)

// 		response.Status = http.StatusNotImplemented
// 		response.Message = "Please check assigned server to the user"
// 		app.writeJSONAPI(w, response, nil)
// 		return

// 	}

// 	apiCall.Logger("INFO", fmt.Sprintf("Server assigned %s@%s", server.GetUserName(), server.Name), false)

// 	app.InjectRequestInfo(r, requesyBodyFlatMap)
// 	app.InjectServerInfo(server, requesyBodyFlatMap)

// 	//log.Printf("%v: %v\n", "SeversCall005", time.Now())

// 	// set remaining values
// 	apiCall.CurrentRPG = rpgPgm
// 	apiCall.Server = server

// 	// apiCall.ResponseString = html.UnescapeString(endPoint.ResponsePlaceholder) //string(jsonByte)

// 	apiCall.Logger("INFO", fmt.Sprintf("Calling RPG %s   on server %s", apiCall.CurrentRPG.Name, server.Name), false)

// 	// call the SP
// 	apiCall.Response, apiCall.SPCallDuration, apiCall.Err = server.RPGAPICall(r.Context(), apiCall.ID, sp, rpgPgm, apiCall.RequestFlatMap, app.GetParamValidatorRegex())
// 	//log.Printf("%v: %v\n", "SeversCall006", time.Now())

// 	if apiCall.Err == nil {
// 		go func() {
// 			concurrent.Recoverer("AddServerLastCall")
// 			app.AddServerLastCall(server.ID)
// 		}()
// 	}

// 	graphStruc.SPResponsetime = apiCall.SPCallDuration.Milliseconds()

// 	apiCall.Logger("INFO", "Finalizing response", false)

// 	apiCall.Finalize()

// 	// // JSON or XML ===> TODO
// 	// //app.writeJSON(w, apiCall.ResponseCode, apiCall.Response, apiCall.GetHttpHeader())
// 	// //app.writeJSON(w, apiCall.ResponseCode, apiCall.Response, apiCall.GetHttpHeader())

// 	if app.allowHtmlTemplates() && rpgPgm.HtmlTemplate != "" {
// 		templateData := map[string]any{
// 			"response": apiCall.Response,
// 			"request":  apiCall.RequestFlatMap,
// 		}

// 		app.spRender(w, r, apiCall.StatusCode, rpgPgm.HtmlTemplate, templateData, apiCall.BuildHeaders()) //apiCall.Response)
// 	} else {

// 		app.writeJSON(w, apiCall.StatusCode, apiCall.Response, apiCall.BuildHeaders())
// 	}

// 	// save SP logid
// 	//goroutine
// 	go func() {

// 		defer concurrent.Recoverer("Recovered in AddLogid")
// 		defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

// 		l := models.SPCallLogEntry{SpID: apiCall.CurrentRPG.ID, LogId: apiCall.ID}
// 		app.spCallLogModel.DataChan <- l
// 	}()

// }

// // ------------------------------------------------------
// //
// // ------------------------------------------------------

// func (app *application) getServerToUseRPG(endPoint *rpg.Program, user *models.User) (*ibmiServer.Server, int) {

// 	//var userServer *ibmiServer.Server = nil
// 	var endPointServer *ibmiServer.Server = nil

// 	endPointServer, err1 := app.servers.Get("e2f40341-3ec8-48db-a2f6-f28e7bf515c7")
// 	if err1 != nil {
// 		endPointServer = nil
// 	}

// 	return endPointServer, 2
// 	// if user != nil {
// 	// 	userServer2, err2 := app.servers.Get(user.ServerId)
// 	// 	if err2 != nil {
// 	// 		userServer = nil
// 	// 	} else {
// 	// 		userServer = userServer2
// 	// 	}
// 	// }

// 	// if userServer != nil && endPoint.IsAllowedForServer(userServer.ID) {

// 	// 	return userServer, 1 // 1= user server

// 	// }

// 	// // allow endpoint server only for unauth users

// 	// if endPoint.AllowWithoutAuth {

// 	// 	if endPointServer != nil && endPoint.IsAllowedForServer(endPointServer.ID) {

// 	// 		return endPointServer, 2 // 2= endpoint server

// 	// 	}
// 	// }

// 	// return nil, 0

// }
