package models

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/onlysumitg/GoQhttp/go_ibm_db"
	"github.com/onlysumitg/GoQhttp/utils/xmlutils"
	bolt "go.etcd.io/bbolt"
)

var infoLog *log.Logger = log.New(os.Stderr, "INFO \t", log.Ldate|log.Ltime)
var errorLog *log.Logger = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime)
var RequestLog *log.Logger = log.New(os.Stderr, "Request\t", log.Ldate|log.Ltime)
var ResponseLog *log.Logger = log.New(os.Stderr, "Response\t", log.Ldate|log.Ltime)

type ApiCall struct {
	ID string
	//Request        map[string]any
	RequestFlatMap map[string]xmlutils.ValueDatatype
	RequestHeader  map[string]string

	Response      *StoredProcResponse
	Err           error
	StatusCode    int
	StatusMessage string

	Log []string

	logMutex sync.Mutex

	LogDB *bolt.DB

	HttpRequest *http.Request

	CurrentSP *StoredProc

	Server *Server
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (a *ApiCall) HasError() bool {
	if a.Err != nil {
		var odbcError *go_ibm_db.Error

		if errors.Is(a.Err, driver.ErrBadConn) {
			a.StatusCode = http.StatusInternalServerError
			a.StatusMessage = a.Err.Error()
			go a.LogError(fmt.Sprintf("Connection Error: %s", a.Server.Name))

			return true
		}

		if errors.As(a.Err, &odbcError) {
			a.StatusCode, a.StatusMessage = OdbcErrMessage(odbcError)
			go a.LogError(fmt.Sprintf("ODBC Error %s:%s", a.StatusMessage, odbcError.Error()))
			return true
		}
		a.StatusCode = http.StatusBadRequest
		a.StatusMessage = a.Err.Error()
		go a.LogError(fmt.Sprintf("Error %s", a.Err.Error()))

		return true
	}
	return false
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (a *ApiCall) Finalize() {

	if a.HasError() {
		a.Response.Message = a.StatusMessage
		a.Response.Status = a.StatusCode
	}
	a.Response.ReferenceId = a.ID

}

// // ------------------------------------------------------
// //
// // ------------------------------------------------------
// func (a *ApiCall) HTTPCall() *httputils.HttpCallResult {

// 	finalUrlToUse, err := a.BuildUrlTOUse()
// 	if err != nil {
// 		return nil
// 	}

// 	a.ActualUrlToUse = finalUrlToUse

// 	if a.CurrentEndPoint.isMethod("GET") {
// 		return a.GETCall(finalUrlToUse)
// 	}
// 	if a.CurrentEndPoint.isMethod("POST") {
// 		return a.POSTCall(finalUrlToUse)
// 	}
// 	if a.CurrentEndPoint.isMethod("PUT") {
// 		return a.PUTCall(finalUrlToUse)
// 	}
// 	if a.CurrentEndPoint.isMethod("DELETE") {
// 		return a.DELETECall(finalUrlToUse)
// 	}

// 	return nil
// }

// // ------------------------------------------------------
// //
// // ------------------------------------------------------
// func (a *ApiCall) BuildUrlTOUse() (string, error) {

// 	host, found := a.CurrentEndPoint.ParsedUrl["Host"]

// 	if !found || host == "" {
// 		a.LogInfo("Skipping as Host var not defined")
// 		return "", errors.New("host not defined")
// 	}

// 	http_https, found := a.CurrentEndPoint.ParsedUrl["Scheme"]
// 	if !found || http_https == "" {
// 		http_https = "http"
// 	}

// 	// build base URL
// 	baseUrl := fmt.Sprintf("%s://%s", http_https, host)

// 	a.LogInfo(fmt.Sprintf("Using base URL %s", baseUrl))

// 	// get path params

// 	pathParms := ""

// 	for i, p := range a.CurrentEndPoint.PathParams {

// 		if p.IsVariable && len(a.PathParams) >= i+1 {
// 			pathParms = pathParms + "/" + a.PathParams[i].StringValue
// 		} else {
// 			pathParms = pathParms + "/" + p.StringValue
// 		}

// 	}

// 	if pathParms != "" {
// 		baseUrl = baseUrl + pathParms
// 	}

// 	// get query string
// 	queryString := a.HttpRequest.URL.RawQuery
// 	if queryString != "" {
// 		baseUrl = baseUrl + "?" + queryString
// 	}
// 	a.LogInfo(fmt.Sprintf("Final URL to use %s", baseUrl))
// 	return baseUrl, nil
// }

// // ------------------------------------------------------
// //
// // ------------------------------------------------------
// func (a *ApiCall) GETCall(urlToUse string) *httputils.HttpCallResult {
// 	var httpCallResult *httputils.HttpCallResult = nil
// 	if a.CurrentEndPoint.isMethod("GET") {
// 		httpCallResult = httputils.HttpGET(urlToUse, a.HttpRequest.Header)
// 	}
// 	return httpCallResult
// }

// // ------------------------------------------------------
// //
// // ------------------------------------------------------
// func (a *ApiCall) POSTCall(urlToUse string) *httputils.HttpCallResult {
// 	var httpCallResult *httputils.HttpCallResult = nil

// 	if a.CurrentEndPoint.isMethod("POST") {
// 		body, err := ioutil.ReadAll(a.HttpRequest.Body)
// 		if err != nil {
// 			body = []byte("")
// 		}

// 		httpCallResult = httputils.HttpPOST(urlToUse, a.HttpRequest.Header, body)
// 	}
// 	return httpCallResult
// }

// // ------------------------------------------------------
// //
// // ------------------------------------------------------
// func (a *ApiCall) PUTCall(urlToUse string) *httputils.HttpCallResult {
// 	var httpCallResult *httputils.HttpCallResult = nil

// 	if a.CurrentEndPoint.isMethod("PUT") {
// 		body, err := ioutil.ReadAll(a.HttpRequest.Body)
// 		if err != nil {
// 			body = []byte("")
// 		}

// 		httpCallResult = httputils.HttpPUT(urlToUse, a.HttpRequest.Header, body)
// 	}
// 	return httpCallResult
// }

// // ------------------------------------------------------
// //
// // ------------------------------------------------------
// func (a *ApiCall) DELETECall(urlToUse string) *httputils.HttpCallResult {
// 	var httpCallResult *httputils.HttpCallResult = nil

// 	if a.CurrentEndPoint.isMethod("DELETE") {
// 		body, err := ioutil.ReadAll(a.HttpRequest.Body)
// 		if err != nil {
// 			body = []byte("")
// 		}

// 		httpCallResult = httputils.HttpDELETE(urlToUse, a.HttpRequest.Header, body)
// 	}
// 	return httpCallResult
// }

// // ------------------------------------------------------
// //
// // ------------------------------------------------------
// func (apiCall *ApiCall) GetHttpHeader() http.Header {
// 	var header http.Header = make(http.Header)

// 	header["CORRELATIONID"] = []string{apiCall.ID}
// 	for key, value := range apiCall.ResponseHeader {
// 		header[key] = []string{value}
// 	}

// 	delete(header, "Content-Length")

// 	return header
// }

// // ------------------------------------------------------
// //
// // ------------------------------------------------------
// func (apiCall *ApiCall) HasSet(keyToCheck string) bool {
// 	hasSet := false
// 	for _, key := range apiCall.HasSetCache {
// 		if strings.EqualFold(key, keyToCheck) {
// 			hasSet = true
// 			break
// 		}

// 	}

// 	return hasSet
// }

// // ------------------------------------------------------
// //
// // ------------------------------------------------------
// func (apiCall *ApiCall) SetKey(keyToCheck string) {
// 	apiCall.HasSetCache = append(apiCall.HasSetCache, keyToCheck)
// }

// ------------------------------------------------------
//
// ------------------------------------------------------
func getLogTableName() []byte {
	return []byte("apilogs")
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (apiCall *ApiCall) LogInfo(logEntry string) {

	defer apiCall.logMutex.Unlock()
	apiCall.logMutex.Lock()

	buf := bytes.NewBufferString("")

	infoLog.SetOutput(buf)
	infoLog.Println(logEntry)

	apiCall.Log = append(apiCall.Log, buf.String())

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (apiCall *ApiCall) LogError(logEntry string) {
	defer apiCall.logMutex.Unlock()

	buf := bytes.NewBufferString("")

	apiCall.logMutex.Lock()
	errorLog.SetOutput(buf)
	errorLog.Println(logEntry)
	apiCall.Log = append(apiCall.Log, buf.String())

}

// // ------------------------------------------------------
// //
// // ------------------------------------------------------
// // func (l *ErrorLogger) Error(err error) {
// // 	// Get the stack trace as a string
// // 	//buf := new(bytes.Buffer)
// // 	//l.logger.withStack(buf, err)

// // 	//sendErrorMail(buf.String())

// //		buf := bytes.NewBufferString(s)
// //	}
// //
// ------------------------------------------------------
//
// ------------------------------------------------------
func (m *ApiCall) SaveLogs() {
	m.LogDB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(getLogTableName())
		if err != nil {
			return err
		}

		for i, s := range m.Log {

			key := fmt.Sprintf("%s_%d", m.ID, i)
			bucket.Put([]byte(key), []byte(fmt.Sprintf("%05d. %s", i+1, s)))

		}
		return nil
	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func SaveLogs(db *bolt.DB, i int, id string, message string) {
	db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(getLogTableName())
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%s_%d", id, i)
		bucket.Put([]byte(key), []byte(fmt.Sprintf("%05d. %s", i+1, message)))

		return nil
	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func GetLogs(db *bolt.DB, id string) []string {

	l := make([]string, 0)

	_ = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(getLogTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}

		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if bytes.HasPrefix(k, []byte(id)) {
				l = append(l, string(v))
			}
		}

		return nil
	})

	return l

}
