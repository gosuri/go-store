package store

import (
	"testing"
)

type TestStruct struct {
	key   string
	Field string
}

func (s *TestStruct) Key() string {
	return s.key
}

func TestExample(t *testing.T) {
	i := &TestStruct{}
	s := NewRedisStore()
	s.Write(i)
	s.Read(i)
}
