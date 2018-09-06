package main

import (
	"log"
	"math/rand"
)

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
