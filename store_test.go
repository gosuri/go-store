package store

import (
	"testing"
)

type TestItem struct {
	key   string
	Field string
}

func (i *TestItem) Key() string {
	return i.key
}

func (i *TestItem) SetKey(key string) {
	i.key = key
}

func TestExample(t *testing.T) {
	// i := &TestStruct{}
	// s := NewRedisStore()
	// s.Write(i)
	// s.Read(i)
}
