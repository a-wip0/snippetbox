package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
)

// secureHeaders sets security headers on the response.
func secureHeaders(next http.Handler) http.Handler {
	// This function sets the following security headers:
	// - Content-Security-Policy: This header prevents the browser from loading resources from other domains.
	// - Referrer-Policy: This header prevents the browser from sending the referrer URL with cross-origin requests.
	// - X-Content-Type-Options: This header prevents the browser from MIME-sniffing a response.
	// - X-Frame-Options: This header prevents the browser from framing the page in an iframe.
	// - X-XSS-Protection: This header enables XSS protection.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the security headers.
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// `logRequest` logs the request information to the application's `infoLog` logger.
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request information.
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		// Serve the request using the next handler.
		next.ServeHTTP(w, r)
	})
}

// `recoverPanic` recovers from panics and sends a 500 Internal Server Error response to the client.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Deferred function that recovers from panics.
		defer func() {
			// If there was a panic, send a 500 Internal Server Error response to the client.
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		// Serve the request using the next handler.
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	// This function is used to require authentication for all requests that
	// are passed to the `next` handler. If the request is not authenticated,
	// the user will be redirected to the `/user/login` page.

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is authenticated.
		if !app.isAuthenticated(r) {
			// If the user is not authenticated, redirect them to the login page.
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		// Set the "Cache-Control" header to "no-store" to prevent the browser from caching
		// the response. This is important for security reasons, as it prevents an attacker
		// from caching a sensitive response and then serving it to other users.
		w.Header().Add("Cache-Control", "no-store")

		// Call the `next` handler to handle the request.
		next.ServeHTTP(w, r)
	})
}

func noSurf(next http.Handler) http.Handler {
	// This function is used to protect against cross-site request forgery (CSRF) attacks.
	// It does this by creating a CSRF token and adding it to the request. The CSRF token
	// is then validated by the `next` handler.

	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		// Set the `HttpOnly` flag to true to prevent the browser from exposing the CSRF token
		// to JavaScript. This is important for security reasons, as it prevents an attacker
		// from stealing the CSRF token and using it to perform unauthorized actions on the user's behalf.
		HttpOnly: true,
		// Set the `Path` to "/" to make the CSRF token valid for all requests on the application.
		Path: "/",
		// Set the `Secure` flag to true to require the CSRF token to be sent over a secure connection.
		Secure: true,
	})

	return csrfHandler
}

func (app *application) authenticate(next http.Handler) http.Handler {
	// This function is used to authenticate users before they are allowed to access
	// the `next` handler. If the user is not authenticated, they will be redirected
	// to the login page.

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the user ID from the session manager.
		id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

		// If the user ID is 0, then the user is not authenticated.
		if id == 0 {
			// In this case, we simply call the `next` handler to handle the request.
			next.ServeHTTP(w, r)
			return
		}

		// Check if the user exists in the database.
		exists, err := app.users.Exists(id)

		// If there is an error, then we handle it and return.
		if err != nil {
			app.serverError(w, err)
			return
		}

		// If the user exists, then we set the `isAuthenticated` context key to `true`.
		if exists {
			ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
			r = r.WithContext(ctx)
		}

		// Finally, we call the `next` handler to handle the request.
		next.ServeHTTP(w, r)
	})
}
