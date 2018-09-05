package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/chroma/styles"

	"github.com/alecthomas/chroma"

	htmlFormatter "github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
)

//Errors declaration for paste
var (
	ErrPasteTooBig        = fmt.Errorf("Paste too Big")
	ErrEmptyPaste         = fmt.Errorf("Empty Paste")
	ErrExpireTimeNotValid = fmt.Errorf("Expire time not valid")

	ErrNoConfigFound = fmt.Errorf("No config found")
)

//PasteNeverExpire is a constant used for indicating a paste that will never expire
const PasteNeverExpire = "Never"

type config struct {
	Addr           string
	TimeFormat     string
	DefaultName    string
	PathLen        int
	HighlightStyle string
	UndefinedLang  string
	Header         string
	AssetsDir      string
	ExpireAfter    []*pasteDuration
	MaxPasteSize   int //in bytes
}

type pasteDuration struct{ time.Duration }

//String implements fmt.Stringer
func (d *pasteDuration) String() string {
	m, _ := d.MarshalText()
	return string(m)
}

//MarshalText implements encoding.TextMarshaler
func (d *pasteDuration) MarshalText() ([]byte, error) {
	if d.Duration == 0 {
		return []byte(PasteNeverExpire), nil
	}
	return []byte(d.Duration.String()), nil
}

//UnmarshalText implements encoding.TextUnmarshaler
func (d *pasteDuration) UnmarshalText(text []byte) error {
	if string(text) == PasteNeverExpire {
		*d = pasteDuration{0}
		return nil
	}

	dur, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = pasteDuration{dur}
	return nil
}

func validateName(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return cfg.DefaultName, nil
	}
	return name, nil
}

func validateCode(code string) (string, error) {
	size := len(strings.TrimSpace(code))
	if size == 0 {
		return code, ErrEmptyPaste
	}
	if size > cfg.MaxPasteSize {
		return code, ErrPasteTooBig
	}

	return code, nil
}

func validateExpire(expire string) (*pasteDuration, error) {
	dur := &pasteDuration{}
	err := dur.UnmarshalText([]byte(expire))
	if err != nil {
		return dur, err
	}

	//Check if the parsed duration is in the config
	valid := false
	for _, d := range cfg.ExpireAfter {
		if dur == d {
			valid = true
		}
	}
	if !valid {
		return dur, ErrExpireTimeNotValid
	}

	return dur, nil
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

func readConfig(path string, cfg *config) error {
	if _, err := os.Stat(path); err != nil {
		return ErrNoConfigFound
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, cfg)
	if err != nil {
		return err
	}

	if len(cfg.ExpireAfter) == 0 {
		cfg.ExpireAfter = []*pasteDuration{&pasteDuration{0}}
		return fmt.Errorf("No Expire Time, defaulting to: %s", PasteNeverExpire)
	}
	return nil
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
