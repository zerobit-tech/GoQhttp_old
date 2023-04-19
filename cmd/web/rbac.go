package main

// import (
// 	"fmt"
// 	"net/http"
// 	"strings"

// 	"github.com/go-chi/chi/v5"
// 	bolt "go.etcd.io/bbolt"
// )

// // ------------------------------------------------------
// //
// // ------------------------------------------------------
// func (app *application) CheckRbacAccess(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 	})

// }

// // ----------------------------------------------------
// //
// // ---------------------------------------------------
// func (app *application) CreatePermission() {
// 	app.rbac.Model.DeleteAllPermissions()
// 	go app.CreateDBPermissions()

// 	go app.CreateHttpPathPermissions()

// }

// // ----------------------------------------------------
// //
// // ---------------------------------------------------
// func (app *application) CreateDBPermissions() {

// 	dbPermissions := []string{"READ", "INSERT", "UPDATE", "DELETE"}

// 	err2 := app.DB.View(func(tx *bolt.Tx) error {
// 		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
// 			tableName := strings.ToUpper(string(name))

// 			for _, p := range dbPermissions {
// 				permissionName := fmt.Sprintf("%s_%s", tableName, p)

// 				app.rbac.Model.SavePermission(permissionName)
// 			}
// 			return nil
// 		})
// 	})

// 	if err2 != nil {

// 	}
// }

// //----------------------------------------------------
// //
// //---------------------------------------------------

// func (app *application) CreateHttpPathPermissions() {

// 	skippedPermissions := []string{
// 		"/static",
// 		"/ws",
// 		"/rbac",
// 		"/user/login",
// 		"/user/logout",
// 		"/api/",
// 	}

// 	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {

// 		permissionName := route

// 		for _, p := range skippedPermissions {
// 			if strings.HasPrefix(permissionName, p) {
// 				return nil
// 			}
// 		}
// 		app.rbac.Model.SavePermission(permissionName)

// 		return nil
// 	}

// 	if err := chi.Walk(app.routes(), walkFunc); err != nil {
// 	}
// }
