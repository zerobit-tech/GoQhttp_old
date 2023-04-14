package main

import (
	"net/http"

	"github.com/onlysumitg/GoQhttp/internal/validator"
	"github.com/onlysumitg/GoQhttp/rbac"

	"github.com/go-chi/chi/v5"
)

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) RbacHandlers(router *chi.Mux) {
	router.Route("/rbac", func(r chi.Router) {

		r.Use(app.RequireSuperAdmin)
		//r.With(paginate).Get("/", listArticles)
		//	r.Get("/", app.EndPointList)

		r.Get("/", app.listRoles)
		r.Get("/permissions", app.listPermissions)
		r.Get("/addrole", app.manageRoles)
		r.Post("/addrole", app.manageRoles)

		r.Get("/managerole/{role}", app.manageRolePermissions)
		r.Post("/managerole/{role}", app.manageRolePermissions)
		r.Post("/addrolepermission/{role}", app.addRolePermissions)
		r.Post("/removerolepermissions/{role}", app.removeRolePermissions)

		//r.Post("/logout", app.userLogoutPost)

	})

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) listRoles(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	roles := app.rbac.Model.ListRoles()
	data.RbacRoles = make([]string, 0)

	for k := range roles {
		data.RbacRoles = append(data.RbacRoles, k)
	}
	app.render(w, r, http.StatusOK, "rbac_roles.tmpl", data)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) listPermissions(w http.ResponseWriter, r *http.Request) {
	permissions := app.rbac.Model.ListPermission()
	data := app.newTemplateData(r)
	data.RbacPermissions = make([]string, 0)

	for k := range permissions {
		data.RbacPermissions = append(data.RbacPermissions, k)
	}
	app.render(w, r, http.StatusOK, "rbac_permissions.tmpl", data)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) manageRolePermissions(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintln(w, "Display a HTML form for signing up a new user...")
	role := chi.URLParam(r, "role")

	rp, err := app.rbac.Model.ListRolePermission(role)
	if err != nil {
		app.rbac.Model.SaveRolePermissions(rbac.RolePermissionMaper{Role: role})
	}

	// if post

	data := app.newTemplateData(r)

	data.RbacRolePermissionsIncluded = make([]string, 0)
	data.RbacRolePermissionsIncluded = append(data.RbacPermissions, rp.Permissions...)

	data.RbacRolePermissionsExcluded = make([]string, 0)
	permissions := app.rbac.Model.ListPermission()

	for _, p := range permissions {
		if !contains(data.RbacRolePermissionsIncluded, p.ID()) {
			data.RbacRolePermissionsExcluded = append(data.RbacRolePermissionsExcluded, p.ID())
		}
	}

	data.RbacRole = role
	app.render(w, r, http.StatusOK, "rbac_role_permissions.tmpl", data)
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func RemoveIndex(s []string, index int) []string {
	ret := make([]string, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}
func SliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) addRolePermissions(w http.ResponseWriter, r *http.Request) {
	role := chi.URLParam(r, "role")

	rp, err := app.rbac.Model.ListRolePermission(role)
	if err != nil {
		app.rbac.Model.SaveRolePermissions(rbac.RolePermissionMaper{Role: role})

	}

	rp.Role = role
	var form rbac.RbackPermissionForm = rbac.RbackPermissionForm{}

	err = app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest, err)
		return
	}
	form.CheckField(validator.NotBlank(form.Permission), "permission", "This field cannot be blank")
	if form.Valid() {
		rp.Permissions = append(rp.Permissions, form.Permission)

		err = app.rbac.Model.SaveRolePermissions(rp)
		if err != nil {
			app.clientError(w, http.StatusBadRequest, err)
			return
		}
	}
	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) removeRolePermissions(w http.ResponseWriter, r *http.Request) {
	role := chi.URLParam(r, "role")

	rp, err := app.rbac.Model.ListRolePermission(role)
	if err != nil {
		app.rbac.Model.SaveRolePermissions(rbac.RolePermissionMaper{Role: role})

	}
	var form rbac.RbackPermissionForm = rbac.RbackPermissionForm{}

	err = app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest, err)
		return
	}
	form.CheckField(validator.NotBlank(form.Permission), "permission", "This field cannot be blank")
	if form.Valid() {
		index := SliceIndex(len(rp.Permissions), func(i int) bool { return rp.Permissions[i] == form.Permission })

		if index >= 0 {
			rp.Permissions = RemoveIndex(rp.Permissions, index)

		}
		app.rbac.Model.SaveRolePermissions(rp)
	}
	app.goBack(w, r, http.StatusSeeOther)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) manageRoles(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintln(w, "Display a HTML form for signing up a new user...")

	var form rbac.RbackRoleForm = rbac.RbackRoleForm{}

	if r.Method == http.MethodPost {
		err := app.decodePostForm(r, &form)
		if err != nil {
			app.clientError(w, http.StatusBadRequest, err)
			return
		}
		form.CheckField(validator.NotBlank(form.Role), "role", "This field cannot be blank")
		if form.Valid() {
			app.rbac.Model.SaveRole(form.Role)
			nextUrl := r.URL.Query().Get("next")
			if nextUrl == "" {
				nextUrl = "/rbac"
			}
			http.Redirect(w, r, nextUrl, http.StatusSeeOther)
			return
		}
	}
	data := app.newTemplateData(r)
	data.Form = form
	app.render(w, r, http.StatusOK, "rbac_addrole.tmpl", data)
}

// ------------------------------------------------------
//  Can not add permissions manually
// ------------------------------------------------------
// func (app *application) managePermissions(w http.ResponseWriter, r *http.Request) {
// 	//fmt.Fprintln(w, "Display a HTML form for signing up a new user...")

// 	var form rbac.RbackPermissionForm = rbac.RbackPermissionForm{}

// 	if r.Method == http.MethodPost {
// 		err := app.decodePostForm(r, &form)
// 		if err != nil {
// 			app.clientError(w, http.StatusBadRequest, err)
// 			return
// 		}
// 		form.CheckField(validator.NotBlank(form.Permission), "role", "This field cannot be blank")
// 		if form.Valid() {
// 			app.rbac.Model.SavePermission(form.Permission)
// 			nextUrl := r.URL.Query().Get("next")
// 			if nextUrl == "" {
// 				nextUrl = "/permissions"
// 			}
// 			http.Redirect(w, r, nextUrl, http.StatusSeeOther)
// 			return
// 		}
// 	}
// 	data := app.newTemplateData(r)
// 	data.Form = form
// 	app.render(w, r, http.StatusOK, "rbac_permission.tmpl", data)
// }
