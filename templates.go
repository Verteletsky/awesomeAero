package main

import (
	"awesomeAero/models"
	"html/template"
	"path/filepath"
)

type templateData struct {
	Restaurants     []*models.RestTableResponse
	RestaurantTable []*models.RestTable
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob(filepath.Join(dir, "*.html"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}
