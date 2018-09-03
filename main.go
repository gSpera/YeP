package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gobuffalo/packr"
)

//Config
var cfg = config{
	Addr:           ":8080",
	TimeFormat:     "2 Jan 2006 15:04:05",
	DefaultName:    "Anonymous",
	PathLen:        5,
	HighlightStyle: "dracula",
	UndefinedLang:  "Undefined",
	Header:         "Yep Another Pastebin",
	AssetsDir:      "assets/",
}

const (
	configPath    = "yep.json"
	compileAssets = "assets/"
)

var assets packr.Box

func main() {
	rand.Seed(time.Now().UnixNano())
	if ok := readConfig(configPath, &cfg); ok {
		log.Println("Loaded config from:", configPath)
	}

	log.SetFlags(log.Flags() | log.Lshortfile)
	assets = packr.NewBox(compileAssets)

	srv := Server{
		db: &MemoryDB{},
		routes: map[string]Route{
			"/": home,
		},
	}

	for _, filename := range assets.List() {
		//Do not return templates
		if strings.HasSuffix(filename, ".tmpl") {
			continue
		}
		srv.routes["/static/"+filename] = handlePackrFile(filename)
	}

	log.Println("Listening on", cfg.Addr)
	http.ListenAndServe(cfg.Addr, srv)
}

//Server is a YeP server
//Implements http.Handler
type Server struct {
	db     Database
	routes map[string]Route
}

//Route is a Server Route
type Route func(s Server, w http.ResponseWriter, req *http.Request)

func (s Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rout, ok := s.routes[req.URL.Path]
	if ok {
		rout(s, w, req)
		return
	}

	getPaste(s, w, req)
}

func home(s Server, w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		postPaste(s, w, req)
		return
	}

	newPaste(s, w, req)
}
func newPaste(s Server, w http.ResponseWriter, req *http.Request) {
	t, err := getTemplate("new")
	if err != nil {
		log.Println("Cannot get template: new", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal server error")
		return
	}

	t.Execute(w, struct {
		Langs       []string
		DefaultName string
		Header      string
	}{
		getLanguages(),
		cfg.DefaultName,
		cfg.Header,
	})
}

func postPaste(s Server, w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		log.Println("Cannot parse form", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
	}
	name := req.PostForm.Get("name")
	code := req.PostForm.Get("code")
	lang := req.PostForm.Get("lang")
	name = validateName(name)
	code = validateCode(code)
	css, code, lang := highlightCode(code, lang)

	path, id := createPastePathAndID()
	paste := Paste{
		ID:      id,
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

	req.Method = "GET"
	http.Redirect(w, req, path, http.StatusTemporaryRedirect)
}

func getPaste(s Server, w http.ResponseWriter, req *http.Request) {
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
