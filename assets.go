package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func getAsset(filename string) (http.File, error) {
	var file http.File
	var err error

	if _, err := os.Stat(cfg.AssetsDir + filename); err == nil {
		file, err = os.Open(cfg.AssetsDir + filename)
	} else if assets.Has(filename) {
		file, err = assets.Open(filename)
	} else {
		log.Println("Cannot find file:", filename)
		return nil, os.ErrNotExist
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
