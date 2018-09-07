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
var defaultCfg = config{
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
	cfg := defaultCfg
	if err := readConfig(configPath, &cfg); err == nil {
		log.Println("Loaded config from:", configPath)
	} else if err == ErrNoConfigFound {
		log.Println("Not loading config file")
	} else {
		log.Println("Error during loading config:", err)
	}

	log.SetPrefix("[YEP] ")
	log.SetFlags(log.Flags() | log.Lshortfile)
	assets = packr.NewBox(compileAssets)

	srv := NewServer(&MemoryDB{}, cfg)

	srv.handleRoute("/", handleHome)
	srv.handleRoute("/api/new", handleAPINewPaste)

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
