package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/zerobit-tech/GoQhttp/internal/validator"
	"github.com/zerobit-tech/GoQhttp/utils/concurrent"
	"github.com/zerobit-tech/GoQhttp/utils/jwtutils"

	"github.com/go-chi/chi/v5"
	"github.com/zerobit-tech/GoQhttp/internal/models"
)

func (app *application) CreateSuperUser(email, password string) {
	// defer concurrent.Recoverer("GetByEmail")
	// defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

	if app.features.AdminEmail != "" {
		email = app.features.AdminEmail
	}

	if app.features.AdminPassword != "" {
		password = app.features.AdminPassword
	}

	user, err := app.users.GetByEmail(email)

	if err == nil {
		if password != "" {
			user.Password = password
		}
	} else {
		user = &models.User{
			Name:        "SuperAdmin",
			Email:       email,
			Password:    password,
			IsSuperUser: true,
			IsStaff:     true,
			HasVerified: true,
		}
	}

	user.HasVerified = true
	user.IsSuperUser = true
	user.MaxAllowedEndpoints = -1
	err = app.users.Save(user, true)
	if err != nil {
		log.Fatalln("Error creating super user.", err.Error())
	}
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) UserHandlers(router *chi.Mux) {
	router.Route("/user", func(r chi.Router) {
		r.Use(app.sessionManager.LoadAndSave)

		//r.With(paginate).Get("/", listArticles)
		//	r.Get("/", app.EndPointList)
		g1 := r.Group(nil)
		g1.Use(noSurf)

		// g1.Get("/signup", app.userSignup)
		// g1.Post("/signup", app.userSignupPost)
		// CSRF
		g1.Get("/login", app.userLogin)
		g1.Post("/login", app.userLoginPost)

		g1.Get("/verify/{userid}/{verificationid}", app.userVerification)

		g1.Get("/trouble", app.userLoginTrouble)
		g1.Post("/trouble", app.userLoginTrouble)

		g1.Get("/resetpwd/{userid}/{verificationid}", app.passwordReset)
		g1.Post("/resetpwd/{userid}/{verificationid}", app.passwordReset)
		r.Post("/logout", app.userLogoutPost)
	})

}

// ------------------------------------------------------
//
//	middleware
//
// ------------------------------------------------------
func (app *application) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the user is not authenticated, redirect them to the login page and
		// return from the middleware chain so that no subsequent handlers in
		// the chain are executed.
		goToUrl := fmt.Sprintf("/user/login?next=%s", r.URL.RequestURI())

		if !app.isAuthenticated(r, true) {
			//	app.sessionManager.Put(r.Context(), "error", "Login required")

			http.Redirect(w, r, goToUrl, http.StatusSeeOther)
			return
		}

		user, err := app.GetUser(r)
		if !user.HasVerified {
			app.sessionManager.Put(r.Context(), "error", "Please verify your email")

			http.Redirect(w, r, goToUrl, http.StatusSeeOther)
			return
		}

		userId := ""
		if err == nil {
			userId = user.ID

			app.UpdateSessionUserToketn(r, *user)
		}
		ctx := context.WithValue(r.Context(), models.ContextUserKey, userId)
		// Otherwise set the "Cache-Control: no-store" header so that pages
		// require authentication are not stored in the users browser cache (or
		// other intermediary cache).
		w.Header().Add("Cache-Control", "no-store")
		// And call the next handler in the chain.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ------------------------------------------------------
func (app *application) RequireSuperAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !app.isSuperAdmin(r) {
			app.UnauthorizedError(w, r)
			return
		}

		// And call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// ------------------------------------------------------
