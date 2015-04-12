// Copyright 2015 Greg Osuri. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
//
// Package store provides basic interfaces to datastore primitives.
// Its primary job is to wrap existing implementations of such primitives,
// such as those in package redis, into shared public interfaces that
// abstract functionality, plus some other related primitives.
//
// Because these interfaces and primitives wrap lower-level operations with
// various implementations, unless otherwise informed clients should not
// assume they are safe for parallel execution.
package store // import "github.com/gosuri/go-store/store"

import (
	"errors"
)

// ErrKeyNotFound means that the object associated with the
// key is not found in the datastore
var ErrKeyNotFound = errors.New("store: key not found")

// ErrEmptyKey means that the key for the object provided is empty
var ErrEmptyKey = errors.New("store: key is empty")

// Store is the interface to stores implemented in package runtime.
//
// Store provides persistence to objects
type Store interface {
	ReadWriter
	Lister
	MultiReadWriter
}

// Item is the interface that wraps Key and SetKey methods
type Item interface {
	Key() string
	SetKey(string)
}

// Writer is the interface that wraps the basic Write method.
//
// Write writes len(i) bytes from i to the underlying data store. It returns
// any error encountered that caused the write to stop early or not start at all.
type Writer interface {
	Write(i Item) error
}

type Reader interface {
	Read(Item) error
}

type ReadWriter interface {
	Reader
	Writer
}

type Lister interface {
	List(interface{}) error
}

type MultiReader interface {
	ReadMultiple(interface{}) error
}

type MultiWriter interface {
	WriteMultiple(items []Item) error
}

type MultiReadWriter interface {
	MultiReader
	MultiWriter
}

type Items interface {
	Item
}
