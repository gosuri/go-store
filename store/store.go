// Copyright 2015 Greg Osuri. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package store provides platform-independent interfaces to persist
// and retrieve data.
//
// Its primary job is to wrap existing implementations of such primitives,
// such as those in package redis, into shared public interfaces that
// abstract functionality, plus some other related primitives.
//
// Because these interfaces and primitives wrap lower-level operations with
// various implementations, unless otherwise informed clients should not
// assume they are safe for parallel execution.
package store

import (
	"errors"
)

// ErrKeyNotFound means that the object associated with the
// key is not found in the datastore
var ErrKeyNotFound = errors.New("store: key not found")

// ErrEmptyKey means that the key for the object provided is empty
var ErrEmptyKey = errors.New("store: key is empty")

// Store is the interface to store implemented in package redis. It groups
// ReadWriter, Lister and MultiReader interfaces.
type Store interface {
	ReadWriter
	Lister
	MultiReadWriter
}

// Item is the interface that wraps Key and SetKey methods.
//
// Key returns a unique idenfier for the item from the implementation. It is called by
// the implementing store when retrieving objects from its underlying store.
//
// SetKey sets the item key. It is called by the store when Unmarshalling objects from
// its underlying store.
//
// The below example illustrates usage:
//
//	type Hacker struct {
//		Id        string
//		Name      string
//		Birthyear int
//	}
//
//	func (h *Hacker) Key() string {
//		return h.Id
//	}
//
//	func (h *Hacker) SetKey(k string) {
//		h.Id = k
//	}
type Item interface {
	Key() string
	SetKey(string)
}

// Writer is the interface that wraps the basic Write method.
//
// Write writes i to the underlying data store. It expects i.Key method
// to return a unique identifier for the item, it will otherwise generate
// a unique identifier and calls i.SetKey method. It returns any error
// encountered that caused the write to stop early.
type Writer interface {
	Write(i Item) error
}

// Reader is the interface that wraps the basic Read method.
//
// Read reads i from the underlying data store and copies to i. It returns
// any error encountered that caused the write to stop early.
type Reader interface {
	Read(i Item) error
}

// Deleter is the interface that wraps the basic Delete method.
//
// Delete deletes i from the underlying datastore. It returns any error
// encountered that prevented the deletion from occurring.
type Deleter interface {
	Delete(i Item) error
}

// ReadWriter is the interface that groups Reader, Writer and Deleter
// interfaces.
type ReadWriter interface {
	Reader
	Writer
	Deleter
}

// Lister is the interface that wraps the basic List method.
type Lister interface {
	List(interface{}) error
}

// MultiReader is the interface that wraps ReadMultiple method.
type MultiReader interface {
	ReadMultiple(interface{}) error
}

// MultiWriter is the interface that wraps WriteMultiple method.
type MultiWriter interface {
	WriteMultiple(items []Item) error
}

// MultiDeleter is the interface that wraps DeleteMultiple method.
type MultiDeleter interface {
	DeleteMultiple(items []Item) (int, error)
}

// MultiReadWriter is the interface that groups MultiReader and
// MultiWriter interfaces.
type MultiReadWriter interface {
	MultiReader
	MultiWriter
	MultiDeleter
}
