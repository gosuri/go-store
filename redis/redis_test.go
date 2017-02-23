// Copyright 2015 Greg Osuri. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.

package redis

import (
	"reflect"
	"testing"

	driver "github.com/garyburd/redigo/redis"
	"github.com/google/uuid"
	"github.com/gosuri/go-store/store"
)

var (
	testNs       = uuid.New().String()
	testRedisURL = "redis://@127.0.0.1:6379/2"
)

type TestR struct {
	ID         string
	Field      string
	FieldFloat float32
	FieldInt   int
	FieldBool  bool
	FieldUint  uint

	fieldPrivate string
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
		ID:         uuid.New().String(),
		Field:      "value",
		FieldInt:   10,
		FieldFloat: 1.234,
		FieldBool:  true,
		FieldUint:  1,
	}

	db := testStore(t)

	if err := db.Write(s); err != nil {
		t.Fatal("err", err)
	}

	if len(s.Key()) == 0 {
		t.Fatalf("key is emtpy %#v", s)
	}

	cfg, err := NewConfig(testRedisURL)
	if err != nil {
		t.Fatal(err)
	}

	pool := NewPool(cfg)
	c := pool.Get()
	defer c.Close()
	reply, err := driver.Values(c.Do("HGETALL", testNs+":TestR:"+s.Key()))
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
	db := testStoreB(b)
	for i := 0; i < b.N; i++ {
		db.Write(&TestR{Field: "BenchmarkWrite"})
	}
}

func TestRead(t *testing.T) {
	s := &TestR{
		ID:    uuid.New().String(),
		Field: "value",
	}
	db := testStore(t)

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
	db := testStore(t)
	got := &TestR{ID: "invalid"}
	if err := db.Read(got); err != store.ErrKeyNotFound {
		t.Fatal("expected ErrNotFound, got: ", err)
	}
}

func TestDelete(t *testing.T) {
	s := &TestR{
		ID:    uuid.New().String(),
		Field: "value",
	}
	db := testStore(t)

	if err := db.Write(s); err != nil {
		t.Fatal("err", err)
	}
	got := &TestR{ID: s.Key()}
	if err := db.Delete(got); err != nil {
		t.Fatal("err", err)
	}

	if err := db.Read(got); err != store.ErrKeyNotFound {
		t.Fatal("expected ErrNotFound, got: ", err)
	}
}

func TestDeleteMultiple(t *testing.T) {
	s := &TestR{
		ID:    uuid.New().String(),
		Field: "value",
	}

	s1 := &TestR{
		ID:    uuid.New().String(),
		Field: "value1",
	}

	db := testStore(t)

	if err := db.Write(s); err != nil {
		t.Fatal("err", err)
	}

	if err := db.Write(s1); err != nil {
		t.Fatal("err", err)
	}

	toDel := []store.Item{s, s1}
	count, err := db.DeleteMultiple(toDel)
	if err != nil {
		t.Fatal("err", err)
	}
	if count != 2 {
		t.Fatal("expected 2 deletions, got: ", count)
	}
}

func TestPartialDeleteMultiple(t *testing.T) {
	db := testStore(t)
	s := &TestR{
		ID:    uuid.New().String(),
		Field: "value",
	}

	s1 := &TestR{
		ID:    uuid.New().String(),
		Field: "value1",
	}

	s2 := &TestR{
		ID:    uuid.New().String(),
		Field: "value2",
	}

	if err := db.Write(s); err != nil {
		t.Fatal("err", err)
	}

	if err := db.Write(s1); err != nil {
		t.Fatal("err", err)
	}

	toDel := []store.Item{s, s1, s2}
	count, err := db.DeleteMultiple(toDel)
	if err != store.ErrKeyNotFound {
		t.Fatal("expected ErrKeyNotFound, got: ", err)
	}

	if count != 2 {
		t.Fatal("expected 2 deletions, got: ", count)
	}
}

func TestDeleteNotFound(t *testing.T) {
	db := testStore(t)
	got := &TestR{ID: "invalid"}
	if err := db.Delete(got); err != store.ErrKeyNotFound {
		t.Fatal("expected ErrKeyNotFound, got: ", err)
	}
}

func TestDeleteNoKey(t *testing.T) {
	db := testStore(t)
	got := &TestR{}
	if err := db.Delete(got); err != store.ErrEmptyKey {
		t.Fatal("expected ErrEmptyKey, got: ", err)
	}
}

func benchmarkRead(n int, b *testing.B) {
	db := testStoreB(b)
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
	db := testStore(t)
	flushRedisDB()

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
	db := testStoreB(b)
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

func TestReadMultiple(t *testing.T) {
	db := testStore(t)
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
	db := testStoreB(b)
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

func testStoreB(b *testing.B) store.Store {
	db, err := NewStore(testRedisURL, testNs)
	if err != nil {
		b.Fatal(err)
	}
	return db
}

func testStore(t *testing.T) store.Store {
	db, err := NewStore(testRedisURL, testNs)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func flushRedisDB() {
	cfg, err := NewConfig(testRedisURL)
	if err != nil {
		panic(err)
	}

	pool := NewPool(cfg)
	c := pool.Get()
	defer c.Close()
	if _, err := c.Do("FLUSHDB"); err != nil {
		panic(err)
	}
}
