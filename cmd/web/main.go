package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log"
	"net/http"
	"os"
	"snippetbox/internal/models"
	"time"
)

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       models.SnippetModelInterface // Use our new interface type.
	users          models.UserModelInterface    // Use our new interface type.
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address") // Define a command-line flag for the HTTP network address.

	dsn := flag.String("dsn", "snippetbox_admin:snippetbox_admin_pass@/snippetbox?parseTime=true", "MySQL data source name") // Define a command-line flag for the MySQL data source name.

	flag.Parse() // Parse the command-line flags.

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)                  // Create a logger for informational logs.
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile) // Create a logger for error logs.

	db, err := openDB(*dsn) // Open a connection to the database using the provided data source name.
	if err != nil {
		errorLog.Fatal(err) // If there's an error, log it as a fatal error and terminate the application.
	}

	defer func(db *sql.DB) {
		err := db.Close() // Close the database connection when the main function exits.
		if err != nil {
			errorLog.Fatal(err) // If there's an error, log it as a fatal error and terminate the application.
		}
	}(db)

	templateCache, err := newTemplateCache() // Create a new template cache.
	if err != nil {
		errorLog.Fatal(err) // If there's an error, log it as a fatal error and terminate the application.
	}

	formDecoder := form.NewDecoder() // Create a form decoder for decoding form data.

	sessionManager := scs.New()               // Create a new session manager.
	sessionManager.Store = mysqlstore.New(db) // Set the session store to use a MySQL-based store.
	sessionManager.Lifetime = 12 * time.Hour  // Set the session lifetime to 12 hours.
	// Make sure that the Secure attribute is set on our session cookies.
	// Setting this means that the cookie will only be sent by a user's web
	// browser when an HTTPS connection is being used (and won't be sent over an
	// unsecure HTTP connection).
	sessionManager.Cookie.Secure = true

	app := &application{
		errorLog:       errorLog,                     // Set the error log for the application.
		infoLog:        infoLog,                      // Set the info log for the application.
		snippets:       &models.SnippetModel{DB: db}, // Create a snippet model with the database connection.
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,  // Set the template cache for the application.
		formDecoder:    formDecoder,    // Set the form decoder for the application.
		sessionManager: sessionManager, // Set the session manager for the application.
	}

	// Initialize a tls.Config struct to hold the non-default TLS settings we
	// want the server to use. In this case the only thing that we're changing
	// is the curve preferences value, so that only elliptic curves with
	// assembly implementations are used.
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", *addr) // Log an informational message indicating the server startup.
	// Use the ListenAndServeTLS() method to start the HTTPS server. We
	// pass in the paths to the TLS certificate and corresponding private key as
	// the two parameters.
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err) // If there's an error, log it as a fatal error and terminate the application.
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
