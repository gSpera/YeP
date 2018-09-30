package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type newPasteRequest struct {
	Name       string
	Code       string
	Lang       string
	ExpireTime string
}

type newPasteResponse struct {
	OK   bool
	Path string
}

type getPasteRequest struct {
	Name   string
	Render bool
}

type getPasteResponse struct {
	OK      bool
	Name    string
	Code    string
	Render  string
	Style   string
	Created int64
	Expire  int64
	User    string
}

func handleAPINewPaste(s Server, w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Cannot read body:", err)
	}

	paste := newPasteRequest{}
	if err := json.Unmarshal(body, &paste); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Cannot read JSON")
		return
	}

	duration := &pasteDuration{}
	if err := duration.UnmarshalText([]byte(paste.ExpireTime)); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Cannot parse ExpireTime")
		return
	}
	path, err := NewPaste(&s, paste.Name, paste.Code, paste.Lang, duration)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Cannot create paste")
		log.Println("Cannot create paste:", err)
		return
	}

	res, err := json.Marshal(newPasteResponse{
		OK:   true,
		Path: path,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Cannot create response")
		log.Println("Cannot marshal newPasteResponse:", err)
		return
	}
	if _, err := w.Write(res); err != nil {
		log.Println("Cannot write response:", err)
		return
	}
}

func handleAPIGetPaste(s Server, w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Cannot read body:", err)
	}

	request := getPasteRequest{}
	if err := json.Unmarshal(body, &request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Cannot read JSON")
		return
	}

	paste, err := s.db.Get(request.Name)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "Paste not found")
		return
	}
	render := ""
	style := ""
	if request.Render {
		render = string(paste.Content)
		style = string(paste.Style)
	}

	res, err := json.Marshal(getPasteResponse{
		OK:      true,
		Name:    request.Name,
		Code:    paste.Source,
		Render:  render,
		Style:   style,
		Created: paste.Created.Unix(),
		User:    paste.User,
		Expire:  paste.Expire.UnixNano(),
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Cannot create response")
		log.Println("Cannot marshal newPasteResponse:", err)
		return
	}
	if _, err := w.Write(res); err != nil {
		log.Println("Cannot write response:", err)
		return
	}
}
