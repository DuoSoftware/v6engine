package main

import (
	"duov6.com/objectstore/client"
)

func main() {
	tmp := Account{}
	tmp.Id = "999"
	tmp.name = "SVD"
	tmp.address = "X"

	settings := make(map[string]interface{})
	settings["DB_Type"] = "ELASTIC"
	settings["Host"] = "localhost"
	settings["Port"] = "9200"

	client.GoSmoothFlow("token", "com.duosoftware.customer", "account", settings).StoreObject().WithKeyField("Id").AndStoreOne(tmp).Ok()
}

type Account struct {
	Id      string
	name    string
	address string
}
