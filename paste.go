package main

import (
	"html/template"
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
