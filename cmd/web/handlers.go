package main

import (
	"errors"
	"fmt"
	"net/http"
	"snippetbox/internal/models"
	"snippetbox/internal/validator"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user's ID from the session.
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	if userID == 0 {
		// Redirect the user to the login page if they are not authenticated.
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Get the latest snippets from the database.
	snippets, err := app.snippets.Latest(userID)
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
	// Get the authenticated user's ID from the session.
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	if userID == 0 {
		// Redirect the user to the login page if they are not authenticated.
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Get the snippet ID from the request URL.
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		// If the ID is invalid, send a 404 error response.
		app.notFound(w)
		return
	}

	// Get the snippet from the database.
	snippet, err := app.snippets.Get(id, userID)
	if err != nil {
		// If there is an error, check if it is a no record error.
		if errors.Is(err, models.ErrNoRecord) {
			// If it is, send a 404 error response.
			// TODO: Create New View For Not Found
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

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	// Use the generic PermittedValue() function instead of the type-specific
	// PermittedInt() function.
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	// Validate the form data.
	if !form.Valid() {
		// If the form data is not valid, render the create.tmpl template with the form data.
		data := app.newTemplateData(r)
		data.Form = form

		app.render(w, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	// Insert the snippet into the database.
	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires, userID)
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

	// Validate the form contents using our helper functions.
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

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

type accountPasswordUpdateForm struct {
	CurrentPassword         string `form:"currentPassword"`
	NewPassword             string `form:"newPassword"`
	NewPasswordConfirmation string `form:"newPasswordConfirmation"`
	validator.Validator     `form:"-"`
}

func (app *application) accountPasswordUpdate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = accountPasswordUpdateForm{}

	app.render(w, http.StatusOK, "password.html", data)
}

func (app *application) accountPasswordUpdatePost(w http.ResponseWriter, r *http.Request) {
	var form accountPasswordUpdateForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.CurrentPassword), "currentPassword", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.NewPassword), "newPassword", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.NewPassword, 8), "newPassword", "This field must be at least 8 characters long")
	form.CheckField(validator.NotBlank(form.NewPasswordConfirmation), "newPasswordConfirmation", "This field cannot be blank")
	form.CheckField(form.NewPassword == form.NewPasswordConfirmation, "newPasswordConfirmation", "Passwords do not match")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form

		app.render(w, http.StatusUnprocessableEntity, "password.html", data)
		return
	}

	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	err = app.users.PasswordUpdate(userID, form.CurrentPassword, form.NewPassword)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddFieldError("currentPassword", "Current password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form

			app.render(w, http.StatusUnprocessableEntity, "password.html", data)
		} else if err != nil {
			app.serverError(w, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your password has been updated!")

	http.Redirect(w, r, "/account/view", http.StatusSeeOther)
}