func (app *application) RequireStaff(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !app.isStaff(r) {
			app.UnauthorizedError(w, r)
			return
		}

		// And call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// // ------------------------------------------------------
// func (app *application) MustHasPathsPermission(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		permissionID := getRoutePattern(r)
// 		permission := gorbac.StdPermission{IDStr: permissionID}
// 		if app.isSuperAdmin(r) {
// 			next.ServeHTTP(w, r)
// 			return
// 		}

// 		// if permission is not available in the system ==> continue
// 		if !app.rbac.Model.PermissionExists(permissionID) {
// 			next.ServeHTTP(w, r)
// 			return
// 		}

// 		// if permission is not in use --> means this permission is allowed for all users
// 		inUse := gorbac.AnyGranted(app.rbac.RBAC, app.rbac.Model.ListRolesAsString(), permission, nil)

// 		if !inUse {
// 			next.ServeHTTP(w, r)
// 			return
// 		}

// 		// check for loggined in user  =>
// 		user, err := app.GetUser(r)

// 		if err == nil {
// 			if app.rbac.RBAC.IsGranted(user.Role, permission, nil) {
// 				next.ServeHTTP(w, r)
// 				return
// 			}

// 		}

// 		// And call the next handler in the chain.
// 		app.UnauthorizedError(w, r)
// 	})
// }

// ------------------------------------------------------
//
// ------------------------------------------------------
// Return true if the current request is from an authenticated user, otherwise
// return false.
func (app *application) isAuthenticated(r *http.Request, allowOnlyStaff bool) bool {
	user, err := app.GetUser(r)

	if err != nil {
		return false
	}

	if allowOnlyStaff && !(user.IsSuperUser || user.IsStaff) {
		return false
	}

	return true
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// Return true if the current request is from an authenticated user, otherwise
// return false.
func (app *application) isSuperAdmin(r *http.Request) bool {
	user, err := app.GetUser(r)

	if err != nil {
		return false
	}

	if user.IsSuperUser {
		return true
	}

	return false

}

// ------------------------------------------------------
//
// ------------------------------------------------------
// Return true if the current request is from an authenticated user, otherwise
// return false.
func (app *application) RemoveSessionUser(r *http.Request) {
	//app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Remove(r.Context(), "authenticatedUserToken")

}

// ------------------------------------------------------
//
// ------------------------------------------------------

func (app *application) UpdateSessionUserToketn(r *http.Request, user models.User) error {

	jwtString, err := user.GetNewToken(120 * time.Minute)
	if err != nil {
		return err
	}

	// Add the ID of the current user to the session, so that they are now
	// 'logged in'.
	// app.sessionManager.Put(r.Context(), "authenticatedUserID", user.ID)
	app.sessionManager.Put(r.Context(), "authenticatedUserToken", jwtString)
	return nil
}

// ------------------------------------------------------
//
// ------------------------------------------------------
// Return true if the current request is from an authenticated user, otherwise
// return false.
func (app *application) isStaff(r *http.Request) bool {
	user, err := app.GetUser(r)

	if err != nil {
		return false
	}

	if user.IsStaff {
		return true
	}

	return false

}

// ------------------------------------------------------
//
// ------------------------------------------------------
// Return true if the current request is from an authenticated user, otherwise
// return false.
func (app *application) GetUser(r *http.Request) (*models.User, error) {
	//userId := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

	usedIDString, err := app.getCurrentUserID(r)

	if err != nil {
		return nil, err
	}

	return app.users.Get(usedIDString)

}

// ------------------------------------------------------
//
// ------------------------------------------------------
// Return true if the current request is from an authenticated user, otherwise
// return false.
func (app *application) getCurrentUserID(r *http.Request) (string, error) {
	userToken := app.sessionManager.GetString(r.Context(), "authenticatedUserToken")

	claims, err := jwtutils.Parse(userToken)

	if err != nil {
		return "UNKNOWN", err
	}

	usedID, found := claims["sub"]
	if !found {
		return "UNKNOWN", fmt.Errorf("invalid token")
	}

	usedIDString, ok := usedID.(string)
	if !ok {
		return "UNKNOWN", fmt.Errorf("invalid token")
	}

	return usedIDString, nil

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintln(w, "Display a HTML form for signing up a new user...")
	// Use the RenewToken() method on the current session to change the session
	// ID again.
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	// Remove the authenticatedUserID from the session data so that the user is
	// 'logged out'.
	app.RemoveSessionUser(r)

	data := app.newTemplateData(r)
	data.Form = models.UserSignUpForm{}
	app.render(w, r, http.StatusOK, "account_signup.tmpl", data)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	// Declare an zero-valued instance of our userSignupForm struct.
	var form models.UserSignUpForm
	// Parse the form data into the userSignupForm struct.
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest, err)
		return
	}
	// Validate the form contents using our helper functions.
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	_, err = app.users.GetByEmail(form.Email)
	if err == nil {
		form.CheckField(false, "email", "Email already in use.")
	}
	// If there are any errors, redisplay the signup form along with a 422
	// status code.
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "account_signup.tmpl", data)
		return
	}

	user := &models.User{
		Name:                form.Name,
		Email:               form.Email,
		Password:            form.Password,
		MaxAllowedEndpoints: app.maxAllowedEndPointsPerUser,
	}

	err = app.users.Save(user, true)

	if err != nil {
		app.serverError500(w, r, err)
		return
	}

	//goroutine
	go func() {
		defer concurrent.Recoverer("BuildverificationEmail")
		defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

		emailRequest := app.users.BuildVerificationEmail(user, app.hostURL)
		app.SendEmail(emailRequest)
	}()

	nextUrl := r.URL.Query().Get("next")
	if nextUrl == "" {
		nextUrl = app.appLangingPage()
	}

	app.sessionManager.Put(r.Context(), "flash", "Please check your email.")

	http.Redirect(w, r, nextUrl, http.StatusSeeOther)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	nextUrl := r.URL.Query().Get("next")
	if nextUrl == "" {
		nextUrl = app.appLangingPage()
	}

	nextUrl = strings.TrimSpace(nextUrl)
	data := app.newTemplateData(r)
	data.Form = models.UserLoginForm{}
	data.Next = nextUrl
	app.render(w, r, http.StatusOK, "account_login.tmpl", data)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	nextUrl := r.URL.Query().Get("next")
	if nextUrl == "" {
		nextUrl = app.appLangingPage()
	}

	nextUrl = strings.TrimSpace(nextUrl)

	var form models.UserLoginForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest, nil)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		data.Next = nextUrl
		app.render(w, r, http.StatusBadRequest, "account_login.tmpl", data)
		return
	}

	user, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			app.sessionManager.Put(r.Context(), "error", "Login Failed")
			form.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			data.Next = nextUrl
			app.render(w, r, http.StatusBadRequest, "account_login.tmpl", data)

		} else {
			app.serverError500(w, r, err)
		}
		return
	}
	// Use the RenewToken() method on the current session to change the session
	// ID. It's good practice to generate a new session ID when the
	// authentication state or privilege levels changes for the user (e.g. login
	// and logout operations).
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError500(w, r, err)
		return
	}

	app.UpdateSessionUserToketn(r, *user)

	if err != nil {
		app.serverError500(w, r, err)
		return
	}

	go func() {
		defer concurrent.Recoverer("UserLogin log")
		logEvent := GetSystemLogEvent(user.ID, "LOGIN", fmt.Sprintf("IP %s", r.RemoteAddr), false)
		app.SystemLoggerChan <- logEvent

	}()

	if user.LandingPage != "" {
		nextUrl = user.LandingPage
	}
	http.Redirect(w, r, nextUrl, http.StatusSeeOther)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	// Use the RenewToken() method on the current session to change the session
	// ID again.

	currentUserId, err := app.getCurrentUserID(r)
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	// Remove the authenticatedUserID from the session data so that the user is
	// 'logged out'.
	app.RemoveSessionUser(r)
	// Add a flash message to the session to confirm to the user that they've been
	// logged out.
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
	// Redirect the user to the application home page.

	go func() {
		defer concurrent.Recoverer("UserOutout log")
		logEvent := GetSystemLogEvent(currentUserId, "LOGOUT", fmt.Sprintf("IP %s", r.RemoteAddr), false)
		app.SystemLoggerChan <- logEvent

	}()

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) userVerification(w http.ResponseWriter, r *http.Request) {
	// Use the RenewToken() method on the current session to change the session
	// ID again.
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	// Remove the authenticatedUserID from the session data so that the user is
	// 'logged out'.
	app.RemoveSessionUser(r)

	verificationid := chi.URLParam(r, "verificationid")
	userid := chi.URLParam(r, "userid")

	user, err := app.users.Get(userid)
	if err != nil {
		app.clientError(w, http.StatusBadRequest, nil)
	}

	if app.users.Verify(user, verificationid, app.users.GetVerificationTableName()) {
		user.HasVerified = true
		app.users.Save(user, false)
		//goroutine
		go app.users.DeleteVerificationRecord(user, app.users.GetVerificationTableName())

		app.sessionManager.Put(r.Context(), "flash", "You've been verified successfully!")

	} else {
		app.notFound(w, nil)
		return

	}
	// Redirect the user to the application home page.
	http.Redirect(w, r, app.appLangingPage(), http.StatusSeeOther)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) passwordReset(w http.ResponseWriter, r *http.Request) {
	// Use the RenewToken() method on the current session to change the session
	// ID again.
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	// Remove the authenticatedUserID from the session data so that the user is
	// 'logged out'.
	app.RemoveSessionUser(r)

	data := app.newTemplateData(r)
	form := models.UserPasswordResetForm{}

	verificationid := chi.URLParam(r, "verificationid")
	userid := chi.URLParam(r, "userid")

	user, err := app.users.Get(userid)
	if err != nil {
		app.clientError(w, http.StatusBadRequest, nil)
		return
	}

	if app.users.Verify(user, verificationid, app.users.GetPasswordResetTableName()) {

		if r.Method == http.MethodPost {
			err := app.decodePostForm(r, &form)
			if err != nil {
				app.clientError(w, http.StatusBadRequest, err)
				return
			}

			form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
			form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

			if form.Valid() {
				user.Password = form.Password
				app.users.Save(user, true)
				app.sessionManager.Put(r.Context(), "flash", "Password updated")
				//goroutine
				go app.users.DeleteVerificationRecord(user, app.users.GetPasswordResetTableName())

				http.Redirect(w, r, app.appLangingPage(), http.StatusSeeOther)

				return
			}
		}

		data.Form = form
		app.render(w, r, http.StatusOK, "account_reset_password.tmpl", data)
		return
	} else {
		app.notFound(w, nil)
		return

	}

}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) userLoginTrouble(w http.ResponseWriter, r *http.Request) {

	// Use the RenewToken() method on the current session to change the session
	// ID again.
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError500(w, r, err)
		return
	}
	// Remove the authenticatedUserID from the session data so that the user is
	// 'logged out'.
	app.RemoveSessionUser(r)

	data := app.newTemplateData(r)
	form := models.UserTroubleForm{}

	if r.Method == http.MethodPost {
		err := app.decodePostForm(r, &form)
		if err != nil {
			app.clientError(w, http.StatusBadRequest, err)
			return
		}
		form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
		form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
		if form.Valid() {
			switch form.Option {
			case "reverify":
				go func() { //goroutine
					defer concurrent.Recoverer("userLoginTrouble 1")
					defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

					user, err := app.users.GetByEmail(form.Email)
					if err == nil {
						emailRequest := app.users.BuildVerificationEmail(user, app.hostURL)
						app.SendEmail(emailRequest)
					}
				}()

			default:
				go func() { //goroutine
					defer concurrent.Recoverer("userLoginTrouble 2")
					defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))

					user, err := app.users.GetByEmail(form.Email)
					if err == nil {
						emailRequest := app.users.BuildPasswordEmail(user, app.hostURL)
						app.SendEmail(emailRequest)
					}
				}()
			}

			app.sessionManager.Put(r.Context(), "flash", "Please check your email")

		}
	}

	data.Form = form
	app.render(w, r, http.StatusOK, "account_forgot_password.tmpl", data)
}

// ------------------------------------------------------
//
// ------------------------------------------------------
func (app *application) UserOwnsEndPoint(w http.ResponseWriter, r *http.Request, endpointID string) bool {
	if endpointID == "" {
		return true
	}

	user, err := app.GetUser(r)
	if err != nil {
		app.UnauthorizedError(w, r)
		return false
	}

	if !user.OwnsEndPoint(endpointID) {
		app.notFound(w, err)
		return false
	}

	return true
}

// ------------------------------------------------------
func (app *application) EndPointOwnership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpointID := chi.URLParam(r, "endpointid")

		if !app.UserOwnsEndPoint(w, r, endpointID) {
			return
		}

		// And call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// ------------------------------------------------------
func (app *application) CanAddMoreEndpoints(r *http.Request) (bool, error) {
	// if app level is good.. check user level
	user, err := app.GetUser(r)
	if err != nil {

		return false, errors.New("User not found") // user not found
	}

	if user.IsSuperUser {
		return true, nil
	}

	return true, nil
}
