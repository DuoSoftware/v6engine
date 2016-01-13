package main

import (
	"duov6.com/cebadapter"
	"duov6.com/duonotifier/endpoints"
	"fmt"
)

func main() {
	initializeCEBConfig()
	httpServer := endpoints.HTTPService{}
	httpServer.Start()
}

func initializeCEBConfig() {
	inititalizeObjectStoreConfig()
	initializeDuoNotifierConfig()
}

func initializeDuoNotifierConfig() {
	forever := make(chan bool)
	cebadapter.Attach("DuoNotifier", func(s bool) {
		cebadapter.GetLatestGlobalConfig("DuoNotifier", func(data []interface{}) {
			fmt.Println("DuoNotifier Configuration Successfully Loaded...")
			fmt.Println(data)
			if data != nil {
				forever <- false
			}
			agent := cebadapter.GetAgent()
			agent.Client.OnEvent("globalConfigChanged.DuoNotifier", func(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
				cebadapter.GetLatestGlobalConfig("DuoNotifier", func(data []interface{}) {
					fmt.Println("Store Configuration Successfully Updated...")
				})
			})
		})
		fmt.Println("Successfully registered DuoNotifier in CEB")
	})

	<-forever
	return
}

func inititalizeObjectStoreConfig() {
	forever := make(chan bool)
	cebadapter.Attach("ObjectStore", func(s bool) {
		cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
			fmt.Println("Store Configuration Successfully Loaded...")
			fmt.Println(data)
			if data != nil {
				forever <- false
			}
			agent := cebadapter.GetAgent()
			agent.Client.OnEvent("globalConfigChanged.StoreConfig", func(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
				cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
					fmt.Println("Store Configuration Successfully Updated...")
				})
			})
		})
		fmt.Println("Successfully registered ObjectStore in CEB")
	})

	<-forever
	return
}
