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
	)

	srv.handleRoute("/", handleHome)

	for _, filename := range assets.List() {
		//Do not return templates
		if strings.HasSuffix(filename, ".tmpl") {
			continue
		}
		srv.mux.HandleFunc("/static/"+filename, routeToHandler(handlePackrFile(filename), &srv))
	}

	log.Println("Listening on", cfg.Addr)
	http.ListenAndServe(cfg.Addr, srv)
}

//Server is a YeP server
//Implements http.Handler
type Server struct {
	db  Database
	mux *http.ServeMux
}

//NewServer creates a new server
func NewServer(db Database) Server {
	s := Server{
		db:  db,
		mux: http.NewServeMux(),
	}
	return s
}

func (s *Server) handleRoute(pattern string, r Route) {
	s.mux.HandleFunc(pattern, routeToHandler(r, s))
}

//Route is a Server Route
type Route func(s Server, w http.ResponseWriter, req *http.Request)

func (s Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.mux.ServeHTTP(w, req)
}
