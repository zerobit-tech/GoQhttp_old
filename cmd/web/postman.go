package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	postman "github.com/rbretecher/go-postman-collection"
	"github.com/zerobit-tech/GoQhttp/internal/ibmiServer"
	"github.com/zerobit-tech/GoQhttp/internal/models"
	"github.com/zerobit-tech/GoQhttp/internal/storedProc"
	"github.com/zerobit-tech/GoQhttp/utils/httputils"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) PostmantHandlers(router *chi.Mux) {
	router.Route("/postman", func(r chi.Router) {
		//r.With(paginate).Get("/", listArticles)
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(CheckLicMiddleware)

		r.Get("/s/{serverid}", app.downloadPostmanCollectionForServer)
		r.Get("/u/{userid}", app.downloadPostmanCollectionForUser)

	})

}

// ------------------------------------------------------
// download file
// ------------------------------------------------------
func (app *application) downloadPostmanCollection(w http.ResponseWriter, r *http.Request, c *postman.Collection, fileName string) {
	buf := bytes.NewBuffer(nil)

	err := c.Write(buf, postman.V210)

	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Description", "File Transfer")                  // can be used multiple times
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName) // can be used multiple times
	w.Header().Set("Content-Type", "application/octet-stream")

	w.Write(buf.Bytes())
}

