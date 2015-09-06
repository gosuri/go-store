package redis_test

import (
	"fmt"
	"github.com/gosuri/go-store/redis"
)

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

func Example() {
	store, err := redis.NewStore("", "")
	if err != nil {
		panic(err) // handle error
	}

	// Save a hacker in the store with a auto-generated uuid
	if err := store.Write(&Hacker{Name: "Alan Turing", Birthyear: 1912}); err != nil {
		panic(err) // handle error
	}

	var hackers []Hacker
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
