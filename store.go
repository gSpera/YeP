package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"time"
)

//ErrDatabaseNotFound is an error used when the database could not find the value
var ErrDatabaseNotFound = fmt.Errorf("Value not found on the database")

//Paste is a paste
type Paste struct {
	Path    string
	User    string
	Lang    string
	Style   template.CSS
	Content template.HTML
	Created time.Time
}

//Database is an interface for all the storage system
type Database interface {
	//Get returns the value associated with the name
	Get(name string) (Paste, error)
	//Store stores the contentent with the associated name
	Store(name string, value Paste) error

	//Delete deletes a paste from the Database
	Delete(name string)

	//CreatePastePath creates a path for the paste with the given length
	CreatePastePath(length int) string
}

//MemoryDB is a memory stored Database
//It has no persistence
type MemoryDB map[string]Paste

//Get implements Database
func (db MemoryDB) Get(name string) (Paste, error) {
	v, ok := db[name]
	if !ok {
		return Paste{}, ErrDatabaseNotFound
	}
	return v, nil
}

//Store implements Database
func (db MemoryDB) Store(name string, value Paste) error {
	db[name] = value
	return nil
}

//Delete implements Database
func (db MemoryDB) Delete(name string) {
	delete(db, name)
}

const alphabeth = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

//CreatePastePath implements Database
func (db MemoryDB) CreatePastePath(length int) string {
	var sPath string

	for {
		path := make([]rune, length)
		for i := range path {
			n := rand.Intn(len(alphabeth))
			path[i] = rune(alphabeth[n])
		}
		sPath = string(path)
		if _, ok := db[sPath]; !ok {
			break
		}
		log.Println("Found collision:", sPath)
	}

	return sPath
}
