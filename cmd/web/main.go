package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"snippetbox/internal/models"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

// The application struct represents the application itself. It contains
// references to the various components that the application needs, such as
// the logger, the database, the template cache, and the session manager.
type application struct {
	debug          bool
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       models.SnippetModelInterface
	users          models.UserModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

// The main function is the entry point for the application. It starts by
// parsing the command-line flags, which specify the HTTP network address and
// the MySQL data source name. It then creates a new logger, opens the database,
// creates a new template cache, creates a new form decoder, and creates a new
// session manager. Finally, it creates a new application instance, starts the
// HTTP server, and waits for the server to terminate.
func main() {
	// Parse the command-line flags.
	addr := flag.String("addr", os.Getenv("APP_PORT"), "HTTP network address")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))

	debug := flag.Bool("debug", false, "Enable debug mode")

	flag.Parse()

	// Create a new logger.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Open the database.
	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	// Create a new template cache.
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// Create a new form decoder.
	formDecoder := form.NewDecoder()

	// Create a new session manager.
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	// Create a new application instance.
	app := &application{
		debug:          *debug,
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	// Start the HTTP server.
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Initialize a new http.Server struct.
	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start the server.
	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

// The openDB function opens a new database connection using the specified
// MySQL data source name.
func openDB(dsn string) (*sql.DB, error) {
	// Open the database connection.
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Ping the database to make sure it's alive.
	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Return the database connection.
	return db, nil
}
