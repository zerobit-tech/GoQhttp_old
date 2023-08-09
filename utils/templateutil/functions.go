package templateutil

import (
	"encoding/json"
	"net/http"
	"text/template"
	"time"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func httpCodeText(i int) string {
	return http.StatusText(i)
}

func humanDate(t time.Time) string {
	// Return the empty string if time has the zero value.
	if t.IsZero() {
		return ""
	}
	// Convert the time to UTC before formatting it.
	//time.Kitchen
	return t.Local().Format("02 Jan 2006 at 03:04:05PM")
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

func toJson(s interface{}) string {
	jsonBytes, err := json.Marshal(s)
	if err != nil {
		return ""
	}

	return string(jsonBytes)
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func yesNo(s bool) string {
	if s {
		return "Yes"
	}

	return "No"
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Initialize a template.FuncMap object and store it in a global variable. This is essentially
// a string-keyed map which acts as a lookup between the names of our custom template
// functions and the functions themselves.
var functions = template.FuncMap{
	"humanDate": humanDate,
	"toJson":    toJson,
	"yesNo":     yesNo,

	"httpCodeText": httpCodeText,
}
