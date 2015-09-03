package main

import (
	"duov6.com/cebadapter"
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

		// cebadapter.GetLatestGlobalConfig("AutoIncrementMetaStore", func(data []interface{}) {
		// 	fmt.Println("AutoIncrementMetaStore Configuration Successfully Loaded...")
		// 	agent := cebadapter.GetAgent()
		// 	fmt.Println(data)
		// 	agent.Client.OnEvent("globalConfigChanged.AutoIncrementMetaStore", func(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
		// 		cebadapter.GetLatestGlobalConfig("AutoIncrementMetaStore", func(data []interface{}) {
		// 			fmt.Println("AutoIncrementMetaStore Configuration Successfully Updated...")
		// 		})
		// 	})
		// })
		fmt.Println("Successfully registered in CEB")
	})

	httpServer := endpoints.HTTPService{}
	go httpServer.Start()

	bulkService := endpoints.BulkTransferService{}
	go bulkService.Start()

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
