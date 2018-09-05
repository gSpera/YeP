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
	ExpireAfter:    []*pasteDuration{&pasteDuration{30 * time.Minute}},
	MaxPasteSize:   15000, //15KB
}

const (
	configPath    = "yep.json"
	compileAssets = "assets/"
)

var assets packr.Box

func main() {
	rand.Seed(time.Now().UnixNano())
	if err := readConfig(configPath, &cfg); err == nil {
		log.Println("Loaded config from:", configPath)
	} else if err == ErrNoConfigFound {
		log.Println("Not loading file")
	} else {
		log.Println("Error during loading config:", err)
	}

	log.SetFlags(log.Flags() | log.Lshortfile)
	assets = packr.NewBox(compileAssets)

	srv := NewServer(
		&MemoryDB{},
		map[string]Route{
			"/": home,
		},
	)

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

//NewServer creates a new server
func NewServer(db Database, routes map[string]Route) Server {
	s := Server{
		db:     db,
		routes: routes,
	}
	return s
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

func postPaste(s Server, w http.ResponseWriter, req *http.Request) {
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
