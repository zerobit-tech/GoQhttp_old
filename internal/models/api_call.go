package models

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/zerobit-tech/GoQhttp/internal/ibmiServer"
	"github.com/zerobit-tech/GoQhttp/internal/rpg"
	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
	"github.com/zerobit-tech/GoQhttp/logger"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"
	"github.com/zerobit-tech/GoQhttp/utils/xmlutils"
	bolt "go.etcd.io/bbolt"
)

// type LogStruct struct {
// 	I        int
// 	Id       string
// 	Message  string
// 	TestMode bool
// }

// var LogChan chan LogStruct = make(chan LogStruct, 5000)

type ApiCall struct {
	ID string
	//Request        map[string]any
	RequestFlatMap map[string]xmlutils.ValueDatatype
	RequestHeader  map[string]string

	Response      *storedProc.StoredProcResponse
	Err           error
	StatusCode    int
	StatusMessage string

	Log []*logger.LogEvent

	logMutex sync.Mutex

	LogDB *bolt.DB

	HttpRequest *http.Request

	CurrentSP          *storedProc.StoredProc
	CurrentRpgEndPoint *rpg.RpgEndPoint

	Server *ibmiServer.Server

	SPCallDuration time.Duration // int64 nanoseconds
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (a *ApiCall) HasError() bool {
	if a.Err != nil {

		if errors.Is(a.Err, driver.ErrBadConn) {
			a.StatusCode = http.StatusInternalServerError
			a.StatusMessage = a.Err.Error()
			go a.Logger("ERROR", fmt.Sprintf("Connection Error: %s", a.Server.Name), false) //goroutine

			return true
		}

		tmpCode, tmpStatus, errMessage, ok := a.Server.ErrorToHttpStatus(a.Err)
		if ok {
			a.StatusCode = tmpCode
			a.StatusMessage = tmpStatus
			go a.Logger("ERROR", fmt.Sprintf("Server Error %s:%s", a.StatusMessage, errMessage), false) //goroutine

			return true
		}
		a.StatusCode = http.StatusBadRequest
		a.StatusMessage = a.Err.Error()
		go a.Logger("ERROR", fmt.Sprintf("Error %s", a.Err.Error()), false) //goroutine

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
	} else {
		a.StatusCode = a.Response.Status
		a.StatusMessage = a.Response.Message
	}
	a.Response.ReferenceId = a.ID

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (a *ApiCall) BuildHeaders() http.Header {
	keysToRemove := make([]string, 0)
	headers := make(http.Header)
	for k, v := range a.Response.Data {
		if strings.HasPrefix(strings.ToUpper(k), "QHTTP_HEADER_") {
			keysToRemove = append(keysToRemove, k)
			hk := strings.TrimPrefix(k, "QHTTP_HEADER_")
			hk = stringutils.ToCamel(hk)
			hk = strings.ReplaceAll(hk, "_", "-")

			hv, ok := v.(string)
			if ok {
				headers[hk] = []string{hv}
			}
		}
	}

	// Remove QHTTP_HEADER_* keys from the response
	for _, k := range keysToRemove {
		delete(a.Response.Data, k)
	}

	return headers
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
func (apiCall *ApiCall) Logger(logType string, logEntry string, scrubeData bool) {
	defer concurrent.Recoverer("LogError")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	defer apiCall.logMutex.Unlock()

	apiCall.logMutex.Lock()

	logE := logger.GetLogEvent(logType, apiCall.ID, logEntry, scrubeData)

	apiCall.Log = append(apiCall.Log, logE)

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
// // //
// // ------------------------------------------------------
// //
// // ------------------------------------------------------
func (m *ApiCall) SaveLogs(testMode bool) {
	defer concurrent.Recoverer("SaveLogs")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
	defer m.logMutex.Unlock()

	m.logMutex.Lock()
	//m.Log = append(m.Log, m.Response.LogData...)

	for _, s := range m.Log {
		//SaveLogs(m.LogDB, i, m.ID, s, testMode)
		//logE := LogStruct{I: i, Id: m.ID, Message: s, TestMode: testMode}
		logger.LoggerChan <- s
	}

	for _, s := range m.Response.LogData {
		//SaveLogs(m.LogDB, i, m.ID, s, testMode)
		//logE := LogStruct{I: i, Id: m.ID, Message: s, TestMode: testMode}

		logger.LoggerChan <- s
	}
}

// // ------------------------------------------------------
// //
// // ------------------------------------------------------
// func SaveLogs(db *bolt.DB) {
// 	defer concurrent.Recoverer("SaveLogs")
// 	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

// 	for {
// 		logS := <-LogChan
// 		scrubed := logS.Message
// 		if !logS.TestMode && logS.I <= 1000 {
// 			scrubed = logger.RemoveNonLogData(logS.Message)
// 		}

// 		db.Update(func(tx *bolt.Tx) error {
// 			bucket, err := tx.CreateBucketIfNotExists(getLogTableName())
// 			if err != nil {
// 				return err
// 			}
// 			key := fmt.Sprintf("%s_%d", logS.Id, logS.I)
// 			bucket.Put([]byte(key), []byte(fmt.Sprintf("%05d. %s", logS.I+1, scrubed)))
// 			return nil
// 		})
// 	}

// }

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

// ------------------------------------------------------
//
// ------------------------------------------------------
func DeleteLog(db *bolt.DB, id string) {

	db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(getLogTableName())
		if err != nil {
			return errors.New("table does not exits")
		}
		c := bucket.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if bytes.HasPrefix(k, []byte(id)) {
				bucket.Delete(k)
			}
		}

		return nil
	})
	//	fmt.Println(">>>>>>>>>>>  DELETE LOG ERROR >>>>>>>>>>", err)
}
