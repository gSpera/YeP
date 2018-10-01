package main

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

type badWriter struct{}

func (badWriter) Write([]byte) (int, error) {
	return 0, errors.New("Bad Writer")
}
func (badWriter) WriteHeader(int)     {}
func (badWriter) Header() http.Header { return nil }

type badReader struct{}

func (badReader) Read([]byte) (int, error) {
	return 0, errors.New("Bad Reader")
}

func getExpireTime(t *testing.T) string {
	if len(defaultCfg.ExpireAfter) == 0 {
		return "0"
	}
	res, err := defaultCfg.ExpireAfter[0].MarshalText()
	if err != nil {
		t.Fatalf("Could not get ExpireTime: %v", err)
	}
	return string(res)
}

//TestDB is a dummy database that intercepts some request used for testing.
type TestDB struct{ db Database }

func NewTestDB() *TestDB {
	return &TestDB{
		MemoryDB{},
	}
}

func (db *TestDB) Get(name string) (Paste, error) {
	if name == "test" {
		return Paste{
			Content: "<h1>test</h1>",
			Source:  "test",
			Created: time.Unix(0, 0),
			Expire:  time.Unix(0, 0),
			Style:   "",
			Path:    "test",
			User:    "test",
			Lang:    "test",
		}, nil
	}

	return db.db.Get(name)
}
func (db *TestDB) Store(name string, value Paste) error { return db.db.Store(name, value) }

func (db *TestDB) Delete(name string) { db.db.Delete(name) }

func (db *TestDB) CreatePastePath(length int) string { return db.db.CreatePastePath(length) }
