// Copyright 2015 Greg Osuri. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package redis provides simple data persitance and retrieval functionality.
//
// Here is a simple example on how to write and read from redis:

// import (
// 	"fmt"
// 	"github.com/gosuri/go-store/redis"
// )
//
// type hacker struct {
// 	Id        string
// 	Name      string
// 	Birthyear int
// }
//
// func (h *hacker) Key() string {
// 	return h.Id
// }
//
// func (h *hacker) SetKey(k string) {
// 	h.Id = k
// }
//
// func main() {
// 	store, err := redis.NewStore("", "")
// 	if err != nil {
// 		panic(err) // handle error
// 	}
//
// 	// Save a hacker in the store with a auto-generated uuid
// 	if err := store.Write(&hacker{Name: "Alan Turing", Birthyear: 1912}); err != nil {
// 		panic(err) // handle error
// 	}
//
// 	var hackers []hacker
// 	// Populate hackers slice with ids of all hackers in the store
// 	store.List(&hackers)
//
// 	alan := hackers[0]
// 	store.Read(&alan)
// 	fmt.Println("Hello,", alan.Name)
//
// 	fmt.Println("Listing all", len(hackers), "hackers")
// 	// Fetches all hackers with names from the store
// 	store.ReadMultiple(hackers)
// 	for _, h := range hackers {
// 		fmt.Printf("%s (%d) (%s)\n", h.Name, h.Birthyear, h.Id)
// 	}
// }
//
package redis
