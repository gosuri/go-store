store
=====

TBD

Examples
--------

The below example store a person, lists ids and fetches the saved records

```go
// Implements store.Item interface methods Key and SetKey 
type Person struct {
	Id   string
	Name string
}

func (p *Person) Key() string {
	return p.Id
}

func (p *Person) SetKey(k string) {
	p.Id = k
}

func main() {
	db := store.NewRedisStore()
  bob := &Person{Name: "Bob"}

  // saves to redis with a generated uuid
  db.Write(bob)

  // list ids, each person struct will have the Id populated
  var people []Person
  db.List(&people)

  // Fetches all people
  db.MultiRead(people)
}
```

Testing
-------

### Running tests

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
