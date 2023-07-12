package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/onlysumitg/GoQhttp/internal/dbserver"
	"github.com/onlysumitg/GoQhttp/internal/models"
	"github.com/onlysumitg/GoQhttp/internal/storedProc"
	"github.com/onlysumitg/GoQhttp/logger"
	"github.com/onlysumitg/GoQhttp/utils/concurrent"
	"github.com/onlysumitg/GoQhttp/utils/httputils"
	"github.com/onlysumitg/GoQhttp/utils/jsonutils"
	"github.com/onlysumitg/GoQhttp/utils/xmlutils"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) APIHandlers(router *chi.Mux) {
	router.Route("/api/{apiname}", func(r chi.Router) {
		//r.With(paginate).Get("/", listArticles)
		// Log response time
		r.Use(app.TimeTook)
		r.Use(app.LogHandler)
		r.Use(app.RequireTokenAuthentication)

		r.Get("/", app.GET)
		r.Post("/", app.POST)
		r.Put("/", app.POST)
		r.Delete("/", app.POST)

		r.Get("/*", app.GET)
		r.Post("/*", app.POST)
		r.Put("/*", app.POST)
		r.Delete("/*", app.POST)
	})

	// for unauthorized end points
	router.Route("/uapi/{apiname}", func(r chi.Router) {
		//r.With(paginate).Get("/", listArticles)
		r.Use(app.TimeTook)
		r.Use(app.LogHandler)
		r.Use(app.RequireUnAuthEndPoint)
		r.Get("/", app.GET)
		r.Post("/", app.POST)
		r.Put("/", app.POST)
		r.Delete("/", app.POST)

		r.Get("/*", app.GET)
		r.Post("/*", app.POST)
		r.Put("/*", app.POST)
		r.Delete("/*", app.POST)
	})

}

