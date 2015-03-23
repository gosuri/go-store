package store

import (
	"testing"
)

type RedisTestStruct struct {
	Id         string
	Field      string
	FieldFloat float32
	FieldInt   int
	FieldBool  bool
	FieldUint  uint
}

func (s *RedisTestStruct) Key() string {
	return s.Id
}

func (s *RedisTestStruct) SetKey(k string) {
	s.Id = k
}

func TestWrite(t *testing.T) {
	s := &RedisTestStruct{
		Field:    "value",
		FieldInt: 10,
	}
	db := NewRedisStore()
	if err := db.Write(s); err != nil {
		t.Fatalf("err", err)
	}
}
