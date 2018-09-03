package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/alecthomas/chroma/styles"

	"github.com/alecthomas/chroma"

	htmlFormatter "github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
)

func validateName(name string) string {
	if strings.TrimSpace(name) == "" {
		return defaultName
	}
	return name
}

func validateCode(code string) string {
	return code
}

var currentPathNum = 0

const alphabeth = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var paths = map[string]struct{}{}

func createPastePathAndID() (string, int) {
	defer func() { currentPathNum++ }()
	var sPath string
	for {
		path := make([]rune, pathLen)
		for i := range path {
			path[i] = rune(alphabeth[rand.Intn(len(alphabeth))])
		}
		sPath = string(path)
		_, ok := paths[sPath]
		if !ok {
			paths[sPath] = struct{}{}
			break
		}
		log.Println("Found collision:", sPath)
	}

	return sPath, currentPathNum
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
		lang = undefinedLang
	}
	lex = chroma.Coalesce(lex)
	style := styles.Get(highlightStyle)
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
		fl, err := assets.Open(filename)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Cannot find file: %s", filename)
			return
		}
		_, err = io.Copy(w, fl)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, "Internal Server Error")
			return
		}
	}
}
