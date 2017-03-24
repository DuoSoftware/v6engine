package main

import (
	"duov6.com/RequestTester/endpoints"
	"fmt"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("Staring Request Tester...")

	httpServer := endpoints.HTTPService{}
	httpServer.Start()
}
