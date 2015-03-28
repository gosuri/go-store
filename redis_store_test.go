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

func BenchmarkRedisWrite(b *testing.B) {
	db := NewRedisStore()
	for i := 0; i < b.N; i++ {
		db.Write(&TestR{Field: "BenchmarkWrite"})
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

func benchmarkRead(n int, b *testing.B) {
	db := NewRedisStore()
	items := make([]TestR, n, n)
	for i := 0; i < n; i++ {
		item := TestR{Field: "..."}
		db.Write(&item)
		items[i] = item
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, item := range items {
			db.Read(&item)
		}
	}
}

func BenchmarkRead(b *testing.B) { benchmarkRead(1, b) }

func BenchmarkRead1k(b *testing.B) { benchmarkRead(1000, b) }

func TestList(t *testing.T) {
	flushRedisDB()
	db := NewRedisStore()
	noItems := 1001

	for i := 0; i < noItems; i++ {
		db.Write(&TestR{Field: "..."})
	}

	var got []TestR
	if err := db.List(&got); err != nil {
		t.Fatalf("err", err)
	}

	if len(got) != noItems {
		t.Fatalf("expected length to be %d, got: %d", noItems, len(got))
	}

	for _, item := range got {
		if len(item.Id) == 0 {
			t.Fatalf("expected id to be present")
		}
	}
}

func benchmarkList(n int, b *testing.B) {
	db := NewRedisStore()
	for i := 0; i < n; i++ {
		db.Write(&TestR{Field: "..."})
	}
	var items []TestR
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.List(&items)
	}
}

func BenchmarkRedisList1k(b *testing.B)  { benchmarkList(1000, b) }
func BenchmarkRedisList10k(b *testing.B) { benchmarkList(10000, b) }

func TestReadMultpile(t *testing.T) {
	db := NewRedisStore()
	i := TestR{Field: "field1"}
	db.Write(&i)
	i2 := TestR{Field: "field1"}
	db.Write(&i2)
	items := []TestR{i, i2}

	got := []TestR{{Id: i.Key()}, {Id: i2.Key()}}
	if err := db.ReadMultiple(got); err != nil {
		t.Fatalf("err: %v", err)
	}

	if !reflect.DeepEqual(got, items) {
		t.Fatalf("Mismatch\nexp: %#v \ngot: %#v", items, got)
	}
}

func benchmarkReadMultiple(n int, b *testing.B) {
	db := NewRedisStore()
	items := make([]TestR, n, n)
	for i := 0; i < n; i++ {
		item := TestR{Field: "..."}
		db.Write(&item)
		items[i] = item
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.ReadMultiple(items)
	}
}

func BenchmarkReadMultiple1k(b *testing.B) { benchmarkReadMultiple(1000, b) }

func flushRedisDB() {
	pool := NewRedisPool(NewRedisConfig())
	c := pool.Get()
	defer c.Close()
	if _, err := c.Do("FLUSHDB"); err != nil {
		panic(err)
	}
}
