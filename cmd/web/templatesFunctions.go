package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"
)

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
func IsPreFormatted(s string) bool {
	// if strings.HasPrefix(s, "00999") || strings.HasPrefix(s, "01000") {
	// 	return true
	// }
	return true
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func httpCodeText(i int) string {
	return http.StatusText(i)
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
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
func (app *application) getServerName(id string) string {
	server, err := app.servers.Get(id)
	if err != nil {
		return "UNKNOWN"
	}

	return server.Name
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) getUserName(id string) string {
	user, err := app.users.Get(id)
	if err != nil {
		return ""
	}

	return user.Email
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) getRpgParamString(id string) string {
	p, err := app.RpgParamModel.Get(id)
	if err != nil {
		return ""
	}

	return p.ToString()
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) indexBy1(id int) int {
	return id + 1
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Initialize a template.FuncMap object and store it in a global variable. This is essentially
// a string-keyed map which acts as a lookup between the names of our custom template
// functions and the functions themselves.

func (app *application) getFunctionMap() template.FuncMap {
	var functions = template.FuncMap{
		"humanDate":         humanDate,
		"toJson":            toJson,
		"yesNo":             yesNo,
		"ispreformatted":    IsPreFormatted,
		"httpCodeText":      httpCodeText,
		"servername":        app.getServerName,
		"username":          app.getUserName,
		"indexby1":          app.indexBy1,
		"getrpgparamstring": app.getRpgParamString,
	}

	return functions
}
