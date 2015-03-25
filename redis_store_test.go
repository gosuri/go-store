package store

import (
	"reflect"
	"testing"

	"code.google.com/p/go-uuid/uuid"
	"github.com/garyburd/redigo/redis"
)

type TestR struct {
	Id         string
	Field      string
	FieldFloat float32
	FieldInt   int
	FieldBool  bool
	FieldUint  uint
}

type TestRs []TestR

func (s *TestR) Key() string {
	return s.Id
}

func (s *TestR) SetKey(k string) {
	s.Id = k
}

func TestWrite(t *testing.T) {
	s := &TestR{
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
	reply, err := redis.Values(c.Do("HGETALL", "TestR:"+s.Key()))
	if err != nil {
		t.Fatalf("err", err)
	}

	got := &TestR{}

	if err := redis.ScanStruct(reply, got); err != nil {
		t.Fatalf("err", err)
	}

	if !reflect.DeepEqual(s, got) {
		t.Fatalf("expected:", s, " got:", got)
	}
}

func TestRead(t *testing.T) {
	s := &TestR{
		Id:    uuid.New(),
		Field: "value",
	}
	db := NewRedisStore()

	if err := db.Write(s); err != nil {
		t.Fatalf("err", err)
	}
	got := &TestR{Id: s.Key()}
	if err := db.Read(got); err != nil {
		t.Fatalf("err", err)
	}
	if !reflect.DeepEqual(s, got) {
		t.Fatalf("expected:", s, " got:", got)
	}
}

func TestReadNotFound(t *testing.T) {
	db := NewRedisStore()
	got := &TestR{Id: "invalid"}
	if err := db.Read(got); err != ErrKeyNotFound {
		t.Fatalf("expected ErrNotFound, got: ", err)
	}
}

func TestList(t *testing.T) {
	db := NewRedisStore()
	i1 := TestR{Field: "field1"}
	db.Write(&i1)
	i2 := TestR{Field: "field2"}
	db.Write(&i2)

	got := []TestR{}
	if err := db.List(&got); err != nil {
		t.Fatalf("err", err)
	}

	items := []TestR{i1, i2}
	if len(got) < 1 {
		t.Fatalf("expected %d, got: %d", len(items), len(got))
	}
}

func TestReadMultpile(t *testing.T) {
	db := NewRedisStore()
	i := TestR{Field: "field1"}
	db.Write(&i)
	items := TestRs{i}

	got := TestRs{{Id: i.Key()}}
	if err := db.ReadMultiple(got); err != nil {
		t.Fatalf("err: %v", err)
	}

	if !reflect.DeepEqual(got, items) {
		t.Fatalf("expected %#v, got %#v", items, got)
	}
}
