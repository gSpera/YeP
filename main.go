package main

import (
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gobuffalo/packr"
)

//Default Config
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
			"/": handleHome,
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
	mux    *http.ServeMux
}

//NewServer creates a new server
func NewServer(db Database, routes map[string]Route) Server {
	s := Server{
		db:     db,
		routes: routes,
		mux:    http.NewServeMux(),
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

	s.mux.ServeHTTP(w, req)
	// handleGetPaste(s, w, req)
}
