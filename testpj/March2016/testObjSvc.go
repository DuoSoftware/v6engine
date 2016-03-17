package main

import (
	"duov6.com/consoleworker/objectstore"
	"fmt"
)

func main() {
	data, err := objectstore.GetAll("tasks.serviceconsole.payload", "f0fe0c98741cbfefc6a63e9826ab1e1b-Book1")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(data)
	}
}
