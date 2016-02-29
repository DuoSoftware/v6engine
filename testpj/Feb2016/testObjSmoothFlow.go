package main

import (
	"duov6.com/objectstore/client"
)

func main() {

	object := make(map[string]interface{})
	object["id"] = "1134"
	object["field1"] = "Value"

	settings := make(map[string]interface{})
	settings["DB_Type"] = "ELASTIC"
	settings["Host"] = "localhost"
	settings["Port"] = "9200"

	client.GoSmoothFlow("token", "com.duosoftware.customer", "account", settings).StoreObject().WithKeyField("id").AndStoreOne(tmp).Ok()
}
