package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
)

func getAsset(filename string) (http.File, error) {
	var file http.File
	var err error
	if _, err := os.Stat(cfg.AssetsDir + filename); err == nil {
		file, err = os.Open(cfg.AssetsDir + filename)
	} else {
		file, err = assets.Open(filename)
	}

	return file, err
}

func getTemplate(name string) (*template.Template, error) {
	file, err := getAsset(name + ".tmpl")
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return template.New(name).Parse(string(content))
}
