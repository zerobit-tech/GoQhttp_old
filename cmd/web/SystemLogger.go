package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/zerobit-tech/GoQhttp/logger"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
	"github.com/zerobit-tech/GoQhttp/utils/stringutils"

	bolt "go.etcd.io/bbolt"
)

// var InfoLog *log.Logger = log.New(os.Stderr, "INFO \t", log.Ldate|log.Ltime)
// var ErrorLog *log.Logger = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime)
// var RequestLog *log.Logger = log.New(os.Stderr, "Request\t", log.Ldate|log.Ltime)
// var ResponseLog *log.Logger = log.New(os.Stderr, "Response\t", log.Ldate|log.Ltime)

// TODO move to app level

// ----------------------------------------------------------------------------------
//
// ----------------------------------------------------------------------------------
type SystemLogEvent struct {
	EventTime          time.Time
	Id                 string
	Type               string
	Message            string
	ScrubeData         bool
	TriggerByUserId    string
	ImpactedUserId     string
	ImpactedServerId   string
	ImpactedEndpointId string
	BeforeUpdate       string
	AfterUpdate        string
}

// ----------------------------------------------------------------------------------
//
// ----------------------------------------------------------------------------------

func (l *SystemLogEvent) String() string {

	return fmt.Sprintf("%s\t%s\t%s", l.EventTime.Format(stringutils.TimestampFormat), l.Type, l.Message)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) SystemLogHandlers(router *chi.Mux) {

	router.Route("/syslogs", func(r chi.Router) {
		// CSRF
		r.Use(app.sessionManager.LoadAndSave)
		r.Use(app.RequireAuthentication)
		r.Use(app.RequireSuperAdmin)
		r.Use(noSurf)
		r.Use(CheckLicMiddleware)
		r.Get("/{page}", app.systemLogs)
		r.Get("/d/{id}", app.systemLogDetail)

	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) systemLogs(w http.ResponseWriter, r *http.Request) {

	page := chi.URLParam(r, "page")

	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		pageNumber = 1
	}

	pageSize := 10

	data := app.newTemplateData(r)
	data.SystemLogEntries = app.GetSystemlogs(pageNumber, pageSize)

	data.NextPageNumber = pageNumber + 1

	app.render(w, r, http.StatusOK, "systemlog_list.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) systemLogDetail(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	data := app.newTemplateData(r)
	data.SystemLogEntry = app.GetSystemlogDetail(id)

	app.render(w, r, http.StatusOK, "systemlog_detail.tmpl", data)

}

// ----------------------------------------------------------------------------------
//
// ----------------------------------------------------------------------------------
func GetSystemLogEvent(triggerBy string, logType string, message string, scrubeData bool) *SystemLogEvent {

	return &SystemLogEvent{

		EventTime: time.Now(),

		Type:            logType,
		Message:         message,
		ScrubeData:      scrubeData,
		TriggerByUserId: triggerBy,
	}

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func getSystemLogTableName() []byte {
	return []byte("systemlogs")
}

// ----------------------------------------------------------------------------------
//
// ----------------------------------------------------------------------------------
func (app *application) SystemLogger() {
	db := app.SystemLogDB
	defer concurrent.Recoverer("SaveSystemLogs")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
	var TimestampFormat string = "20060102150405.000000"

	for {
		logEvent, ok := <-app.SystemLoggerChan
		if !ok {
			continue
		}

		logEvent.Id = url.QueryEscape(strings.ReplaceAll(logEvent.EventTime.Format(TimestampFormat), ".", ""))

		if logEvent.ScrubeData {
			logEvent.Message = logger.RemoveNonLogData(logEvent.Message)

		}

		db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists(getSystemLogTableName())
			if err != nil {
				return err
			}

			jData, err := json.Marshal(logEvent)

			if err != nil {
				return err
			}

			bucket.Put([]byte(logEvent.Id), jData)
			return nil
		})

	}
}

// ----------------------------------------------------------------------------------
//
// ----------------------------------------------------------------------------------
func (app *application) GetSystemlogs(page, pageSize int) []SystemLogEvent {
	if page <= 1 {
		page = 1
	}

	if pageSize <= 1 {
		pageSize = 10
	}

	logs := make([]SystemLogEvent, 0)

	_ = app.SystemLogDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(getSystemLogTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		c := bucket.Cursor()
		counter := 0
		for k, v := c.First(); k != nil; k, v = c.Next() {
			counter += 1

			if counter <= ((page - 1) * pageSize) {
				continue
			}

			log := SystemLogEvent{}
			err := json.Unmarshal(v, &log)
			if err == nil {
				//server.Load()
				logs = append([]SystemLogEvent{log}, logs...)
			}

			if len(logs) >= pageSize {
				break
			}
		}

		return nil
	})
	return logs

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// We'll use the Exists method to check if a user exists with a specific ID.
func (app *application) GetSystemlogDetail(id string) SystemLogEvent {

	if id == "" {
		return SystemLogEvent{}
	}
	var jData []byte // = make([]byte, 0)

	err := app.SystemLogDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(getSystemLogTableName())
		if bucket == nil {
			return errors.New("table does not exits")
		}
		jData = bucket.Get([]byte(id))

		return nil

	})
	log := SystemLogEvent{}
	if err != nil {
		return log
	}

	// log.Println("savedQueryJSON >2 >>", savedQueryJSON)

	if jData != nil {
		json.Unmarshal(jData, &log)

	}

	return log

}
