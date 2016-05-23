package main

import (
	"duov6.com/cebadapter"
	"duov6.com/gorest"
	gg "duov6.com/objectstore/endpoints2"
	"duov6.com/objectstore/unittesting"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	arguments := os.Args[1:]
	var IsLoggable bool
	var IsJsonStackEnabled bool
	if len(arguments) > 0 {

		if strings.EqualFold(arguments[0], "true") {
			IsLoggable = true
		} else if strings.EqualFold(arguments[1], "true") {
			IsJsonStackEnabled = true
		} else {
			IsLoggable = false
			IsJsonStackEnabled = false
		}

	} else {
		IsLoggable = false
		IsJsonStackEnabled = false
	}

	var isUnitTestMode bool = false

	if isUnitTestMode {
		unittesting.Start()
	} else {
		splash()
		initialize(IsLoggable, IsJsonStackEnabled)
	}
}

func initialize(isLogEnabled bool, isJsonStack bool) {

	cebadapter.Attach("ObjectStore", func(s bool) {
		cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
			fmt.Println("Store Configuration Successfully Loaded...")
			fmt.Println(data)
			agent := cebadapter.GetAgent()

			agent.Client.OnEvent("globalConfigChanged.StoreConfig", func(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
				cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
					fmt.Println("Store Configuration Successfully Updated...")
				})
			})
		})
		fmt.Println("Successfully registered in CEB")
	})
	go runRestFul()
	//httpServer := endpoints.HTTPService{}
	//go httpServer.Start(isLogEnabled, isJsonStack)

	//bulkService := endpoints.BulkTransferService{}
	//go bulkService.Start()

	forever := make(chan bool)
	<-forever
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

func runRestFul() {
	gorest.RegisterService(new(gg.HTTPService2))

	err := http.ListenAndServe(":3000", gorest.Handle())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

}
