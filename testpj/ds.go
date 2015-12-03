package main

import (
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/datastore"
	"io/ioutil"
	"log"
	"reflect"
	"time"
)

func main() {

	//put()
	//get()
	//delete()
	//getKeys()
	//curser()
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

	ctx = datastore.WithNamespace(ctx, "huehuehue")
	key := datastore.NewKey(ctx, "asdf", "708", 0, nil)

	var dk map[string]interface{}
	dk = make(map[string]interface{})
	dk["Name"] = "ayyo"
	dk["Age"] = 789

	var props datastore.PropertyList

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

func getKeys() {
	props := make([]datastore.PropertyList, 0)
	ctx := context.Background()
	client := Example_auth()
	ctx = datastore.WithNamespace(ctx, "huehuehue")
	q := datastore.NewQuery("asdf")
	_, err := client.GetAll(ctx, q, &props)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(props)
	}
}

func curser() {
	ctx := context.Background()
	client := Example_auth()
	ctx = datastore.WithNamespace(ctx, "huehuehue")
	q := datastore.NewQuery("asdf")

	t := client.Run(ctx, q)
	for {
		fmt.Println("--------------------")
		cur, _ := t.Cursor()
		fmt.Println(cur.String())
		var props datastore.PropertyList
		_, err := t.Next(&props)
		if err == datastore.Done {
			fmt.Println("Done")
			break
		}
		if err != nil {
			fmt.Println("Error Fetching Next : " + err.Error())
			break
		}
		// Do something with the Person p
		fmt.Println(props)
	}
}

func get() {
	ctx := context.Background()
	client := Example_auth()
	//ctx = datastore.WithNamespace(ctx, "Default")
	key := datastore.NewKey(ctx, "asdf", "702", 0, nil)

	var props datastore.PropertyList

	if err := client.Get(ctx, key, &props); err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(props)
		for _, v := range props {
			fmt.Println(v.Name)     //string
			fmt.Println(v.Value)    //interface{}
			fmt.Println(v.NoIndex)  //bool
			fmt.Println(v.Multiple) //bool
			fmt.Println("----------------------")
		}
	}
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

/*
func setManyDataStore(objects []map[string]interface{}) {
		ctx := context.Background()
		client, err := getDataStoreClient() //have connection code in another function
		ctx = datastore.WithNamespace(ctx, "CompanyA")

		if err == nil {

			var keys []*datastore.Key
			keys = make([]*datastore.Key, len(Objects))
			var propArray []datastore.PropertyList
			propArray = make([]datastore.PropertyList, len(Objects))

			for index := 0; index < len(Objects); index++ {
				keys[index] = datastore.NewKey(ctx, "users", "", index, nil)

				var props datastore.PropertyList

				for key, value := range Objects[index] {
					props = append(props, datastore.Property{Name: key, Value: value})
				}
				propArray[index] = props
			}

			if _, err := client.PutMulti(ctx, keys, &propArray); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Success!")
			}

		} else {
			fmt.Println("Connection Failed")
		}
	}
*/
