package main

import (
	"net/http"
	"snippetbox/ui"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	// The `router` variable is used to create a router.
	router := httprouter.New()

	// The `NotFound` function is called when a request is made to a route that does not exist.
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w, r)
	})

	// The `fileServer` variable is used to serve static files.
	fileServer := http.FileServer(http.FS(ui.Files))
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	// The `ping` function is used to test the application.
	router.HandlerFunc(http.MethodGet, "/ping", ping)

	// The `dynamic` variable is used to chain middleware functions to the router.
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	// The `home` function is used to handle requests to the home page.
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))

	router.Handler(http.MethodGet, "/about", dynamic.ThenFunc(app.about))

	// The `snippetView` function is used to handle requests to view a snippet.
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))

	// The `userSignup` function is used to handle requests to sign up for a user account.
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignup))

	// The `userSignupPost` function is used to handle requests to post a user signup form.
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignupPost))

	// The `userLogin` function is used to handle requests to log in to a user account.
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))

	// The `userLoginPost` function is used to handle requests to post a user login form.
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))

	// The `protected` variable is used to add additional middleware functions to the router.
	protected := dynamic.Append(app.requireAuthentication)

	// The `snippetCreate` function is used to handle requests to create a snippet.
	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(app.snippetCreate))

	// The `snippetCreatePost` function is used to handle requests to post a snippet create form.
	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(app.snippetCreatePost))

	router.Handler(http.MethodGet, "/account/view", protected.ThenFunc(app.accountView))

	router.Handler(http.MethodGet, "/account/password/update", protected.ThenFunc(app.accountPasswordUpdate))

	router.Handler(http.MethodPost, "/account/password/update", protected.ThenFunc(app.accountPasswordUpdatePost))

	// The `userLogoutPost` function is used to handle requests to logout of a user account.
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))

	// The `standard` variable is used to add additional features to the router.
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// The `return` statement returns the router.
	return standard.Then(router)
}