// ------------------------------------------------------
//
//	middleware
//
// ------------------------------------------------------
func (app *application) RequireUnAuthEndPoint(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := &storedProc.StoredProcResponse{ReferenceId: middleware.GetReqID(r.Context())}

		endpointName, _ := app.GetPathParameters(r)
		endPoint, err := app.GetEndPoint(fmt.Sprintf("%s_%s", strings.ToUpper(endpointName), strings.ToUpper(r.Method)))

		if err != nil || !endPoint.AllowWithoutAuth {
			response.Status = http.StatusNotFound
			response.Message = http.StatusText(http.StatusNotFound)
			app.writeJSONAPI(w, response, nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ------------------------------------------------------
//
// ------------------------------------------------------

func (app *application) InjectRequestInfo(r *http.Request, requesyBodyFlatMap map[string]xmlutils.ValueDatatype) {
	requesyBodyFlatMap["QHTTP_CLIENT_IP"] = xmlutils.ValueDatatype{Value: strings.TrimSpace(r.RemoteAddr), DataType: "STRING"}
	requesyBodyFlatMap["QHTTP_METHOD"] = xmlutils.ValueDatatype{Value: r.Method, DataType: "STRING"}

	user, ok := r.Context().Value(models.ContextUserName).(string)
	if ok {
		requesyBodyFlatMap["QHTTP_USER"] = xmlutils.ValueDatatype{Value: user, DataType: "STRING"}

	} else {
		requesyBodyFlatMap["QHTTP_USER"] = xmlutils.ValueDatatype{Value: "ANONYMOUS", DataType: "STRING"}

	}

	currentUser, found := app.getCurrentUser(r)
	if found {
		requesyBodyFlatMap["QHTTP_USER_TOKEN"] = xmlutils.ValueDatatype{Value: currentUser.Token, DataType: "STRING"}
		requesyBodyFlatMap["QHTTP_USER_EMAIL"] = xmlutils.ValueDatatype{Value: currentUser.Email, DataType: "STRING"}

	} else {
		requesyBodyFlatMap["QHTTP_USER_TOKEN"] = xmlutils.ValueDatatype{Value: "", DataType: "STRING"}
		requesyBodyFlatMap["QHTTP_USER_EMAIL"] = xmlutils.ValueDatatype{Value: "", DataType: "STRING"}

	}

	requestId := middleware.GetReqID(r.Context())
	if requestId != "" {
		requesyBodyFlatMap["QHTTP_CORRELATION_ID"] = xmlutils.ValueDatatype{Value: requestId, DataType: "STRING"}

	} else {
		requesyBodyFlatMap["QHTTP_CORRELATION_ID"] = xmlutils.ValueDatatype{Value: "", DataType: "STRING"}

	}
}

// ------------------------------------------------------
//
// ------------------------------------------------------

func (app *application) InjectServerInfo(server *dbserver.Server, requesyBodyFlatMap map[string]xmlutils.ValueDatatype) {
	requesyBodyFlatMap["QHTTP_SERVER"] = xmlutils.ValueDatatype{Value: server.Name, DataType: "STRING"}
	requesyBodyFlatMap["QHTTP_SERVER_USER"] = xmlutils.ValueDatatype{Value: server.UserName, DataType: "STRING"}

}

// ------------------------------------------------------
//
// ------------------------------------------------------

func (app *application) GetPathParameters(r *http.Request) (string, []httputils.PathParam) {

	endpointName := ""
	pathParams := make([]httputils.PathParam, 0)

	params, err := httputils.GetPathParamMap(r.URL.Path, "")
	if err == nil {
		for i, p := range params {
			switch i {
			case 0:
				fmt.Println("")
			case 1:
				endpointName = p.Value.(string)

			default:
				p.Name = fmt.Sprintf("*PATH_%d", i-2)
				pathParams = append(pathParams, *p)

			}
		}
	}

	return strings.TrimSpace(endpointName), pathParams
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) GET(w http.ResponseWriter, r *http.Request) {
	response := &storedProc.StoredProcResponse{ReferenceId: middleware.GetReqID(r.Context())}

	endpointName, pathParams := app.GetPathParameters(r)
	queryString := fmt.Sprint(r.URL)
	//apiName := chi.URLParam(r, "apiname")

	// apiName, err := httputils.QueryParamPath(queryString, "/api/")
	// if err != nil {
	// 	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
	// 	return
	// }

	requestJson, err := httputils.QueryParamToMap(queryString)
	if err != nil {
		response.Status = http.StatusBadRequest
		response.Message = err.Error()
		app.writeJSONAPI(w, response, nil)
		return

	}

	for _, p := range pathParams {
		requestJson[p.Name] = p.Value
	}

	requestBodyFlatMap := jsonutils.JsonToFlatMapFromMap(requestJson)

	app.ProcessAPICall(w, r, endpointName, pathParams, requestBodyFlatMap)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) POST(w http.ResponseWriter, r *http.Request) {

	response := &storedProc.StoredProcResponse{ReferenceId: middleware.GetReqID(r.Context())}

	endpointName, pathParams := app.GetPathParameters(r)

	requestBodyMap := make(map[string]any)

	// //need to handle xml body

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestBodyMap)
	switch {
	case err == io.EOF:
		// empty body
	case err != nil:
		response.Status = http.StatusBadRequest
		response.Message = err.Error()
		app.writeJSONAPI(w, response, nil)
		return

	}

	requestBodyFlatMap := jsonutils.JsonToFlatMapFromMap(requestBodyMap)

	for _, p := range pathParams {
		requestBodyFlatMap[p.Name] = xmlutils.ValueDatatype{Value: p.Value, DataType: "STRING"}
	}

	app.ProcessAPICall(w, r, endpointName, pathParams, requestBodyFlatMap)

}

// ------------------------------------------------------
//
//	actual api call processing
//
// ------------------------------------------------------
func (app *application) ProcessAPICall(w http.ResponseWriter, r *http.Request, endpointName string,
	pathParams []httputils.PathParam,
	requesyBodyFlatMap map[string]xmlutils.ValueDatatype) {

	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	requestId := middleware.GetReqID(r.Context())

	response := &storedProc.StoredProcResponse{ReferenceId: requestId}
	//log.Printf("%v: %v\n", "SeversCall001", time.Now())
	apiCall := &models.ApiCall{
		ID: requestId,

		RequestFlatMap: requesyBodyFlatMap,
		RequestHeader:  httputils.GetHeadersAsMap(r),

		StatusCode: http.StatusOK,

		Log:         make([]*logger.LogEvent, 0, 10),
		LogDB:       app.LogDB,
		HttpRequest: r,

		Response: response,
	}

	// log api data
	defer func() {
		go apiCall.SaveLogs(app.debugMode) //goroutine
	}()

	apiCall.Logger("INFO", fmt.Sprintf("Received call for EndPoint %s | Method %s", endpointName, strings.ToUpper(r.Method)), false)
	endPoint, err := app.GetEndPoint(fmt.Sprintf("%s_%s", strings.ToUpper(endpointName), strings.ToUpper(r.Method)))
	if err != nil {
		apiCall.Logger("INFO", fmt.Sprintf("%s endpoint %s not found", r.Method, endpointName), false)

		response.Status = http.StatusNotImplemented
		response.Message = err.Error()
		app.writeJSONAPI(w, response, nil)
		return

	}

	graphStruc := GetGraphStruct(r.Context())
	graphStruc.Spid = endPoint.ID
	graphStruc.SpName = endPoint.EndPointName
	graphStruc.SpUrl = fmt.Sprintf("/sp/%s", endPoint.ID)

	user, found := app.getCurrentUser(r)

	if !found && !endPoint.AllowWithoutAuth {
		apiCall.Logger("INFO", "Unauthoerized user", false)

		response.Status = http.StatusUnauthorized
		response.Message = http.StatusText(http.StatusUnauthorized)
		app.writeJSONAPI(w, response, nil)
		return
	}
	if found {
		apiCall.Logger("INFO", fmt.Sprintf("Request user %s %s", user.Name, user.Email), false)
	} else {
		apiCall.Logger("INFO", "Processing request without Auth", false)
	}

	server, level := app.getServerToUse(endPoint, user)
	if server == nil || level == 0 {

		apiCall.Logger("ERROR", "Could not find Server to use", false)

		response.Status = http.StatusNotImplemented
		response.Message = "Please check assigned server to the user"
		app.writeJSONAPI(w, response, nil)
		return

	}

	apiCall.Logger("INFO", fmt.Sprintf("Server assigned %s@%s", server.UserName, server.Name), false)

	app.InjectRequestInfo(r, requesyBodyFlatMap)
	app.InjectServerInfo(server, requesyBodyFlatMap)

	//log.Printf("%v: %v\n", "SeversCall005", time.Now())

	// set remaining values
	apiCall.CurrentSP = endPoint
	apiCall.Server = server

	// apiCall.ResponseString = html.UnescapeString(endPoint.ResponsePlaceholder) //string(jsonByte)

	apiCall.Logger("INFO", fmt.Sprintf("Calling SP %s (specific %s) on server %s", apiCall.CurrentSP.Name, apiCall.CurrentSP.SpecificName, server.Name), false)

	// call the SP
	apiCall.Response, apiCall.SPCallDuration, apiCall.Err = server.APICall(r.Context(), apiCall.ID, endPoint, apiCall.RequestFlatMap)
	//log.Printf("%v: %v\n", "SeversCall006", time.Now())

	if apiCall.Err == nil {
		go func() {
			concurrent.Recoverer("AddServerLastCall")
			app.AddServerLastCall(server.ID)
		}()
	}

	graphStruc.SPResponsetime = apiCall.SPCallDuration.Milliseconds()

	apiCall.Logger("INFO", "Finalizing response", false)

	apiCall.Finalize()

	// // JSON or XML ===> TODO
	// //app.writeJSON(w, apiCall.ResponseCode, apiCall.Response, apiCall.GetHttpHeader())
	// //app.writeJSON(w, apiCall.ResponseCode, apiCall.Response, apiCall.GetHttpHeader())

	app.writeJSON(w, apiCall.StatusCode, apiCall.Response, nil)

	// save SP logid
	//goroutine
	go func() {

		defer concurrent.Recoverer("Recovered in AddLogid")
		defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

		l := models.SPCallLogEntry{SpID: apiCall.CurrentSP.ID, LogId: apiCall.ID}
		app.spCallLogModel.DataChan <- l
	}()

}

// ------------------------------------------------------
//
// ------------------------------------------------------

func (app *application) getCurrentUser(r *http.Request) (*models.User, bool) {
	userid, ok := r.Context().Value(models.ContextUserKey).(string)
	if !ok {
		return nil, false
	}

	user, err := app.users.Get(userid)
	if err != nil {
		return nil, false
	}

	return user, true
}

// ------------------------------------------------------
//
// ------------------------------------------------------

func (app *application) getServerToUse(endPoint *storedProc.StoredProc, user *models.User) (*dbserver.Server, int) {

	var userServer *dbserver.Server = nil
	var endPointServer *dbserver.Server = nil

	endPointServer, err1 := app.servers.Get(endPoint.DefaultServer.ID)
	if err1 != nil {
		endPointServer = nil
	}

	if user != nil {
		userServer2, err2 := app.servers.Get(user.ServerId)
		if err2 != nil {
			userServer = nil
		} else {
			userServer = userServer2
		}
	}

	if userServer != nil && endPoint.IsAllowedForServer(userServer.ID) {

		return userServer, 1 // 1= user server

	}

	// allow endpoint server only for unauth users

	if endPoint.AllowWithoutAuth {

		if endPointServer != nil && endPoint.IsAllowedForServer(endPointServer.ID) {

			return endPointServer, 2 // 2= endpoint server

		}
	}

	return nil, 0

}
