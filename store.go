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

func ReadMultiple(items []Item) {
}

func WriteMultiple(items []Item) {
}

type Item interface {
	Key() string
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

type MultiReader interface {
	ReadMultiple(items []Item)
}

type MultiWriter interface {
	WriteMultiple(items []Item)
}

type MultiReadWriter interface {
	MultiReader
	MultiWriter
}

type RedisItem struct {
	data map[string][]byte
}

func (i *RedisItem) Save() error {
	return nil
}
