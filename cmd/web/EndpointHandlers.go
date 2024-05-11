package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zerobit-tech/GoQhttp/internal/endpoints"
	"github.com/zerobit-tech/GoQhttp/internal/rpg"
	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) EndpointHandlers(router *chi.Mux) {
	router.Route("/endpoints", func(r chi.Router) {
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(CheckLicMiddleware)

		//r.Use(app.CurrentServerMiddleware)
		//r.With(paginate).Get("/", listArticles)
		r.Get("/", app.EndpointList)
		r.Get("/help", app.EndpointHelp)
		r.Get("/add", app.EndpointAdd)
		r.Get("/getform", app.getform)
	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) getform(w http.ResponseWriter, r *http.Request) {
	eptype := r.URL.Query().Get("eptype")
	switch eptype {
	case "pgm":
		app.RpgEndpointAdd(w, r)
	default:
		app.SPAdd(w, r)
	}
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) EndpointAdd(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	app.render(w, r, http.StatusOK, "endpoint_add.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) EndpointHelp(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	app.render(w, r, http.StatusOK, "rpg_endpoint_help_inbuilt_param.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) EndpointList(w http.ResponseWriter, r *http.Request) {

	serverID := r.URL.Query().Get("server")

	_, err := app.servers.Get(serverID)
	if err != nil {
		serverID = ""

	}

	loadSpecialparam := r.URL.Query().Get("loadspecial")

	loadSpecial := false
	if loadSpecialparam == "Y" {
		loadSpecial = true
	}

	data := app.newTemplateData(r)
	RpgEndPoints := app.GetRPGEndpintList(serverID)
	SQLEndpoints := app.GetSQLSPEndpintList(serverID, loadSpecial)

	endPoints := make([]endpoints.Endpoint, 0)

	for _, rpg := range RpgEndPoints {

		endPoints = append(endPoints, rpg)
	}

	for _, sqlsp := range SQLEndpoints {

		endPoints = append(endPoints, sqlsp)
	}
	data.Endpoints = endPoints
	nextUrl := r.URL.Query().Get("next") //filters=["color", "price", "brand"]
	data.Next = nextUrl
	app.render(w, r, http.StatusOK, "endpoint_list.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) GetRPGEndpintList(serverID string) []*rpg.RpgEndPoint {

	RpgEndPoints := make([]*rpg.RpgEndPoint, 0, 10)

	storedPs := app.RpgEndpointModel.List()

	if serverID != "" {
		for _, s := range storedPs {
			if s == nil {
				continue
			}
			allowed := false
			if s.DefaultServerId == serverID {
				allowed = true
			} else {

				for _, als := range s.AllowedOnServers {
					if als == serverID {
						allowed = true
					}

				}
			}
			if allowed {
				RpgEndPoints = append(RpgEndPoints, s)
			}
		}
	} else {
		RpgEndPoints = storedPs
	}

	return RpgEndPoints

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) GetSQLSPEndpintList(serverID string, loadSpecial bool) []*storedProc.StoredProc {

	StoredProcs := make([]*storedProc.StoredProc, 0, 10)

	storedPs := app.storedProcs.List(loadSpecial)

	if serverID != "" {
		for _, s := range storedPs {
			if s == nil || s.DefaultServer == nil {
				continue
			}
			allowed := false
			if s.DefaultServer.ID == serverID {
				allowed = true
			} else {

				for _, als := range s.AllowedOnServers {
					if als.ID == serverID {
						allowed = true
					}

				}
			}
			if allowed {
				StoredProcs = append(StoredProcs, s)
			}
		}
	} else {
		StoredProcs = storedPs
	}

	return StoredProcs

}
