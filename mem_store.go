package store

import (
	"errors"
)

// MemItem represent the data structure
// used to store values in memory
type MemItem struct {
	prefix string
	key    string
	data   map[string]interface{}
}

type MemStore struct{}

var (
	ErrImplPending = errors.New("Implementation pending")
)

func (s *MemStore) Read(i Item) error {
	return ErrImplPending
}

func (s *MemStore) Write(i Item) error {
	return ErrImplPending
}

func (s *MemStore) List(i interface{}) error {
	return ErrImplPending
}

func (s *MemStore) ReadMultple(i interface{}) error {
	return ErrImplPending
}

func (s *MemStore) WriteMultple(i interface{}) error {
	return ErrImplPending
}
