package main

import (
	"duov6.com/DuoEtlObjectCollector/logger"
	"duov6.com/DuoEtlObjectCollector/service"
	"duov6.com/cebadapter"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func main() {
	draw()
	var configData []interface{}

	//Loading CEB
	forever := make(chan bool)
	cebadapter.Attach("DuoEtl", func(s bool) {
		cebadapter.GetLatestGlobalConfig("DuoEtl", func(data []interface{}) {
			logger.Log("Duo ETL Service Configuration Successfully Loaded...")
			if data != nil {
				configData = data
				forever <- false
			}
		})
		logger.Log("Successfully registered in CEB")
	})

	<-forever

	fmt.Println(configData)
	writePathConfigFile(configData[0])
	service.Start()
}

func draw() {
	logger.Log("Starting Duo ETL Json Stack Collector.....")
	logger.Log("\n")
	logger.Log("\n")
	logger.Log("______             _____ _____ _           ____ ")
	logger.Log("|  _  \\           |  ___|_   _| |         / ___|")
	logger.Log("| | | |_   _  ___ | |__   | | | |  __   _/ /___ ")
	logger.Log("| | | | | | |/ _ \\|  __|  | | | |  \\ \\ / / ___ \\")
	logger.Log("| |/ /| |_| | (_) | |___  | | | |___\\ V /| \\_/ |")
	logger.Log("|___/  \\__,_|\\___/\\____/  \\_/ \\_____/\\_/ \\_____/")
	logger.Log("\n")
	logger.Log("\n")

}

func writePathConfigFile(obj interface{}) {
	var path string
	for key, value := range obj.(map[string]interface{}) {
		if key == "DataPath" {
			path = value.(string)
		}
	}
	var fileContent map[string]string
	fileContent = make(map[string]string)
	fileContent["Path"] = path
	byte, _ := json.Marshal(fileContent)
	err := ioutil.WriteFile("config.config", byte, 0666)
	if err != nil {
		logger.Log(err.Error())
	}
}
