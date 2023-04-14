package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/onlysumitg/GoQhttp/internal/models"
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

	router.Route("/apilogs", func(r chi.Router) {
		// CSRF
		r.Use(app.RequireAuthentication)
		r.Use(noSurf)
		r.Get("/", app.apilogs)
		r.Get("/{logid}", app.apilogs)
		r.Post("/", app.apilogs)

		logGroup := r.Group(nil)
		logGroup.Use(app.RequireSuperAdmin)
		logGroup.Get("/clear", app.clearapilogs)

	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) apilogs(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	objectid := strings.TrimSpace(r.PostForm.Get("objectid"))
	logid := chi.URLParam(r, "logid")

	if objectid == "" {
		objectid = logid
	}
	logEntries := make([]string, 0)
	if objectid != "" {
		//logEntries = models.GetLogs(app.DB, objectid)
	}

	data := app.newTemplateData(r)
	data.LogEntries = logEntries

	app.render(w, r, http.StatusOK, "api_logs.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------

func (app *application) InjectClientInfo(r *http.Request, requesyBodyFlatMap map[string]xmlutils.ValueDatatype) {
	requesyBodyFlatMap["*CLIENT_IP"] = xmlutils.ValueDatatype{r.RemoteAddr, "STRING"}

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

	return endpointName, pathParams
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) GET(w http.ResponseWriter, r *http.Request) {

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
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
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

	endpointName, pathParams := app.GetPathParameters(r)

	requestBodyMap := make(map[string]any)

	// //need to handle xml body

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&requestBodyMap)
	switch {
	case err == io.EOF:
		// empty body
	case err != nil:
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	requestBodyFlatMap := jsonutils.JsonToFlatMapFromMap(requestBodyMap)

	for _, p := range pathParams {
		requestBodyFlatMap[p.Name] = xmlutils.ValueDatatype{p.Value, "STRING"}
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

	userid := r.Context().Value(models.ContextUserKey).(string)
	user, err := app.users.Get(userid)
	if err != nil {
		app.ErrorJSON(w, r, http.StatusUnauthorized)
		return
	}
	serverID := user.ServerId

	server, err := app.servers.Get(serverID)
	if err != nil {
		app.ErrorJSON(w, r, http.StatusNotImplemented)
		return
	}

	app.InjectClientInfo(r, requesyBodyFlatMap)
	endPoint, err := app.GetEndPoint(fmt.Sprintf("%s_%s", strings.ToUpper(endpointName), strings.ToUpper(r.Method)))
	if err != nil {
		app.errorResponse(w, r, 404, err.Error())
		return
	}

	apiCall := &models.ApiCall{
		ID: uuid.NewString(),

		RequestFlatMap: requesyBodyFlatMap,
		RequestHeader:  httputils.GetHeadersAsMap(r),

		StatusCode: http.StatusOK,

		Log:         make([]string, 0),
		DB:          app.DB,
		HttpRequest: r,

		CurrentSP: endPoint,
	}

	// apiCall.ResponseString = html.UnescapeString(endPoint.ResponsePlaceholder) //string(jsonByte)

	endPoint.APICall(*server, apiCall)

	apiCall.Finalize()

	// // JSON or XML ===> TODO
	// //app.writeJSON(w, apiCall.ResponseCode, apiCall.Response, apiCall.GetHttpHeader())
	// //app.writeJSON(w, apiCall.ResponseCode, apiCall.Response, apiCall.GetHttpHeader())

	app.writeJSON(w, apiCall.StatusCode, apiCall.Response, nil)

	//go apiCall.SaveLogs()

}

// ------------------------------------------------------
//
// ------------------------------------------------------

func (app *application) clearapilogs(w http.ResponseWriter, r *http.Request) {
	//models.ClearLogs(app.DB)
	app.sessionManager.Put(r.Context(), "flash", "Api logs has been cleared")

	app.goBack(w, r, http.StatusSeeOther)
}
