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
	//MultiReadWriter
}

var DefaultStore = NewRedisStore()

func Read(i Item) {
}

func Write(i Item) {
}

func ReadMultiple(items []Item, key string) {
}

func WriteMultiple(items []Item) {
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
	ListAll() error
}

type MultiReader interface {
	ReadAll(items []Item)
	ReadMultiple(items []Item)
}

type MultiWriter interface {
	WriteMultiple(items []Item)
}

type MultiReadWriter interface {
	MultiReader
	MultiWriter
}

func (i *RedisItem) Save() error {
	return nil
}
