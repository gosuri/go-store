store
=====

TBD

Example

```go
type Person struct {
	Id   string
	Name string
}

func (p *Person) Key() string {
	return p.Id
}

func main() {
  db := store.NewRedisStore()
  bob := &Person{Name: "Bob"}

  // saves to redis with a generate id
  db.Write(bob)

  // list ids, each person struct will have the Id populated
  var people []Person
  db.List(people)
 
  // Fetches all and stores in people
  db.MultiRead(people)
}
```
