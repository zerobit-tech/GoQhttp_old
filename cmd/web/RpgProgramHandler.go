package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/onlysumitg/GoQhttp/internal/rpg"
	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/utils/concurrent"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RpgProgramHandlers(router *chi.Mux) {
	router.Route("/rpgprogram", func(r chi.Router) {
		//r.With(paginate).Get("/", listArticles)rpgParam
		r.Use(app.sessionManager.LoadAndSave)

		r.Use(app.RequireAuthentication)
		r.Use(CheckLicMiddleware)

		r.Get("/", app.rpgProgramList) // list

		r.Get("/{id}", app.rpgProgramView) // one single
		r.Get("/add", app.rpgProgramAdd)
		r.Post("/add", app.rpgProgramAddPost)

		r.Get("/update/{id}", app.rpgProgramUpdate)
		r.Post("/update/{id}", app.rpgProgramAddPost)

		r.Get("/delete/{spId}", app.SPDelete)
		r.Post("/delete", app.SPDeleteConfirm)

	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) rpgProgramList(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)
	data.RpgPrograms = app.RpgProgramModel.List()

	nextUrl := r.URL.Query().Get("next") //filters=["color", "price", "brand"]
	data.Next = nextUrl
	app.render(w, r, http.StatusOK, "rpg_program_list.tmpl", data)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) rpgProgramView(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	id := chi.URLParam(r, "id")

	param, err := app.RpgProgramModel.Get(id)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.RpgProgram = param
	app.render(w, r, http.StatusOK, "rpg_program_view.tmpl", data)

}

// ------------------------------------------------------

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) rpgProgramAdd(w http.ResponseWriter, r *http.Request) {

	pgm := rpg.Program{}

	data := app.newTemplateData(r)
	pgm.Init()
	data.Form = pgm
	data.RpgParams = app.RpgParamModel.List()

	app.render(w, r, http.StatusOK, "rpg_program_add.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) rpgProgramAddPost(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("001 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	var RpgProgram rpg.Program
	err = app.formDecoder.Decode(&RpgProgram, r.PostForm)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("002 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}
	RpgProgram.CheckField(validator.NotBlank(RpgProgram.Name), "name", "This field cannot be blank")
	RpgProgram.CheckField(validator.NotBlank(RpgProgram.Lib), "lib", "This field cannot be blank")
	RpgProgram.CheckField(!app.RpgProgramModel.DuplicateName(&RpgProgram), "name", "Duplicate name")

	logBeforeImage := ""

	// Check SP details from iBMI
	if RpgProgram.Valid() {

		if RpgProgram.ID != "" {
			// orginalSp, err := app.RpgProgramModel.Get(RpgProgram.ID)
			// if err == nil {
			// 	//logBeforeImage = orginalSp.LogImage()

			// }
		}

	}

	if !RpgProgram.Valid() {
		data := app.newTemplateData(r)
		data.Form = RpgProgram
		app.sessionManager.Put(r.Context(), "error", "Please fix error(s) and resubmit")
		data.RpgParams = app.RpgParamModel.List()

		app.render(w, r, http.StatusUnprocessableEntity, "rpg_program_add.tmpl", data)
		return
	}

	logAction := "RPG Param Created"
	if RpgProgram.ID != "" {
		logAction = "RPG Param Modified"
	}

	// finally save
	id, err := app.RpgProgramModel.Save(&RpgProgram)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", fmt.Sprintf("Param %s added sucessfully", RpgProgram.Name))

	go func() {
		defer concurrent.Recoverer("rpgProgramCREATED")
		defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
		userID, _ := app.getCurrentUserID(r)

		logEvent := GetSystemLogEvent(userID, logAction, fmt.Sprintf(" %s,IP %s", RpgProgram.Name, r.RemoteAddr), false)
		logEvent.ImpactedEndpointId = id
		logEvent.BeforeUpdate = logBeforeImage
		//logEvent.AfterUpdate = RpgProgram.LogImage()
		app.SystemLoggerChan <- logEvent

	}()

	//http.Redirect(w, r, fmt.Sprintf("/savesql/%s", id), http.StatusSeeOther)
	http.Redirect(w, r, fmt.Sprintf("/RpgProgram/%s", id), http.StatusSeeOther)

}

// ------------------------------------------------------
// Delete saved query
// ------------------------------------------------------
func (app *application) rpgProgramDelete(w http.ResponseWriter, r *http.Request) {

	spId := chi.URLParam(r, "spId")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error: %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	data := app.newTemplateData(r)
	data.StoredProc = sP

	app.render(w, r, http.StatusOK, "rpg_program_delete.tmpl", data)

}

// ------------------------------------------------------
// Delete saved query confirm
// ------------------------------------------------------
func (app *application) rpgProgramDeleteConfirm(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("001 Error processing form %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	spId := r.PostForm.Get("spId")
	app.invalidateEndPointCache()

	err = app.storedProcs.Delete(spId)
	if err != nil {

		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Delete error: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	go app.deleteSPData(spId) //goroutine
	app.sessionManager.Put(r.Context(), "flash", "Endpoint deleted sucessfully")

	go func() {
		defer concurrent.Recoverer("SPMODIFIED")
		userID, _ := app.getCurrentUserID(r)

		logEvent := GetSystemLogEvent(userID, "Endpoint Deleted", fmt.Sprintf("IP %s", r.RemoteAddr), false)
		logEvent.ImpactedEndpointId = spId
		app.SystemLoggerChan <- logEvent

	}()

	http.Redirect(w, r, "/sp", http.StatusSeeOther)

}

// ------------------------------------------------------
// add new server
// ------------------------------------------------------
func (app *application) rpgProgramUpdate(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	RpgProgram, err := app.RpgProgramModel.Get(id)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Update error: %s", err.Error()))
		app.goBack(w, r, http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r)

	data.Form = RpgProgram
	data.RpgParams = app.RpgParamModel.List()

	app.render(w, r, http.StatusOK, "rpg_program_add.tmpl", data)

}

// ------------------------------------------------------
// Server details
// ------------------------------------------------------
func (app *application) rpgProgramValidator(w http.ResponseWriter, r *http.Request) {

	data := app.newTemplateData(r)

	spId := chi.URLParam(r, "spId")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	data.StoredProc = sP

	data.ParamRegexs = app.paramRegexModel.List()
	app.render(w, r, http.StatusOK, "rpg_program_validator.tmpl", data)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) rpgProgramSaveValidator(w http.ResponseWriter, r *http.Request) {

	spId := r.FormValue("id")

	sP, err := app.storedProcs.Get(spId)
	if err != nil {
		app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
		app.goBack(w, r, http.StatusSeeOther)
		return
	}

	//	app.storedProcs.Save(sP)
	maped := formToMap(r)

	changed := false
	for _, p := range sP.Parameters {
		validator, found := maped[strings.ToUpper(p.Name)]
		if found {
			validatorString, ok := validator.(string)
			if ok && p.ValidatorRegex != validatorString {

				p.ValidatorRegex = validatorString

				if p.ValidatorRegex != "" {
					err := p.CheckValidatorRegex(app.paramRegexModel.Map())
					if err != nil {
						app.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Error  : %s", err.Error()))
						app.goBack(w, r, http.StatusSeeOther)
						return
					}
				}

				changed = true

			}
		}
	}

	if changed {
		app.invalidateEndPointCache()
		app.storedProcs.Save(sP)
		app.sessionManager.Put(r.Context(), "flash", "Done")

	} else {
		app.sessionManager.Put(r.Context(), "flash", "Nothing changed")

	}

	app.goBack(w, r, http.StatusSeeOther)

}
