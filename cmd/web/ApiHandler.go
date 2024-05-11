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
	"github.com/zerobit-tech/GoQhttp/internal/ibmiServer"
	"github.com/zerobit-tech/GoQhttp/internal/models"
	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
	"github.com/zerobit-tech/GoQhttp/logger"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
	"github.com/zerobit-tech/GoQhttp/utils/httputils"
	"github.com/zerobit-tech/GoQhttp/utils/jsonutils"
	"github.com/zerobit-tech/GoQhttp/utils/xmlutils"
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
		r.Delete("/", app.GET)
		r.Patch("/", app.POST)

		r.Get("/*", app.GET)
		r.Post("/*", app.POST)
		r.Put("/*", app.POST)
		r.Delete("/*", app.GET)
		r.Patch("/*", app.POST)
	})

	// only allowed from html templats
	if app.allowHtmlTemplates() {
		// endpoints with HTML templates
		router.Route("/tapi/{apiname}", func(r chi.Router) {
			r.Use(app.TimeTook)
			r.Use(app.sessionManager.LoadAndSave)
			r.Use(app.LogHandler)
			r.Use(app.RequireAuthenticationForTemplatedAPI)

			r.Use(app.RequireTokenOrSessionAuthentication) // todo check if need both

			//	r.Use(app.RequireTemplatedEndPoint)

			r.Get("/", app.GET)
			r.Post("/", app.POST)
			r.Put("/", app.POST)
			r.Delete("/", app.GET)
			r.Patch("/", app.POST)

			r.Get("/*", app.GET)
			r.Post("/*", app.POST)
			r.Put("/*", app.POST)
			r.Delete("/*", app.GET)
			r.Patch("/*", app.POST)
		})
	}

	// for unauthorized end points
	router.Route("/uapi/{apiname}", func(r chi.Router) {
		//r.With(paginate).Get("/", listArticles)
		r.Use(app.TimeTook)
		r.Use(app.LogHandler)
		r.Use(app.RequireUnAuthEndPoint)

		r.Get("/", app.GET)
		r.Post("/", app.POST)
		r.Put("/", app.POST)
		r.Delete("/", app.GET)
		r.Patch("/", app.POST)

		r.Get("/*", app.GET)
		r.Post("/*", app.POST)
		r.Put("/*", app.POST)
		r.Delete("/*", app.GET)
		r.Patch("/*", app.POST)

	})

}

// ------------------------------------------------------
//
//	middleware
//
// ------------------------------------------------------
func (app *application) RequireUnAuthEndPoint(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// response := &storedProc.StoredProcResponse{ReferenceId: middleware.GetReqID(r.Context())}

		// namespace, endpointName, _ := app.GetPathParameters(r)
		// endPoint, err := app.GetEndPoint(namespace, endpointName, r.Method)

		// if err != nil || !endPoint.AllowWithoutAuth {
		// 	response.Status = http.StatusNotFound
		// 	response.Message = http.StatusText(http.StatusNotFound)
		// 	app.writeJSONAPI(w, response, nil)
		// 	return
		// }

		next.ServeHTTP(w, r)
	})
}

