package main

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"snippetbox/internal/models"
	"snippetbox/internal/validator"
	"strconv"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Get the latest snippets from the database.
	snippets, err := app.snippets.Latest()
	if err != nil {
		// If there is an error, send a 500 error response.
		app.serverError(w, err)
		return
	}

	// Create a new template data object.
	data := app.newTemplateData(r)

	// Set the snippets on the template data object.
	data.Snippets = snippets

	// Render the home.html template with the template data object.
	app.render(w, http.StatusOK, "home.html", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	// Get the snippet ID from the request URL.
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		// If the ID is invalid, send a 404 error response.
		app.notFound(w)
		return
	}

	// Get the snippet from the database.
	snippet, err := app.snippets.Get(id)
	if err != nil {
		// If there is an error, check if it is a no record error.
		if errors.Is(err, models.ErrNoRecord) {
			// If it is, send a 404 error response.
			app.notFound(w)
		} else {
			// If it is not, send a 500 error response.
			app.serverError(w, err)
		}
		return
	}

	// Create a new template data object.
	data := app.newTemplateData(r)

	// Set the snippet on the template data object.
	data.Snippet = snippet

	// Render the view.html template with the template data object.
	app.render(w, http.StatusOK, "view.html", data)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	// Create a new template data object.
	data := app.newTemplateData(r)

	// Set the form expiration on the template data object.
	data.Form = snippetCreateForm{
		Expires: 365,
	}

	// Render the create.html template with the template data object.
	app.render(w, http.StatusOK, "create.html", data)
}

// snippetCreateForm defines the form data for creating a snippet.
type snippetCreateForm struct {
	// Title is the title of the snippet.
	Title string `form:"title"`

	// Content is the content of the snippet.
	Content string `form:"content"`

	// Expires is the number of days after which the snippet will expire.
	Expires int `form:"expires"`

	// Validator is a validator that is used to validate the form data.
	validator.Validator `form:"-"`
}

// snippetCreatePost creates a new snippet from the form data submitted in the request.
func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	// Create a new form.
	var form snippetCreateForm

	// Decode the form data from the request.
	err := app.decodePostForm(r, &form)
	if err != nil {
		// If there is an error decoding the form data, return a 400 Bad Request error.
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form data.
	if !form.Valid() {
		// If the form data is not valid, render the create.tmpl template with the form data.
		data := app.newTemplateData(r)
		data.Form = form

		app.render(w, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	// Insert the snippet into the database.
	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		// If there is an error inserting the snippet into the database, return a 500 Internal Server Error error.
		app.serverError(w, err)
		return
	}

	// Set a flash message indicating that the snippet was created successfully.
	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	// Redirect the user to the snippet view page.
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

// userSignupForm defines the fields for a user signup form.
type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

// userSignup renders the signup form template.
func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	// Create a new template data object.
	data := app.newTemplateData(r)

	// Set the form object on the template data.
	data.Form = userSignupForm{}

	// Render the signup form template.
	app.render(w, http.StatusOK, "signup.html", data)
}

// userSignupPost handles the POST request for the signup form.
func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	// Create a new userSignupForm object.
	var form userSignupForm

	// Decode the POST form data into the userSignupForm object.
	err := app.decodePostForm(r, &form)
	if err != nil {
		// If there was an error decoding the form data, return a bad request error.
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the userSignupForm object.
	if !form.Valid() {
		// If the form is not valid, render the signup form template with the errors.
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	// Insert the user into the database.
	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		// If there was an error inserting the user into the database, handle the error.
		if errors.Is(err, models.ErrDuplicateEmail) {
			// If the email address is already in use, add an error to the form.
			form.AddFieldError("email", "Email address is already in use")

			// Render the signup form template with the error.
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
			return
		} else {
			// If the error is not a duplicate email error, return a server error.
			app.serverError(w, err)
			return
		}
	}

	// Set a flash message indicating that the signup was successful.
	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")

	// Redirect the user to the login page.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// userLoginForm defines the fields for a user login form.
type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

// userLogin renders the login form template.
func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	// Create a new template data object.
	data := app.newTemplateData(r)

	// Set the form object on the template data.
	data.Form = userLoginForm{}

	// Render the login form template.
	app.render(w, http.StatusOK, "login.html", data)
}

// userLoginPost handles the POST request for the login form.
func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	// Create a new userLoginForm object.
	var form userLoginForm

	// Decode the POST form data into the userLoginForm object.
	err := app.decodePostForm(r, &form)
	if err != nil {
		// If there was an error decoding the form data, return a bad request error.
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the userLoginForm object.
	if !form.Valid() {
		// If the form is not valid, render the login form template with the errors.
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	// Authenticate the user with the email and password.
	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		// If the authentication failed, handle the error.
		if errors.Is(err, models.ErrInvalidCredentials) {
			// If the email or password is incorrect, add an error to the form.
			form.AddNonFieldError("Email or password is incorrect")

			// Render the login form template with the error.
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.html", data)
			return
		} else {
			// If the error is not an invalid credentials error, return a server error.
			app.serverError(w, err)
			return
		}
	}

	// Renew the user's session token.
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		// If there was an error renewing the session token, return a server error.
		app.serverError(w, err)
		return
	}

	// Set the authenticated user ID on the session.
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	// Redirect the user to the snippet create page.
	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

// userLogoutPost logs the user out of the application.
func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	// Renew the user's session token.
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		// If an error occurs, return a server error.
		app.serverError(w, err)
		return
	}

	// Remove the authenticated user ID from the session.
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	// Set a flash message indicating that the user has been logged out successfully.
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")

	// Redirect the user to the homepage.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "about.html", data)
}

func (app *application) accountView(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	user, err := app.users.Get(userID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.User = user

	app.render(w, http.StatusOK, "account.html", data)
}

// ping returns an HTTP 200 OK response.
func ping(w http.ResponseWriter, r *http.Request) {
	// Write the "OK" string to the response.
	_, err := w.Write([]byte("OK"))
	if err != nil {
		// If an error occurs, return.
		return
	}
}
