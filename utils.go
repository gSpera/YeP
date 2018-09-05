package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/chroma/styles"

	"github.com/alecthomas/chroma"

	htmlFormatter "github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
)

type config struct {
	Addr           string
	TimeFormat     string
	DefaultName    string
	PathLen        int
	HighlightStyle string
	UndefinedLang  string
	Header         string
	AssetsDir      string
	ExpireAfter    time.Duration
}

func validateName(name string) string {
	if strings.TrimSpace(name) == "" {
		return cfg.DefaultName
	}
	return name
}

func validateCode(code string) string {
	return code
}

//hightlightCode formattes the code string passed and returns the css, code highlight in HTML and the language
func highlightCode(code string, lang string) (string, string, string) {
	var lex chroma.Lexer
	if lang != "" {
		lex = lexers.Get(lang)
	}

	//Cannot get selected language, analysing
	if lex == nil {
		lex = lexers.Analyse(code)
	}
	//Cannot Analyse lang, use fallback
	if lex == nil {
		lex = lexers.Fallback
	}

	lang = lex.Config().Name
	//If cannot find lang
	if lex == lexers.Fallback {
		lang = cfg.UndefinedLang
	}
	lex = chroma.Coalesce(lex)
	style := styles.Get(cfg.HighlightStyle)
	if style == nil {
		style = styles.Fallback
	}
	form := htmlFormatter.New(
		htmlFormatter.WithClasses(),
		htmlFormatter.WithLineNumbers(),
		htmlFormatter.LineNumbersInTable(),
	)

	it, err := lex.Tokenise(nil, code)
	if err != nil {
		return "", code, ""
	}
	buf := new(bytes.Buffer)
	if err := form.Format(buf, style, it); err != nil {
		return "", code, ""
	}
	code = buf.String()
	buf.Reset()
	if err := form.WriteCSS(buf, style); err != nil {
		return "", code, ""
	}
	css := buf.String()

	return css, code, lang
}

func getLanguages() []string {
	return lexers.Names(false)
}

func handlerToRoute(h http.Handler) Route {
	return func(s Server, w http.ResponseWriter, req *http.Request) {
		h.ServeHTTP(w, req)
	}
}

func handlePackrFile(filename string) Route {
	return func(s Server, w http.ResponseWriter, req *http.Request) {
		file, err := getAsset(filename)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Cannot find file: %s", filename)
			return
		}

		_, err = io.Copy(w, file)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, "Internal Server Error")
			return
		}
	}
}

func readConfig(path string, cfg *config) bool {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return false
	}
	err = json.Unmarshal(content, cfg)
	if err != nil {
		return false
	}
	return true
}
