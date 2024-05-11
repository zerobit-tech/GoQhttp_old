package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form/v4"
	"github.com/zerobit-tech/GoQhttp/internal/models"
	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Create a new decodePostForm() helper method. The second parameter here, dst,
// is the target destination that we want to decode the form data into.
func (app *application) decodePostForm(r *http.Request, dst any) error {
	// Call ParseForm() on the request, in the same way that we did in our
	// createSnippetPost handler.
	err := r.ParseForm()
	if err != nil {
		return err
	}
	// Call Decode() on our decoder instance, passing the target destination as
	// the first parameter.
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		// If we try to use an invalid target destination, the Decode() method
		// will return an error with the type *form.InvalidDecoderError.We use
		// errors.As() to check for this and raise a panic rather than returning
		// the error.
		var invalidDecoderError *form.InvalidDecoderError
		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}
		// For all other errors, we return them as normal.
		return err
	}
	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
type envelope map[string]interface{}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Define a writeJSON() helper for sending responses. This takes the destination
// http.ResponseWriter, the HTTP status code to send, the data to encode to JSON, and a
// header map containing any additional HTTP headers we want to include in the response.
func (app *application) writeJSON(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	// Encode the data to JSON, returning the error if there was one.
	js, err := json.MarshalIndent(data, " ", "  ")
	if err != nil {
		return err
	}
	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')
	// At this point, we know that we won't encounter any more errors before writing the
	// response, so it's safe to add any headers that we want to include. We loop
	// through the header map and add each header to the http.ResponseWriter header map.
	// Note that it's OK if the provided header map is nil. Go doesn't throw an error
	// if you try to range over (or generally, read from) a nil map.
	for key, value := range headers {
		w.Header()[key] = value
	}
	// Add the "Content-Type: application/json" header, then write the status code and
	// JSON response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

func (app *application) writeJSONAPI(w http.ResponseWriter, data *storedProc.StoredProcResponse, headers http.Header) error {
	// Encode the data to JSON, returning the error if there was one.
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(data.Status)
	w.Write(js)
	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Define a writeJSON() helper for sending responses. This takes the destination
// http.ResponseWriter, the HTTP status code to send, the data to encode to JSON, and a
// header map containing any additional HTTP headers we want to include in the response.
func (app *application) writeJSONorXML(responseType string, w http.ResponseWriter, status int, data string, headers http.Header) error {

	// Encode the data to JSON, returning the error if there was one.
	js := []byte(data)

	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')
	// At this point, we know that we won't encounter any more errors before writing the
	// response, so it's safe to add any headers that we want to include. We loop
	// through the header map and add each header to the http.ResponseWriter header map.
	// Note that it's OK if the provided header map is nil. Go doesn't throw an error
	// if you try to range over (or generally, read from) a nil map.
	for key, value := range headers {
		w.Header()[key] = value
	}
	// Add the "Content-Type: application/json" header, then write the status code and
	// JSON response.
	w.Header().Set("Content-Type", fmt.Sprintf("application/%s", strings.ToLower(responseType)))
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func getTabIds(r *http.Request) (string, string) {
	oldTabId := ""
	cookielast, err := r.Cookie("lasttabid")
	if err != nil {
		// switch {
		// case errors.Is(err, http.ErrNoCookie):
		// 	http.Error(w, "cookie not found", http.StatusBadRequest)
		// default:
		// 	log.Println(err)
		// 	http.Error(w, "server error", http.StatusInternalServerError)
		// }
		//return
	} else {
		oldTabId = cookielast.Value
		//fmt.Println("last TABDiD >>>>>>>>>>>>>>>>..", oldTabId)
	}

	currentTabId := ""
	cookie, err := r.Cookie("tabid")
	if err != nil {
		// switch {
		// case errors.Is(err, http.ErrNoCookie):
		// 	http.Error(w, "cookie not found", http.StatusBadRequest)
		// default:
		// 	log.Println(err)
		// 	http.Error(w, "server error", http.StatusInternalServerError)
		// }
		//return
	} else {
		currentTabId = cookie.Value
		//fmt.Println("TABDiD >>>>>>>>>>>>>>>>..", currentTabId)
	}

	return currentTabId, oldTabId
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) goBack(w http.ResponseWriter, r *http.Request, status int) {

	http.Redirect(w, r, r.Header.Get("Referer"), status)
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) serverError500(w http.ResponseWriter, r *http.Request, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// app.errorLog.Println(trace)
	app.errorLog.Output(2, trace) // make sure error log does not show this helper.go as the error trigger
	// http.Error(w, err.Error(), http.StatusInternalServerError)

	data := app.newTemplateData(r)

	app.sessionManager.Put(r.Context(), "error", err.Error())

	app.render(w, r, http.StatusUnprocessableEntity, "500.tmpl", data)

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) UnauthorizedErrorJSON(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

}

func (app *application) ErrorJSON(w http.ResponseWriter, r *http.Request, code int) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) UnauthorizedError(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)

	app.sessionManager.Put(r.Context(), "error", "Unauthorized")
	w.WriteHeader(http.StatusUnauthorized)

	app.render(w, r, http.StatusUnprocessableEntity, "401.tmpl", data)

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) Http404(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	w.WriteHeader(http.StatusUnauthorized)
	app.render(w, r, http.StatusUnprocessableEntity, "404.tmpl", data)

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) clientError(w http.ResponseWriter, status int, err error) {
	message := http.StatusText(status)
	if err != nil {
		message = err.Error()
	}
	http.Error(w, message, status)

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func (app *application) notFound(w http.ResponseWriter, err error) {
	app.clientError(w, http.StatusNotFound, err)
}

// The errorResponse() method is a generic helper for sending JSON-formatted error
// messages to the client with a given status code. Note that we're using an interface{}
// type for the message parameter, rather than just a string type, as this gives us
// more flexibility over the values that we can include in the response.
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}
	// Write the response using the writeJSON() helper. If this happens to return an
	// error then log it, and fall back to sending the client an empty response with a
	// 500 Internal Server Error status code.
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		//app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError) // 500
	}
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func openbrowser(url string) {
	log.Println("Opening browser:", url)
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------

func getRoutePattern(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if pattern := rctx.RoutePattern(); pattern != "" {
		// Pattern is already available
		return pattern
	}

	routePath := r.URL.Path
	if r.URL.RawPath != "" {
		routePath = r.URL.RawPath
	}

	tctx := chi.NewRouteContext()
	if !rctx.Routes.Match(tctx, r.Method, routePath) {
		// No matching pattern, so just return the request path.
		// Depending on your use case, it might make sense to
		// return an empty string or error here instead
		return routePath
	}

	// tctx has the updated pattern, since Match mutates it
	return tctx.RoutePattern()
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Create a new decodePostForm() helper method. The second parameter here, dst,
// is the target destination that we want to decode the form data into.
func (app *application) deleteSPData(spid string) {
	defer concurrent.Recoverer("deleteSPData")
	defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	spcalllog, _ := app.spCallLogModel.Get(spid)
	for _, l := range spcalllog.Logs {
		logid := l.LogID
		models.DeleteLog(app.LogDB, logid)
	}
	app.spCallLogModel.Delete(spid)
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
// Create a new decodePostForm() helper method. The second parameter here, dst,
// is the target destination that we want to decode the form data into.
func (app *application) deleteRpgEndpointData(spid string) {}
