package main

import (
	"duov6.com/cebadapter"
	"duov6.com/duonotifier/endpoints"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func main() {
	splash()
	initializeCEBConfig()
	httpServer := endpoints.HTTPService{}
	httpServer.Start()
}

func initializeCEBConfig() {
	initializeSettingsFile()
	inititalizeObjectStoreConfig()
}

func initializeSettingsFile() {
	content, err := ioutil.ReadFile("settings.config")
	if err != nil {
		data := make(map[string]interface{})
		data["From"] = "DuoWorld.com <mail-noreply@duoworld.com>"
		dataBytes, _ := json.Marshal(data)
		_ = ioutil.WriteFile("settings.config", dataBytes, 0666)
	} else {
		vv := make(map[string]interface{})
		_ = json.Unmarshal(content, &vv)
		fmt.Println(vv)
	}
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
