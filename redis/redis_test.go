// Copyright 2015 Greg Osuri. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.

package redis

import (
	"reflect"
	"testing"

	"github.com/gosuri/go-store/_vendor/src/code.google.com/p/go-uuid/uuid"
	driver "github.com/gosuri/go-store/_vendor/src/github.com/garyburd/redigo/redis"
	"github.com/gosuri/go-store/store"
)

type TestR struct {
	ID         string
	Field      string
	FieldFloat float32
	FieldInt   int
	FieldBool  bool
	FieldUint  uint
}

type TestRs []TestR

func (s *TestR) Key() string {
	return s.ID
}

func (s *TestR) SetKey(k string) {
	s.ID = k
}

func TestWrite(t *testing.T) {
	s := &TestR{
		ID:         uuid.New(),
		Field:      "value",
		FieldInt:   10,
		FieldFloat: 1.234,
		FieldBool:  true,
		FieldUint:  1,
	}

	db := NewStore()

	if err := db.Write(s); err != nil {
		t.Fatal("err", err)
	}

	if len(s.Key()) == 0 {
		t.Fatalf("key is emtpy %#v", s)
	}

	pool := NewPool(NewConfig())
	c := pool.Get()
	defer c.Close()
	reply, err := driver.Values(c.Do("HGETALL", "TestR:"+s.Key()))
	if err != nil {
		t.Fatal("err", err)
	}

	got := &TestR{}

	if err := driver.ScanStruct(reply, got); err != nil {
		t.Fatal("err", err)
	}

	if !reflect.DeepEqual(s, got) {
		t.Fatal("expected:", s, " got:", got)
	}
}

func BenchmarkRedisWrite(b *testing.B) {
	db := NewStore()
	for i := 0; i < b.N; i++ {
		db.Write(&TestR{Field: "BenchmarkWrite"})
	}
}

func TestRead(t *testing.T) {
	s := &TestR{
		ID:    uuid.New(),
		Field: "value",
	}
	db := NewStore()

	if err := db.Write(s); err != nil {
		t.Fatal("err", err)
	}
	got := &TestR{ID: s.Key()}
	if err := db.Read(got); err != nil {
		t.Fatal("err", err)
	}
	if !reflect.DeepEqual(s, got) {
		t.Fatal("expected:", s, " got:", got)
	}
}

func TestReadNotFound(t *testing.T) {
	db := NewStore()
	got := &TestR{ID: "invalid"}
	if err := db.Read(got); err != store.ErrKeyNotFound {
		t.Fatal("expected ErrNotFound, got: ", err)
	}
}

func benchmarkRead(n int, b *testing.B) {
	db := NewStore()
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
	db := NewStore()
	noItems := 1001

	for i := 0; i < noItems; i++ {
		db.Write(&TestR{Field: "..."})
	}

	var got []TestR
	if err := db.List(&got); err != nil {
		t.Fatal("err", err)
	}

	if len(got) != noItems {
		t.Fatalf("expected length to be %d, got: %d", noItems, len(got))
	}

	for _, item := range got {
		if len(item.ID) == 0 {
			t.Fatal("expected id to be present")
		}
	}
}

func benchmarkList(n int, b *testing.B) {
	db := NewStore()
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
	db := NewStore()
	i := TestR{Field: "field1"}
	db.Write(&i)
	i2 := TestR{Field: "field1"}
	db.Write(&i2)
	items := []TestR{i, i2}

	got := []TestR{{ID: i.Key()}, {ID: i2.Key()}}
	if err := db.ReadMultiple(got); err != nil {
		t.Fatalf("err: %v", err)
	}

	if !reflect.DeepEqual(got, items) {
		t.Fatalf("Mismatch\nexp: %#v \ngot: %#v", items, got)
	}
}

func benchmarkReadMultiple(n int, b *testing.B) {
	db := NewStore()
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
	pool := NewPool(NewConfig())
	c := pool.Get()
	defer c.Close()
	if _, err := c.Do("FLUSHDB"); err != nil {
		panic(err)
	}
}
