package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
)

// serverError writes a server error to the response writer.
// The error is logged to the application's error log.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Println(trace)

	if app.debug {
		http.Error(w, trace, http.StatusInternalServerError)
		return
	}

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// clientError writes a client error to the response writer.
// The error code is passed as an argument.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// notFound writes a 404 Not Found error to the response writer.
func (app *application) notFound(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusNotFound, "notfound.html", app.newTemplateData(r))
}

// render renders a template with the given data.
// The template is looked up in the application's template cache.
// If the template does not exist, a server error is returned.
func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	// Create a new buffer to store the rendered template.
	buf := new(bytes.Buffer)

	// Execute the template, passing in the data.
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Write the HTTP response header.
	w.WriteHeader(status)

	// Write the rendered template to the response writer.
	_, err = buf.WriteTo(w)
	if err != nil {
		return
	}
}

// newTemplateData creates a new template data object.
// The object contains the current year, the flash message,
// the user's authentication status, and the CSRF token.
func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
		CSRFToken:       nosurf.Token(r),
	}
}

// decodePostForm decodes the POST form data from the request into the given destination object.
// The destination object must be a struct.
func (app *application) decodePostForm(r *http.Request, dst any) error {
	// Parse the POST form data.
	err := r.ParseForm()
	if err != nil {
		return err
	}

	// Decode the POST form data into the destination object.
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		// If the error is a `form.InvalidDecoderError`, panic.
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		return err
	}

	// Return nil to indicate success.
	return nil
}

// isAuthenticated returns whether the user is authenticated.
// The authentication status is stored in the request context.
func (app *application) isAuthenticated(r *http.Request) bool {
	// Get the authentication status from the request context.
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		// If the authentication status is not present in the request context, return false.
		return false
	}

	// Return the authentication status.
	return isAuthenticated
}
