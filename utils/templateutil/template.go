package templateutil

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func NewTemplateCache() (map[string]*template.Template, error) {
	fsys := os.DirFS("templates")

	cache := map[string]*template.Template{}
	pages, err := fs.Glob(fsys, "*.html")
	if err != nil {
		return nil, err
	}
	err = loadPages(cache, pages)

	// pages, err = fs.Glob(fsys, "*/*.html")
	// if err != nil {
	// 	return nil, err
	// }
	// loadPages(cache, pages)

	// pages, err = fs.Glob(ui.Files, "html/emails/*.tmpl")
	// if err != nil {
	// 	return nil, err
	// }
	// loadPages(cache, pages)

	// pages, err = fs.Glob(ui.Files, "html/public/*.tmpl")
	// if err != nil {
	// 	return nil, err
	// }
	// loadPages(cache, pages)

	return cache, err
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func loadPages(cache map[string]*template.Template, pages []string) error {
	// Use fs.Glob() to get a slice of all filepaths in the ui.Files embedded
	// filesystem which match the pattern 'html/pages/*.tmpl'. This essentially
	// gives us a slice of all the 'page' templates for the application, just
	// like before.
	fsys := os.DirFS("templates")

	for _, page := range pages {
		name := filepath.Base(page)
		// Create a slice containing the filepath patterns for the templates we
		// want to parse.

		baseTemplateName := ""

		if strings.Contains(page, "_") {
			splitName := strings.Split(page, "_")
			baseTemplateName = fmt.Sprintf("%s.html", splitName[0])
		}

		patterns := make([]string, 0)

		if baseTemplateName != "" {
			patterns = append(patterns, baseTemplateName)

		}

		patterns = append(patterns, page)

		// Use ParseFS() instead of ParseFiles() to parse the template files
		// from the ui.Files embedded filesystem.
		ts, err := template.New(name).Funcs(functions).ParseFS(fsys, patterns...)

		// in case of error just skip the template
		if err == nil {
			cache[name] = ts
		} else {
			log.Println("Template: ", name, " failed. Due to:", err)
		}

	}
	return nil

}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func TemplateToString(templateCache map[string]*template.Template, page string, data any) (string, error) {
	ts, ok := templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)

		return "", err
	}
	// Initialize a new buffer.
	buf := new(bytes.Buffer)

	baseTemplateName := ""

	if strings.Contains(page, "_") {
		splitName := strings.Split(page, "_")
		baseTemplateName = splitName[0]
	}

	var err error = nil

	if baseTemplateName == "" {
		err = ts.Execute(buf, data)
	} else {
		err = ts.ExecuteTemplate(buf, baseTemplateName, data)
	}

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// -----------------------------------------------------------------
//
// -----------------------------------------------------------------
func TemplateToBuffer(templateCache map[string]*template.Template, page string, data any) (*bytes.Buffer, error) {
	ts, ok := templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)

		return nil, err
	}
	// Initialize a new buffer.
	buf := new(bytes.Buffer)

	baseTemplateName := ""

	if strings.Contains(page, "_") {
		splitName := strings.Split(page, "_")
		baseTemplateName = splitName[0]
	}

	var err error = nil

	if baseTemplateName == "" {
		err = ts.Execute(buf, data)
	} else {
		err = ts.ExecuteTemplate(buf, baseTemplateName, data)
	}

	if err != nil {
		return nil, err
	}

	return buf, nil
}
