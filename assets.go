package main

import (
	"html/template"
)

func getTemplate(name string) (*template.Template, error) {
	return template.New(name).Parse(assets.String(name + ".tmpl"))
}
