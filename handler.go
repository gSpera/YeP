package main

import (
	"fmt"
	"log"
	"net/http"
)

//Handle: /
//Transfer to: /PASTE
//Transfer to: / GET
//Transfer to: / POST
func handleHome(s Server, w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		handleGetPaste(s, w, req)
		return
	}

	if req.Method == "POST" {
		handlePostPaste(s, w, req)
		return
	}

	handleNewPaste(s, w, req)
}

//Handle: / GET
func handleNewPaste(s Server, w http.ResponseWriter, req *http.Request) {
	t, err := getTemplate(s.cfg.AssetsDir, "new")
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
		s.cfg.DefaultName,
		s.cfg.Header,
		s.cfg.ExpireAfter,
		len(s.cfg.ExpireAfter),
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

	expireTime, err := validateExpire(expireTimeS, s.cfg.ExpireAfter)
	if err != nil {
		handleError(w, req, s.cfg.AssetsDir, err)
		return
	}

	path, err := NewPaste(&s, name, code, lang, expireTime)
	if err != nil {
		handleError(w, req, s.cfg.AssetsDir, err)
		return
	}

	http.Redirect(w, req, path, http.StatusFound)
}

//Handle: /PASTE
func handleGetPaste(s Server, w http.ResponseWriter, req *http.Request) {
	t, err := getTemplate(s.cfg.AssetsDir, "paste")
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
		paste.Created.Format(s.cfg.TimeFormat),
	}); err != nil {
		log.Println("Cannot execute template:", err)
	}
}

func handleError(w http.ResponseWriter, req *http.Request, assetsDir string, err error) {
	w.WriteHeader(http.StatusBadRequest)
	t, tErr := getTemplate(assetsDir, "error")

	//Cannot get template
	if tErr != nil {
		log.Printf("Cannot get template while handling error:\nTemplate Error: %v\nError: %v\n", tErr, err)
		fmt.Fprintf(w, "Error: %v", err)
		return
	}

	log.Println(err)
	t.Execute(w, err)
}
