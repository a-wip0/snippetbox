package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

// routes() defines the routes for the application.
func (app *application) routes() http.Handler {
	// Create a new router.
	router := httprouter.New()

	// Set the NotFound handler to call the app.notFound() function.
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	// Create a new file server for serving static files.
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Serve static files under the /static path.
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	// Create a new middleware chain for handling dynamic requests.
	dynamic := alice.New(app.sessionManager.LoadAndSave)

	// Handle GET requests to the root path.
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))

	// Handle GET requests to the /snippet/view/:id path.
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))

	// Handle GET requests to the /snippet/create path.
	router.Handler(http.MethodGet, "/snippet/create", dynamic.ThenFunc(app.snippetCreate))

	// Handle POST requests to the /snippet/create path.
	router.Handler(http.MethodPost, "/snippet/create", dynamic.ThenFunc(app.snippetCreatePost))

	// Create a new middleware chain for handling standard requests.
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Return a new handler that chains the standard and dynamic middleware chains.
	return standard.Then(router)
}
