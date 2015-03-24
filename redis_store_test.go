package store

import (
	"reflect"
	"testing"

	"code.google.com/p/go-uuid/uuid"
	"github.com/garyburd/redigo/redis"
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
		Id:         uuid.New(),
		Field:      "value",
		FieldInt:   10,
		FieldFloat: 1.234,
		FieldBool:  true,
		FieldUint:  1,
	}

	db := NewRedisStore()

	if err := db.Write(s); err != nil {
		t.Fatalf("err", err)
	}

	if len(s.Key()) == 0 {
		t.Fatalf("key is emtpy %#v", s)
	}

	pool := NewRedisPool(NewRedisConfig())
	c := pool.Get()
	defer c.Close()
	reply, err := redis.Values(c.Do("HGETALL", "RedisTestStruct:"+s.Key()))
	if err != nil {
		t.Fatalf("err", err)
	}

	got := &RedisTestStruct{}

	if err := redis.ScanStruct(reply, got); err != nil {
		t.Fatalf("err", err)
	}

	if !reflect.DeepEqual(s, got) {
		t.Fatalf("expected:", s, " got:", got)
	}
}

func TestRead(t *testing.T) {
	s := &RedisTestStruct{
		Id:    uuid.New(),
		Field: "value",
	}
	db := NewRedisStore()

	if err := db.Write(s); err != nil {
		t.Fatalf("err", err)
	}
	got := &RedisTestStruct{Id: s.Key()}
	if err := db.Read(got); err != nil {
		t.Fatalf("err", err)
	}
	if !reflect.DeepEqual(s, got) {
		t.Fatalf("expected:", s, " got:", got)
	}
}

func TestReadNotFound(t *testing.T) {
	db := NewRedisStore()
	got := &RedisTestStruct{Id: "invalid"}
	if err := db.Read(got); err != ErrKeyNotFound {
		t.Fatalf("expected ErrNotFound, got: ", err)
	}
}

func TestList(t *testing.T) {
	db := NewRedisStore()
	i1 := RedisTestStruct{Field: "field1"}
	i2 := RedisTestStruct{Field: "field2"}
	items := []RedisTestStruct{i1, i2}
	for _, item := range items {
		db.Write(&item)
	}
	got := []RedisTestStruct{}
	if err := db.List(got); err != nil {
		t.Fatalf("err", err)
	}
}
