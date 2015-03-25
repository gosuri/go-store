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

  	// saves to redis with a generated uuid
	db.Write(bob)

  	// Fetches all keys
  	var people []Person
  	db.List(people)
 
  	// Fetches all people
  	db.MultiRead(people)
}
```
