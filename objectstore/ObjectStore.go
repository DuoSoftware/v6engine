package main

import (
	"duov6.com/objectstore/endpoints"
	"duov6.com/objectstore/unittesting"
)

func main() {
	var isUnitTestMode bool = false

	if isUnitTestMode {
		unittesting.Start()
	} else {
		initialize()
	}
}

func initialize() {
	httpServer := endpoints.HTTPService{}
	httpServer.Start()
}