// ------------------------------------------------------
// download file
// ------------------------------------------------------
func (app *application) downloadPostmanCollectionForUser(w http.ResponseWriter, r *http.Request) {

	userid := chi.URLParam(r, "userid")

	user, err := app.users.Get(userid)
	if err != nil {

		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	c, err := app.UserToPostmanCollection(user)
	if err != nil {

		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}
	fileName := fmt.Sprintf("QHTTP_%s.json", user.Name)

	app.downloadPostmanCollection(w, r, c, fileName)

}

// ------------------------------------------------------
// download file
// ------------------------------------------------------
func (app *application) downloadPostmanCollectionForServer(w http.ResponseWriter, r *http.Request) {

	serverID := chi.URLParam(r, "serverid")

	server, err := app.servers.Get(serverID)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error deleting server: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	c := app.ServerToPostmanCollection(server)

	fileName := fmt.Sprintf("QHTTP_%s.json", server.Name)

	app.downloadPostmanCollection(w, r, c, fileName)

}

// https://learning.postman.com/collection-format/getting-started/structure-of-a-collection/

// -----------------------------------------------------------------------
//
// -----------------------------------------------------------------------
func (app *application) getServerSPs(s *ibmiServer.Server) []*storedProc.StoredProc {

	splist := make([]*storedProc.StoredProc, 0, 10)
	for _, sp := range app.storedProcs.List(false) {
		if sp == nil || sp.DefaultServer == nil {
			continue
		}
		allowed := false
		if sp.DefaultServer.ID == s.ID {
			allowed = true
		} else {

			for _, als := range sp.AllowedOnServers {
				if als.ID == s.ID {
					allowed = true
				}

			}
		}
		if allowed {
			splist = append(splist, sp)
		}
	}

	return splist
}

// -----------------------------------------------------------------------
//
// -----------------------------------------------------------------------

func (app *application) UserToPostmanCollection(user *models.User) (*postman.Collection, error) {

	userServer, err := app.servers.Get(user.ServerId)
	if err != nil {
		return nil, err
	}

	c := app.ServerToPostmanCollection(userServer)
	c.Info.Name = fmt.Sprintf("QHTTP %s", user.Name)
	c.Info.Description.Content = fmt.Sprintf("QHTTP collection for user %s", user.Name)

	authTokenVar := postman.CreateVariable("authtoken", user.Token)
	authTokenVar.ID = "authtoken"
	authTokenVar.Key = "authtoken"

	c.Variables = append(c.Variables, authTokenVar)

	//auth := postman.CreateAuth(postman.Bearer, postman.CreateAuthParam("bearer", "{{authtoken}}"))

	//c.Auth = auth

	return c, nil

	//return c
}

// -----------------------------------------------------------------------
//
// -----------------------------------------------------------------------

func (app *application) ServerToPostmanCollection(s *ibmiServer.Server) *postman.Collection {
	c := postman.CreateCollection(fmt.Sprintf("QHTTP %s", s.Name), fmt.Sprintf("QHTTP collection for server %s", s.Name))
	c.Variables = make([]*postman.Variable, 0)
	folder := c.AddItemGroup(s.Name)

	for _, sp := range app.getServerSPs(s) {
		folder.AddItem(app.EndPointToItem(*sp))

	}

	return c
}

// -----------------------------------------------------------------------
//
// -----------------------------------------------------------------------
func (app *application) EndPointToItem(sp storedProc.StoredProc) *postman.Items {

	/*
			Name                    string      `json:"name"`
		Description             string      `json:"description,omitempty"`
		Variables               []*Variable `json:"variable,omitempty"`
		Events                  []*Event    `json:"event,omitempty"`
		ProtocolProfileBehavior interface{} `json:"protocolProfileBehavior,omitempty"`
		ID                      string      `json:"id,omitempty"`
		Request                 *Request    `json:"request,omitempty"`
		Responses               []*Response `json:"response,omitempty"`
	*/

	urlAddress := sp.MockUrl
	if sp.AllowWithoutAuth {
		urlAddress = sp.MockUrlWithoutAuth
	}

	urlAddress = fmt.Sprintf("%s/%s", app.hostURL, urlAddress)

	host := ""
	port := ""
	path := make([]string, 0)

	query := make([]*postman.QueryParam, 0)

	queryPrams, err := httputils.QueryParamToMap(urlAddress)
	if err == nil {
		for k, v := range queryPrams {
			q := &postman.QueryParam{
				Key:   k,
				Value: (v).(string),
			}
			query = append(query, q)

		}
	}

	u, err := url.Parse(urlAddress)
	if err == nil {

		// fmt.Println(">>>>>>>Scheme>>>>>>>", u.Scheme)

		// fmt.Println(">>>>>>>Opaque>>>>>>>", u.Opaque)

		if strings.Contains(u.Host, ":") {
			broken := strings.Split(u.Host, ":")
			host = broken[0]
			port = broken[1]

		}
		// fmt.Println(">>>>>>>>Host>>>>>>", u.Host)
		// fmt.Println(">>>>>>>>>>Path>>>>", u.Path)

		path = strings.Split(u.Path, "/")

		// fmt.Println(">>>>>>>>RawPath>>>>>>", u.RawPath)
		// fmt.Println(">>>>>>>>OmitHost>>>>>>", u.OmitHost)
		// fmt.Println(">>>>>>>>>ForceQuery>>>>>", u.ForceQuery)
		// fmt.Println(">>>>>>RawQuery>>>>>>>>", u.RawQuery)
		// fmt.Println(">>>>>>>>Fragment>>>>>>", u.Fragment)
		// fmt.Println(">>>>>>>>>>>RawFragment>>>", u.RawFragment)

	}

	postManItem := postman.CreateItem(postman.Item{
		Name:        sp.EndPointName,
		Description: fmt.Sprintf("Based on stored proc %s/%s", sp.Lib, sp.Name),
		ID:          sp.ID,
		Request: &postman.Request{
			URL: &postman.URL{
				Raw:      urlAddress,
				Protocol: u.Scheme,
				Host:     []string{host},
				Port:     port,
				Path:     path,
				Query:    query,
			},
			Method: postman.Method(strings.ToUpper(sp.HttpMethod)),
			Auth:   postman.CreateAuth(postman.Bearer, postman.CreateAuthParam("bearer", "{{authtoken}}")),
			Body: &postman.Body{
				Mode:    "raw",
				Raw:     sp.InputPayload,
				Options: &postman.BodyOptions{Raw: postman.BodyOptionsRaw{Language: "json"}},
			},
		},
	})

	return postManItem
}
