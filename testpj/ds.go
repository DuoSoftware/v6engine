package main

import (
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/datastore"
	"io/ioutil"
	"log"
	"time"
)

func main() {

	put()
	//get()
	//delete()
}

func Example_auth() *datastore.Client {
	jsonKey, err := ioutil.ReadFile("DUOWORLD-60d1c8c347de.json")
	if err != nil {
		log.Fatal(err)
	}
	conf, err := google.JWTConfigFromJSON(
		jsonKey,
		datastore.ScopeDatastore,
		datastore.ScopeUserEmail,
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "duo-world", cloud.WithTokenSource(conf.TokenSource(ctx)))
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func put() {
	ctx := context.Background()
	client := Example_auth()

	key := datastore.NewKey(ctx, "huehuehue", "701", 0, nil)

	// _, err := client.Put(ctx, key, &Book{
	// 	Title:       "111",
	// 	Description: "22222",
	// 	Body:        "33333",
	// 	Author:      "4444",
	// 	PublishedAt: time.Now(),
	// })

	var dk map[string]interface{}
	dk = make(map[string]interface{})
	dk["1"] = "hehe"
	dk["2"] = 123

	// asdf := datastore.PropertyList{
	// 	datastore.Property{Name: "time", Value: time.Now()},
	// 	datastore.Property{Name: "email", Value: "me@myhost.com"},
	// 	datastore.Property{Name: "1", Value: 1231452},
	// 	datastore.Property{Name: "2", Value: "fdsa"},
	// 	datastore.Property{Name: "3", Value: "huehuehue"},
	// }

	var props datastore.PropertyList
	//props = append(props, datastore.Property{Name: "time", Value: time.Now()})
	//props = append(props, datastore.Property{Name: "email", Value: "me@myhost.com"})

	for key, value := range dk {
		props = append(props, datastore.Property{Name: key, Value: value})
	}

	_, err := client.Put(ctx, key, &props)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(key)
	}

}

type DynEnt map[string]interface{}

func (d *DynEnt) Load(ch <-chan datastore.Property) error {
	// Note: you might want to clear current values from the map or create a new map
	for p := range ch { // Read until channel is closed
		(*d)[p.Name] = p.Value
	}
	return nil
}

func (d *DynEnt) Save(ch chan<- datastore.Property) error {
	for k, v := range *d {
		ch <- datastore.Property{Name: k, Value: v}
	}
	close(ch) // Channel must be closed
	return nil
}

func get() {
	ctx := context.Background()
	client := Example_auth()

	key := datastore.NewKey(ctx, "Book", "700", 0, nil)
	fmt.Println(key.Namespace())
	book := &Book{}
	if err := client.Get(ctx, key, book); err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(key.Namespace())
		fmt.Println(book)
	}
}

type Book struct {
	Title       string
	Description string
	Body        string `datastore:",noindex"`
	Author      string
	PublishedAt time.Time
}

type Record struct {
	_os_id string
	object map[string]interface{}
}

func delete() {
	ctx := context.Background()
	client := Example_auth()
	key := datastore.NewKey(ctx, "Book", "700", 0, nil)

	if err := client.Delete(ctx, key); err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Deleted!")
	}
}

type Post struct {
	Title       string
	PublishedAt time.Time
	Comments    int
}

func ExampleGetMulti() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "duo-world")
	if err != nil {
		log.Fatal(err)
	}

	keys := []*datastore.Key{
		datastore.NewKey(ctx, "Post", "post1", 0, nil),
		datastore.NewKey(ctx, "Post", "post2", 0, nil),
		datastore.NewKey(ctx, "Post", "post3", 0, nil),
	}
	posts := make([]Post, 3)
	if err := client.GetMulti(ctx, keys, posts); err != nil {
		log.Println(err)
	}
}

func ExamplePutMulti_slice() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "duo-world")
	if err != nil {
		log.Fatal(err)
	}

	keys := []*datastore.Key{
		datastore.NewKey(ctx, "Post", "post1", 0, nil),
		datastore.NewKey(ctx, "Post", "post2", 0, nil),
	}

	// PutMulti with a Post slice.
	posts := []*Post{
		{Title: "Post 1", PublishedAt: time.Now()},
		{Title: "Post 2", PublishedAt: time.Now()},
	}
	if _, err := client.PutMulti(ctx, keys, posts); err != nil {
		log.Fatal(err)
	}
}

func ExamplePutMulti_interfaceSlice() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "duo-world")
	if err != nil {
		log.Fatal(err)
	}

	keys := []*datastore.Key{
		datastore.NewKey(ctx, "Post", "post1", 0, nil),
		datastore.NewKey(ctx, "Post", "post2", 0, nil),
	}

	// PutMulti with an empty interface slice.
	posts := []interface{}{
		&Post{Title: "Post 1", PublishedAt: time.Now()},
		&Post{Title: "Post 2", PublishedAt: time.Now()},
	}
	if _, err := client.PutMulti(ctx, keys, posts); err != nil {
		log.Fatal(err)
	}
}

func ExampleQuery() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "duo-world")
	if err != nil {
		log.Fatal(err)
	}

	// Count the number of the post entities.
	q := datastore.NewQuery("Post")
	n, err := client.Count(ctx, q)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("There are %d posts.", n)

	// List the posts published since yesterday.
	yesterday := time.Now().Add(-24 * time.Hour)
	q = datastore.NewQuery("Post").Filter("PublishedAt >", yesterday)
	it := client.Run(ctx, q)
	// Use the iterator.
	_ = it

	// Order the posts by the number of comments they have recieved.
	datastore.NewQuery("Post").Order("-Comments")

	// Start listing from an offset and limit the results.
	datastore.NewQuery("Post").Offset(20).Limit(10)
}

func ExampleTransaction() {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, "duo-world")
	if err != nil {
		log.Fatal(err)
	}
	const retries = 3

	// Increment a counter.
	// See https://cloud.google.com/appengine/articles/sharding_counters for
	// a more scalable solution.
	type Counter struct {
		Count int
	}

	key := datastore.NewKey(ctx, "counter", "CounterA", 0, nil)

	for i := 0; i < retries; i++ {
		tx, err := client.NewTransaction(ctx)
		if err != nil {
			break
		}

		var c Counter
		if err := tx.Get(key, &c); err != nil && err != datastore.ErrNoSuchEntity {
			break
		}
		c.Count++
		if _, err := tx.Put(key, &c); err != nil {
			break
		}

		// Attempt to commit the transaction. If there's a conflict, try again.
		if _, err := tx.Commit(); err != datastore.ErrConcurrentTransaction {
			break
		}
	}

}
