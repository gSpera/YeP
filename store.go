package main

import (
	"fmt"
)

//ErrDatabaseNotFound is an error used when the database could not find the value
var ErrDatabaseNotFound = fmt.Errorf("Value not found on the database")

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
