package main

import (
	"duov6.com/fws"
	"duov6.com/objectstore/endpoints"
	"duov6.com/objectstore/unittesting"
	"fmt"
)

func main() {
	var isUnitTestMode bool = false

	if isUnitTestMode {
		unittesting.Start()
	} else {
		splash()
		initialize()
	}
}

func initialize() {
	fws.Attach("ObjectStore")

	httpServer := endpoints.HTTPService{}
	httpServer.Start()
}

func splash() {

	fmt.Println("")
	fmt.Println("")
	fmt.Println("                                                 ~~")
	fmt.Println("    ____             _____ __                  | ][ |")
	fmt.Println("   / __ \\__  ______ / ___// /_____  ________     ~~")
	fmt.Println("  / / / / / / / __ \\__ \\/ __/ __ \\/ ___/ _ \\")
	fmt.Println(" / /_/ / /_/ / /_/ /__/ / /_/ /_/ / /  /  __/")
	fmt.Println("/_____/\\__,_/\\____/____/\\__/\\____/_/   \\___/ ")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("")
}
