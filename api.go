package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

//Error strings for the APIs
const (
	ErrInternalServerError = "Internal Server Error"
	ErrMethodNotAllowed    = "Method not allowed"
	ErrCannotDecodeJSON    = "Invalid JSON"
)

type newPasteRequest struct {
	Name       string
	Code       string
	Lang       string
	ExpireTime string
}

type newPasteResponse struct {
	OK    bool
	Error string
	Path  string
}

type getPasteRequest struct {
	Name   string
	Render bool
}

type getPasteResponse struct {
	OK      bool
	Error   string
	Name    string
	Code    string
	Render  string
	Style   string
	Created int64
	Expire  int64
	User    string
}

func handleAPINewPaste(s Server, w http.ResponseWriter, req *http.Request) {
	res := newPasteResponse{}
	var path string
	var err error
	var body []byte
	paste := newPasteRequest{}
	duration := &pasteDuration{}

	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		res = newPasteResponse{
			OK:    false,
			Error: ErrMethodNotAllowed,
		}
		goto response
	}

	body, err = ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Cannot read body:", err)
		res = newPasteResponse{
			OK:    false,
			Error: ErrInternalServerError,
		}
		goto response
	}

	if err := json.Unmarshal(body, &paste); err != nil {
		res = newPasteResponse{
			OK:    false,
			Error: ErrCannotDecodeJSON,
		}
		goto response
	}

	if err := duration.UnmarshalText([]byte(paste.ExpireTime)); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res = newPasteResponse{
			OK:    false,
			Error: ErrExpireTimeNotValid.Error(),
		}
		goto response
	}
	path, err = NewPaste(&s, paste.Name, paste.Code, paste.Lang, duration)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Cannot create paste:", err)
		res = newPasteResponse{
			OK:    false,
			Error: ErrInternalServerError,
		}
		goto response
	}

	res = newPasteResponse{
		OK:   true,
		Path: path,
	}

response:
	response, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Cannot create response")
		log.Println("Cannot marshal newPasteResponse:", err)
		return
	}
	if _, err := w.Write(response); err != nil {
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
