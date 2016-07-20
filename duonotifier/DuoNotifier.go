package main

import (
	"duov6.com/cebadapter"
	"duov6.com/duonotifier/endpoints"
	"fmt"
)

func main() {
	splash()
	initializeCEBConfig()
	httpServer := endpoints.HTTPService{}
	httpServer.Start()
}

func initializeCEBConfig() {
	inititalizeObjectStoreConfig()
}

func inititalizeObjectStoreConfig() {
	forever := make(chan bool)
	cebadapter.Attach("DuoNotifier", func(s bool) {
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
		fmt.Println("Successfully registered DuoNotifier in CEB")
	})

	<-forever
	return
}

func splash() {

	fmt.Println()
	fmt.Println("______             _   _       _   _  __ _          ")
	fmt.Println("|  _  \\           | \\ | |     | | (_)/ _(_)          ")
	fmt.Println("| | | |_   _  ___ |  \\| | ___ | |_ _| |_ _  ___ _ __ ")
	fmt.Println("| | | | | | |/ _ \\| . ` |/ _ \\| __| |  _| |/ _ \\ '__|")
	fmt.Println("| |/ /| |_| | (_) | |\\  | (_) | |_| | | | |  __/ |   ")
	fmt.Println("|___/  \\__,_|\\___/\\_| \\_/\\___/ \\__|_|_| |_|\\___|_|   ")
	fmt.Println()

}
