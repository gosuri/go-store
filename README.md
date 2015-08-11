store [![GoDoc](https://godoc.org/github.com/gosuri/go-store?status.svg)](https://godoc.org/github.com/gosuri/go-store) [![Build Status](https://travis-ci.org/gosuri/go-store.svg?branch=master)](https://travis-ci.org/gosuri/go-store)
=======

store is a data-store library for [Go](http://golang.org) that provides a set of platform-independent interfaces to persist and retrieve data.

Its primary goal is to wrap existing implementations of such primitives, such as those in package redis, into shared public interfaces that abstract functionality, plus some other related primitives.

It currently supports [Redis](http://redis.io) from the [redis](redis/) package.

**NOTE**: This library is currently under **active development** and not ready for production use.

Example
-------

The below example stores, lists and fetches the saved records

```go
package main

import (
  "fmt"

  "github.com/gosuri/go-store/redis"
)

type hacker struct {
  Id        string
  Name      string
  Birthyear int
}

func (h *hacker) Key() string {
  return h.Id
}

func (h *hacker) SetKey(k string) {
  h.Id = k
}

func main() {
  store, err := redis.NewStore("", "")
  if err != nil {
    panic(err) // handle error
  }

  // Save a hacker in the store with a auto-generated uuid
  if err := store.Write(&hacker{Name: "Alan Turing", Birthyear: 1912}); err != nil {
    panic(err) // handle error
  }

  var hackers []hacker
  // Populate hackers slice with ids of all hackers in the store
  store.List(&hackers)

  alan := hackers[0]
  store.Read(&alan)
  fmt.Println("Hello,", alan.Name)

  fmt.Println("Listing all", len(hackers), "hackers")
  // Fetches all hackers with names from the store
  store.ReadMultiple(hackers)
  for _, h := range hackers {
    fmt.Printf("%s (%d) (%s)\n", h.Name, h.Birthyear, h.Id)
  }
}
```

Roadmap
-------

Below are items I am and plan on working on in the near future. Contributions are welcome.

Feature | Status
--- | ---
Save multiple records in a single call using pipelining | **implementing**
Query (using finders) and indexing |

Contributing
------------

### Dependency management

Users who import `store` into their package main are responsible to organize and maintain all of their dependencies to ensure code compatibility and build reproducibility.
Store makes no direct use of dependency management tools like [Godep](https://github.com/tools/godep).

We will use a variety of continuous integration providers to find and fix compatibility problems as soon as they occur.

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
BenchmarkRedisWrite     10000   178342 ns/op
BenchmarkRead           10000   119449 ns/op
BenchmarkRead1k         10   120388644 ns/op
BenchmarkRedisList1k    50    33211769 ns/op
BenchmarkRedisList10k   20    79867558 ns/op
BenchmarkReadMultiple1k 200   10372213 ns/op
```
