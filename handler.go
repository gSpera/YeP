package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

//Handle: /
func handleHome(s Server, w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		handlePostPaste(s, w, req)
		return
	}

	handleNewPaste(s, w, req)
}

//Handle: / GET
func handleNewPaste(s Server, w http.ResponseWriter, req *http.Request) {
	t, err := getTemplate("new")
	if err != nil {
		log.Println("Cannot get template: new", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal server error")
		return
	}

	err = t.Execute(w, struct {
		Langs         []string
		DefaultName   string
		Header        string
		ExpireTime    []*pasteDuration
		ExpireTimeLen int
	}{
		getLanguages(),
		cfg.DefaultName,
		cfg.Header,
		cfg.ExpireAfter,
		len(cfg.ExpireAfter),
	})

	if err != nil {
		log.Println("Error in executin template new:", err)
	}
}

//Handle: / POST
func handlePostPaste(s Server, w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		log.Println("Cannot parse form", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
	}
	name := req.PostForm.Get("name")
	code := req.PostForm.Get("code")
	lang := req.PostForm.Get("lang")
	expireTimeS := req.PostForm.Get("expire")

	name, err := validateName(name)
	if err != nil {
		handleError(w, req, err)
		return
	}
	code, err = validateCode(code)
	if err != nil {
		handleError(w, req, err)
		return
	}
	expireTime, err := validateExpire(expireTimeS)
	if err != nil {
		handleError(w, req, err)
		return
	}

	css, code, lang := highlightCode(code, lang)

	path := s.db.CreatePastePath(cfg.PathLen)
	paste := Paste{
		Path:    path,
		User:    name,
		Lang:    lang,
		Style:   template.CSS(css),
		Content: template.HTML(code),
		Created: time.Now(),
	}
	if name == "" {
		name = cfg.DefaultName
	}
	if err := s.db.Store(path, paste); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Could not post paste")
		log.Println("Could not paste paste", err)
		return
	}

	//If ExpireTime is 0 do not delete pastes
	if expireTime.Duration != 0 {
		time.AfterFunc(expireTime.Duration, func() {
			s.db.Delete(paste.Path)
		})
	}

	http.Redirect(w, req, path, http.StatusFound)
}

//Handle: /PASTE
func handleGetPaste(s Server, w http.ResponseWriter, req *http.Request) {
	t, err := getTemplate("paste")
	if err != nil {
		log.Println("Cannot get template: paste", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal server error")
		return
	}
	paste, err := s.db.Get(req.URL.Path[1:])
	if err == ErrDatabaseNotFound {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Could not find paste: %s", req.URL.Path)
		return
	}

	if err != nil {
		log.Println("Cannot get from Database", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal server error")
		return
	}

	if err := t.Execute(w, struct {
		Paste
		CreatedFormatted string
	}{
		paste,
		paste.Created.Format(cfg.TimeFormat),
	}); err != nil {
		log.Println("Cannot execute template:", err)
	}
}

func handleError(w http.ResponseWriter, req *http.Request, err error) {
	w.WriteHeader(http.StatusBadRequest)
	t, tErr := getTemplate("error")

	//Cannot get template
	if tErr != nil {
		log.Printf("Cannot get template while handling error:\nTemplate Error: %v\nError: %v\n", tErr, err)
		fmt.Fprintf(w, "Error: %v", err)
		return
	}

	log.Println(err)
	t.Execute(w, err)
}
