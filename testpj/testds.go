package main

import (
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/datastore"
	"io/ioutil"
)

func main() {

	// var allMaps []map[string]interface{}
	// allMaps = make([]map[string]interface{}, 2)

	// var map1 map[string]interface{}
	// map1 = make(map[string]interface{})
	// map1["Id"] = "700"
	// map1["Name"] = "Jay"
	// map1["Age"] = "23"

	// var map2 map[string]interface{}
	// map2 = make(map[string]interface{})
	// map2["Id"] = "800"
	// map2["Name"] = "Peter"
	// map2["Age"] = "30"

	// allMaps[0] = map1
	// allMaps[1] = map2

	// setManyDataStore(allMaps)
	//ExampleTransaction(allMaps)
	getMultiDataStore()

	//ExamplePutMulti_interfaceSlice()
}

func getDataStoreClient() (client *datastore.Client, err error) {

	keyFile := "DUOWORLD-60d1c8c347de.json"
	projectID := "duo-world"

	jsonKey, err := ioutil.ReadFile(keyFile)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		conf, err := google.JWTConfigFromJSON(
			jsonKey,
			datastore.ScopeDatastore,
			datastore.ScopeUserEmail,
		)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			ctx := context.Background()
			client, err = datastore.NewClient(ctx, projectID, cloud.WithTokenSource(conf.TokenSource(ctx)))
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}

	return
}

func setManyDataStore(Objects []map[string]interface{}) {
	ctx := context.Background()
	client, err := getDataStoreClient() //have connection code in another function
	ctx = datastore.WithNamespace(ctx, "CompanyA")

	if err == nil {

		var keys []*datastore.Key
		keys = make([]*datastore.Key, len(Objects))

		// var propArray []*datastore.PropertyList
		// propArray = make([]datastore.PropertyList, len(Objects))

		// for i, _ := range propArray {
		// 	propArray[i] = datastore.PropertyList{}
		// }

		propArray := make([]interface{}, len(Objects))

		for index := 0; index < len(Objects); index++ {
			keys[index] = datastore.NewKey(ctx, "users", Objects[index]["Id"].(string), 0, nil)

			props := datastore.PropertyList{}

			for key, value := range Objects[index] {
				props = append(props, datastore.Property{Name: key, Value: value})
			}
			propArray[index] = &props
		}

		if _, err := client.PutMulti(ctx, keys, propArray); err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("Success!")
		}

	} else {
		fmt.Println("Connection Failed")
	}
}

func getMultiDataStore() {
	ctx := context.Background()
	client, err := getDataStoreClient() //have connection code in another function
	ctx = datastore.WithNamespace(ctx, "CompanyA")

	if err == nil {

		var keys []*datastore.Key
		keys = make([]*datastore.Key, 1)
		//var propArray []datastore.PropertyList
		//propArray = make([]datastore.PropertyList, 1)

		keys[0] = datastore.NewKey(ctx, "users", "600", 0, nil)

		if err := client.GetMulti(ctx, keys, nil); err == nil {
			//fmt.Println(propArray)
			fmt.Println(keys)
		} else {
			fmt.Println(err.Error())
		}

	} else {
		fmt.Println("Connection Failed")
	}
}

func PutMultiTransaction(Objects []map[string]interface{}) {
	ctx := context.Background()
	client, _ := getDataStoreClient() //have connection code in another function
	ctx = datastore.WithNamespace(ctx, "CompanyA")

	const retries = 3

	var keys []*datastore.Key
	keys = make([]*datastore.Key, len(Objects))

	var propArray []datastore.PropertyList
	propArray = make([]datastore.PropertyList, len(Objects))

	for idx, _ := range propArray {
		propArray[idx] = datastore.PropertyList{}
	}

	for index := 0; index < len(Objects); index++ {
		keys[index] = datastore.NewKey(ctx, "users", Objects[index]["Id"].(string), 0, nil)

		props := datastore.PropertyList{}

		for key, value := range Objects[index] {
			props = append(props, datastore.Property{Name: key, Value: value})
		}
		propArray[index] = props
	}

	for i := 0; i < retries; i++ {
		tx, err := client.NewTransaction(ctx)
		if err != nil {
			fmt.Println(err.Error())
			break
		}

		if _, err := tx.PutMulti(keys, propArray); err != nil {
			fmt.Println(err.Error())
			break
		}

		if _, err := tx.Commit(); err != datastore.ErrConcurrentTransaction {
			break
		}
	}
}

func ExamplePutMulti_interfaceSlice() {
	ctx := context.Background()
	client, _ := getDataStoreClient() //have connection code in another function
	ctx = datastore.WithNamespace(ctx, "anemanda")

	keys := []*datastore.Key{
		datastore.NewKey(ctx, "Post", "post1", 0, nil),
		datastore.NewKey(ctx, "Post", "post2", 0, nil),
	}

	// PutMulti with an empty interface slice.

	// posts := []interface{}{
	// 	&datastore.PropertyList{datastore.Property{Name: "1", Value: "asdf"}, datastore.Property{Name: "11", Value: "1111111111"}},
	// 	&datastore.PropertyList{datastore.Property{Name: "2", Value: "zxcv"}, datastore.Property{Name: "22", Value: "2222222222"}},
	// }

	// posts := []interface{}{
	// 	&datastore.PropertyList{datastore.Property{Name: "1", Value: "asdf"}, datastore.Property{Name: "11", Value: "1111111111"}},
	// 	&datastore.PropertyList{datastore.Property{Name: "2", Value: "zxcv"}, datastore.Property{Name: "22", Value: "2222222222"}},
	// }

	posts2 := make([]interface{}, len(keys))
	posts2[0] = &datastore.PropertyList{datastore.Property{Name: "1", Value: "asdf"}, datastore.Property{Name: "11", Value: "1111111111"}}
	posts2[1] = &datastore.PropertyList{datastore.Property{Name: "2", Value: "zxcv"}, datastore.Property{Name: "22", Value: "2222222222"}}

	fmt.Println(1)
	if _, err := client.PutMulti(ctx, keys, posts2); err != nil {
		fmt.Println(2)
		fmt.Println(err.Error)
	} else {
		fmt.Println(3)
		fmt.Println("YAY")
	}
	fmt.Println(4)
}

type Post struct {
	Title       string
	PublishedAt string
	Comments    int
}
