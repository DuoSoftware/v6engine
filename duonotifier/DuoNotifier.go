package main

import (
	"duov6.com/cebadapter"
	"duov6.com/common"
	"duov6.com/duonotifier/endpoints"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	splash()
	initializeCEBConfig()
	/*httpServer := endpoints.HTTPService{}
	httpServer.Start()*/
}

func initializeCEBConfig() {
	endpoints.StartTime = time.Now()
	common.VerifyConfigFiles()
	initializeSettingsFile()
	inititalizeObjectStoreConfig()
}

func initializeSettingsFile() {
	From := os.Getenv("SMTP_ADDRESS")
	content, err := ioutil.ReadFile("settings.config")
	if err != nil {
		data := make(map[string]interface{})
		if From == "" {
			data["From"] = "DuoWorld.com <mail-noreply@duoworld.com>"
		} else {
			data["From"] = From
		}
		dataBytes, _ := json.Marshal(data)
		_ = ioutil.WriteFile("settings.config", dataBytes, 0666)
	} else {
		vv := make(map[string]interface{})
		_ = json.Unmarshal(content, &vv)
		if From != "" {
			vv["From"] = From
		}
		dataBytes, _ := json.Marshal(vv)
		_ = ioutil.WriteFile("settings.config", dataBytes, 0666)
		fmt.Println(vv)
	}
}

/*func inititalizeObjectStoreConfig() {
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
}*/

func inititalizeObjectStoreConfig() {
	cebadapter.Attach("DuoNotifier", func(s bool) {
		cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
			fmt.Println()
			fmt.Println(data)
			fmt.Println()
			fmt.Println("Store Configuration Successfully Loaded...")
			agent := cebadapter.GetAgent()

			agent.Client.OnEvent("globalConfigChanged.StoreConfig", func(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
				cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
					fmt.Println("Store Configuration Successfully Updated...")
				})
			})
		})
		fmt.Println("Successfully registered DuoNotifier in CEB")
	})

	httpServer := endpoints.HTTPService{}
	go httpServer.Start()

	forever := make(chan bool)
	<-forever
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
