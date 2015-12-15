package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/bigtable"
	"io/ioutil"
	"strings"
)

func main() {
	fmt.Println("Start!")
	//put()
	getAll()
	//update()
	//delete()
}

func getClient() *bigtable.Client {
	jsonKey, err := ioutil.ReadFile("TestProject-ee4c6215cc69.json")
	if err != nil {
		fmt.Println(err.Error())
	}

	config, err := google.JWTConfigFromJSON(
		jsonKey,
		bigtable.Scope,
	) // or bigtable.AdminScope, etc.

	if err != nil {
		fmt.Println(err.Error())
	}

	ctx := context.Background()
	client, err := bigtable.NewClient(ctx, "tidy-groove-113806", "asia-east1-b", "test1-supun-bigtable", cloud.WithTokenSource(config.TokenSource(ctx)))

	if err != nil {
		fmt.Println(err.Error())
	}

	return client
}

func getAdminClient() *bigtable.AdminClient {
	jsonKey, err := ioutil.ReadFile("TestProject-ee4c6215cc69.json")
	if err != nil {
		fmt.Println(err.Error())
	}

	config, err := google.JWTConfigFromJSON(
		jsonKey,
		bigtable.Scope,
		bigtable.AdminScope,
	) // or bigtable.AdminScope, etc.

	if err != nil {
		fmt.Println(err.Error())
	}

	ctx := context.Background()
	client, err := bigtable.NewAdminClient(ctx, "tidy-groove-113806", "asia-east1-b", "test1-supun-bigtable", cloud.WithTokenSource(config.TokenSource(ctx)))

	if err != nil {
		fmt.Println(err.Error())
	}

	return client
}

func put() {
	ctx := context.Background()
	adminClient := getAdminClient()

	table := "com.duosoftware.com"

	tableNames, err := adminClient.Tables(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	} else {
		fmt.Println(tableNames)
	}

	tableString := ""

	for _, name := range tableNames {
		tableString += name + "|"
	}

	if !strings.Contains(tableString, table) {
		fmt.Println("Creating table")
		err = adminClient.CreateTable(ctx, table)
		if err != nil {
			fmt.Println(1)
			fmt.Println(err.Error())
		}
	}

	err = adminClient.CreateColumnFamily(ctx, table, "settings")
	if err != nil {
		fmt.Println(1.5)
		fmt.Println(err.Error())
	}

	client := getClient()
	tbl := client.Open(table)
	mut := bigtable.NewMutation()
	v1 := "prasad"
	v2 := 28
	v3 := true
	v11, _ := json.Marshal(v1)
	v22, _ := json.Marshal(v2)
	v33, _ := json.Marshal(v3)

	mut.Set("settings", "col1", bigtable.Now(), v11)
	mut.Set("settings", "col2", bigtable.Now(), v22)
	mut.Set("settings", "col3", bigtable.Now(), v33)
	err = tbl.Apply(ctx, "com.duosoftware.com.settings.1", mut)
	if err != nil {
		fmt.Println(2)
		fmt.Println(err.Error())
	}
}

func getAll() {
	ctx := context.Background()
	client := getClient()
	tbl := client.Open("com.duosoftware.com")
	// Read all the rows starting with "com.google.",
	// but only fetch the columns in the "links" family.

	var data []map[string]interface{}

	rr := bigtable.PrefixRange("")
	err := tbl.ReadRows(ctx, rr, func(r bigtable.Row) bool {
		// do something with r
		for _, v := range r {
			var single map[string]interface{}
			single = make(map[string]interface{})
			for _, o := range v {
				fmt.Println(o.Row)
				single[o.Column] = o.Value
			}

			data = append(data, single)

		}
		return true // keep going
	}, bigtable.RowFilter(bigtable.FamilyFilter("settings")))

	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("************")
	fmt.Println(data)

}

func update() {
	ctx := context.Background()
	table := "com.duos.com"
	client := getClient()
	tbl := client.Open(table)
	mut := bigtable.NewMutation()
	mut.Set("settings", "col1", bigtable.Now(), []byte("3dd"))
	mut.Set("settings", "col2", bigtable.Now(), []byte("4dd"))
	err := tbl.Apply(ctx, "com.duo.com.settings.7", mut)
	if err != nil {
		fmt.Println(2)
		fmt.Println(err.Error())
	}

}

func delete() {
	ctx := context.Background()
	table := "com.duosoftware.com"
	client := getClient()
	tbl := client.Open(table)
	mut := bigtable.NewMutation()
	mut.DeleteRow()
	err := tbl.Apply(ctx, "com.duosoftware.com.settings.1", mut)
	if err != nil {
		fmt.Println(2)
		fmt.Println(err.Error())
	}
}
