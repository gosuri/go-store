go-store
========

A simple, and fast [Redis](http://redis.io) backed key-value store library for [Go](http://golang.org).

**NOTE**: This library is currently under **active development** and not ready for production use.

Example
-------

The below example stores, lists and fetches the saved records

```go
package main

import (
  "fmt"

  "github.com/gosuri/go-store"
)

// Hacker implements store.Item interface methods Key and SetKey
type Hacker struct {
  Id        string
  Name      string
  Birthyear int
}

func (h *Hacker) Key() string {
  return h.Id
}

func (h *Hacker) SetKey(k string) {
  h.Id = k
}

func main() {
  db := store.NewRedisStore()

  // Save a hacker in the store with a auto-generated uuid
  db.Write(&Hacker{Name: "Alan Turing", Birthyear: 1912})

  var hackers []Hacker
  // Populate hackers slice with ids of all hackers in the store
  db.List(&hackers)

  alan := hackers[0]
  db.Read(&alan)
  fmt.Println("Hello,", alan.Name)

  fmt.Println("Listing all", len(hackers), "hackers")
  // Fetches all hackers with names from the store
  db.ReadMultiple(hackers)
  for _, hacker := range hackers {
    fmt.Printf("%s (%d) (%s)\n", hacker.Name, hacker.Birthyear, hacker.Id)
  }
}
```

Running Testing
----------------

```
$ go test
```

Benchmarks
----------

```
$ go test -bench=.
...
PASS
BenchmarkRedisWrite    10000      158512 ns/op
ok    github.com/gosuri/store 1.612s
```
