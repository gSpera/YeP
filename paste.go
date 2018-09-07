package main

import (
	"html/template"
	"log"
	"time"
)

type pasteDuration struct{ time.Duration }

//Paste is a paste
type Paste struct {
	Path    string
	User    string
	Lang    string
	Style   template.CSS
	Content template.HTML
	Created time.Time
}

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

//NewPaste creates a new paste
func NewPaste(s *Server, name, code, lang string, expireTime *pasteDuration) (string, error) {

	name, err := validateName(name, s.cfg.DefaultName)
	if err != nil {
		return "", err
	}
	code, err = validateCode(code, s.cfg.MaxPasteSize)
	if err != nil {
		return "", err
	}

	css, code, lang := highlightCode(code, lang, s.cfg.UndefinedLang, s.cfg.HighlightStyle)

	path := s.db.CreatePastePath(s.cfg.PathLen)
	paste := Paste{
		Path:    path,
		User:    name,
		Lang:    lang,
		Style:   template.CSS(css),
		Content: template.HTML(code),
		Created: time.Now(),
	}
	if name == "" {
		name = s.cfg.DefaultName
	}
	if err := s.db.Store(path, paste); err != nil {
		log.Println("Could not paste paste", err)
		return "", err
	}

	//If ExpireTime is 0 do not delete pastes
	if expireTime.Duration != 0 {
		time.AfterFunc(expireTime.Duration, func() {
			s.db.Delete(paste.Path)
		})
	}

	return paste.Path, nil
}
