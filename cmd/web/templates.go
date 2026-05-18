package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"snippetbox/internal/models"
	"snippetbox/ui"
	"time"
)

// `templateData` struct contains all the data that will be passed to the template.
type templateData struct {
	Snippet         *models.SnippetGet
	CurrentYear     int
	Snippets        []*models.SnippetLatest
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
	User            *models.User
}

// `humanDate` function formats a `time.Time` object into a human-readable string.
func humanDate(t time.Time) string {
	// If the `time.Time` object is empty, return an empty string.
	if t.IsZero() {
		return ""
	}

	// Return a formatted string of the `time.Time` object.
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// `functions` is a template function map that contains the `humanDate` function.
var functions = template.FuncMap{
	"humanDate": humanDate,
}

// `newTemplateCache` function creates a cache of templates.
func newTemplateCache() (map[string]*template.Template, error) {
	// `cache` is a map that stores the templates by their name.
	cache := map[string]*template.Template{}

	// `pages` is a slice of the paths to all the HTML pages.
	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		// If an error occurs, return an error.
		return nil, err
	}

	// Iterate over the pages.
	for _, page := range pages {
		// Get the name of the page.
		name := filepath.Base(page)

		// Get the patterns for the templates that should be parsed for the page.
		patterns := []string{
			"html/base.html",
			"html/partials/*.html",
			page,
		}

		// Create a new template with the given name and functions.
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			// If an error occurs, return an error.
			return nil, err
		}

		// Add the template to the cache.
		cache[name] = ts
	}

	// Return the cache.
	return cache, nil
}
