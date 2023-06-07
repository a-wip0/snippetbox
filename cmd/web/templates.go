package main

import (
	"html/template"
	"path/filepath"
	"snippetbox/internal/models"
	"time"
)

// Define a templateData type to act as the holding structure for
// any dynamic data that we want to pass to our HTML templates.
// At the moment it only contains one field, but we'll add more
// to it as the build progresses.
type templateData struct {
	CurrentYear int
	Snippet     *models.Snippet
	Snippets    []*models.Snippet
	Form        any
	Flash       string // Add a Flash field to the templateData struct.
}

// Create a humanDate function which returns a nicely formatted string
// representation of a time.Time object.
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

// Initialize a template.FuncMap object and store it in a global variable. This is
// essentially a string-keyed map which acts as a lookup between the names of our
// custom template functions and the functions themselves.
var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{} // Initialize an empty map to store templates.

	pages, err := filepath.Glob("./ui/html/pages/*.html") // Find all HTML files in the specified directory.
	if err != nil {
		return nil, err // Return an error if there was a problem with the file path.
	}

	for _, page := range pages { // Iterate over each HTML file found.
		name := filepath.Base(page) // Extract the base name of the file.

		// Create a new template with the given name and register custom functions.
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err // Return an error if there was a problem parsing the base template.
		}

		// Parse all HTML files in the specified directory and add them to the template set.
		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err // Return an error if there was a problem parsing the partial templates.
		}

		// Parse the current HTML file and add it to the template set.
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err // Return an error if there was a problem parsing the current page template.
		}

		cache[name] = ts // Add the template set to the cache using the file name as the key.
	}

	return cache, nil // Return the populated cache of templates.
}
