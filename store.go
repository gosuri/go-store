package store

import (
	"errors"
)

var (
	ErrKeyNotFound = errors.New("store: key not found")
	ErrEmptyKey    = errors.New("store: key is empty")
)

type Store interface {
	ReadWriter
	Lister
	MultiReadWriter
}

type Item interface {
	Key() string
	SetKey(string)
}

type Writer interface {
	Write(Item) error
}

type Reader interface {
	Read(Item) error
}

type ReadWriter interface {
	Reader
	Writer
}

type Lister interface {
	List(interface{}) error
}

type MultiReader interface {
	ReadMultiple(interface{}) error
}

type MultiWriter interface {
	WriteMultiple(items []Item) error
}

type MultiReadWriter interface {
	MultiReader
	MultiWriter
}

type Items interface {
	Item
}