// ------------------------------------------------------
//
//	middleware
//
// ------------------------------------------------------
func (app *application) RequireTemplatedEndPoint(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := &storedProc.StoredProcResponse{ReferenceId: middleware.GetReqID(r.Context())}

		namespace, endpointName, _ := app.GetPathParameters(r)
		endPoint, err := app.GetEndPoint(namespace, endpointName, r.Method)

		if err != nil || endPoint.HtmlTemplate == "" {
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

	namespace, endpointName, _ := app.GetPathParameters(r)
	requesyBodyFlatMap["QHTTP_ENDPOINT_NAMESPACE"] = xmlutils.ValueDatatype{Value: namespace, DataType: "STRING"}
	requesyBodyFlatMap["QHTTP_ENDPOINT_NAME"] = xmlutils.ValueDatatype{Value: endpointName, DataType: "STRING"}

}

// ------------------------------------------------------
//
// ------------------------------------------------------

func (app *application) InjectServerInfo(server *ibmiServer.Server, requesyBodyFlatMap map[string]xmlutils.ValueDatatype) {
	requesyBodyFlatMap["QHTTP_SERVER"] = xmlutils.ValueDatatype{Value: server.Name, DataType: "STRING"}
	requesyBodyFlatMap["QHTTP_SERVER_USER"] = xmlutils.ValueDatatype{Value: server.GetUserName(), DataType: "STRING"}

}

// ------------------------------------------------------
//
// ------------------------------------------------------

func (app *application) GetPathParameters(r *http.Request) (string, string, []httputils.PathParam) {
	namespace := ""
	endpointName := ""
	pathParams := make([]httputils.PathParam, 0)

	params, err := httputils.GetPathParamMap(r.URL.Path, "")
	if err == nil {
		for i, p := range params {
			switch i {
			case 0:
				// do nothing
			case 1:
				namespace = p.Value.(string)
			case 2:
				endpointName = p.Value.(string)

			default:
				p.Name = fmt.Sprintf("*PATH_%d", i-3)
				pathParams = append(pathParams, *p)

			}
		}
	}

	return strings.TrimSpace(namespace), strings.TrimSpace(endpointName), pathParams
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) GET(w http.ResponseWriter, r *http.Request) {
	response := &storedProc.StoredProcResponse{ReferenceId: middleware.GetReqID(r.Context())}

	namespace, endpointName, pathParams := app.GetPathParameters(r)
	//apiName := chi.URLParam(r, "apiname")

	// apiName, err := httputils.QueryParamPath(queryString, "/api/")
	// if err != nil {
	// 	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
	// 	return
	// }

	queryString := fmt.Sprint(r.URL)
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

	app.ProcessAPICall(w, r, namespace, endpointName, pathParams, requestBodyFlatMap)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) POST(w http.ResponseWriter, r *http.Request) {

	response := &storedProc.StoredProcResponse{ReferenceId: middleware.GetReqID(r.Context())}

	namespace, endpointName, pathParams := app.GetPathParameters(r)

	requestBodyMap := make(map[string]any)

	queryString := fmt.Sprint(r.URL)
	queryParamJson, err := httputils.QueryParamToMap(queryString)
	if err != nil {
		response.Status = http.StatusBadRequest
		response.Message = err.Error()
		app.writeJSONAPI(w, response, nil)
		return

	}

	formData := false

	//parse form data for html templates
	if app.allowHtmlTemplates() && httputils.HasFormData(r) {
		formMap, err := httputils.FormToJson(r)
		if err == nil {
			formData = true
			requestBodyMap = formMap
		}

	}

	if !formData {

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

	}

	queryParameters := jsonutils.JsonToFlatMapFromMap(queryParamJson)

	bodyParameters := jsonutils.JsonToFlatMapFromMap(requestBodyMap)

	// body param can override query param
	for k, v := range bodyParameters {
		queryParameters[k] = v
	}

	// path param can override query and body
	for _, p := range pathParams {
		queryParameters[p.Name] = xmlutils.ValueDatatype{Value: p.Value, DataType: "STRING"}
	}

	app.ProcessAPICall(w, r, namespace, endpointName, pathParams, queryParameters)

}

// ------------------------------------------------------
//
//	actual api call processing
//
// ------------------------------------------------------
func (app *application) CheckAPIType(r *http.Request, namespace string, endpointName string) string {

	endPointType := "SQLSP"

	_, endPointNotfoundError := app.GetEndPoint(namespace, endpointName, r.Method)
	if endPointNotfoundError != nil {
		_, rpgErr := app.GetRPGEndPoint(namespace, endpointName, r.Method)
		if rpgErr == nil {
			endPointType = "PGM"
		}
	}

	return endPointType
}

// ------------------------------------------------------
//
//	actual api call processing
//
// ------------------------------------------------------
func (app *application) ProcessAPICall(w http.ResponseWriter, r *http.Request, namespace string, endpointName string,
	pathParams []httputils.PathParam,
	requesyBodyFlatMap map[string]xmlutils.ValueDatatype) {

	endPointType := app.CheckAPIType(r, namespace, endpointName)

	switch endPointType {
	case "PGM":
		app.ProcessRPGAPICall(w, r, namespace, endpointName, pathParams, requesyBodyFlatMap)
	default:
		app.ProcessSQLSPAPICall(w, r, namespace, endpointName, pathParams, requesyBodyFlatMap)
	}

}

// ------------------------------------------------------
//
//	actual api call processing
//
// ------------------------------------------------------
func (app *application) ProcessSQLSPAPICall(w http.ResponseWriter, r *http.Request, namespace string, endpointName string,
	pathParams []httputils.PathParam,
	requesyBodyFlatMap map[string]xmlutils.ValueDatatype) {

	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	endPoint, endPointNotfoundError := app.GetEndPoint(namespace, endpointName, r.Method)

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

	if endPointNotfoundError != nil {
		apiCall.Logger("INFO", fmt.Sprintf("%s endpoint %s/%s not found", r.Method, namespace, endpointName), false)
		response.Status = http.StatusNotImplemented
		response.Message = endPointNotfoundError.Error()
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

	apiCall.Logger("INFO", fmt.Sprintf("Server assigned %s@%s", server.GetUserName(), server.Name), false)

	app.InjectRequestInfo(r, requesyBodyFlatMap)
	app.InjectServerInfo(server, requesyBodyFlatMap)

	//log.Printf("%v: %v\n", "SeversCall005", time.Now())

	// set remaining values
	apiCall.CurrentSP = endPoint
	apiCall.Server = server

	// apiCall.ResponseString = html.UnescapeString(endPoint.ResponsePlaceholder) //string(jsonByte)

	apiCall.Logger("INFO", fmt.Sprintf("Calling SP %s (specific %s) on server %s", apiCall.CurrentSP.Name, apiCall.CurrentSP.SpecificName, server.Name), false)

	// call the SP
	apiCall.Response, apiCall.SPCallDuration, apiCall.Err = server.APICall(r.Context(), apiCall.ID, endPoint, apiCall.RequestFlatMap, app.GetParamValidatorRegex())
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

	if app.allowHtmlTemplates() && endPoint.HtmlTemplate != "" {
		templateData := map[string]any{
			"response": apiCall.Response,
			"request":  apiCall.RequestFlatMap,
		}

		app.spRender(w, r, apiCall.StatusCode, endPoint.HtmlTemplate, templateData, apiCall.BuildHeaders()) //apiCall.Response)
	} else {

		app.writeJSON(w, apiCall.StatusCode, apiCall.Response, apiCall.BuildHeaders())
	}

	// save SP logid
	//goroutine
	go func() {

		defer concurrent.Recoverer("Recovered in AddLogid")
		defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

		l := models.SPCallLogEntry{EndPoint: apiCall.CurrentSP, LogId: apiCall.ID}
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

func (app *application) getServerToUse(endPoint *storedProc.StoredProc, user *models.User) (*ibmiServer.Server, int) {

	var userServer *ibmiServer.Server = nil
	var endPointServer *ibmiServer.Server = nil

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
