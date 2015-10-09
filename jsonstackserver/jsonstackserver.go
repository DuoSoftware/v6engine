package main

import (
	"duov6.com/jsonstackserver/service"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"
)

func main() {
	draw()
	if getETLHost() == "err" {
		return
	}

	service.Log("ETL Host : " + getETLHost())

	executeTask()
	go delaySecond(200)
	select {}

}

func executeTask() {
	fmt.Println("\n")
	fmt.Println("\n")
	service.Log("Executing Stack Transfer @ " + getTime())
	if getETLHost() == "err" {
		return
	}
	service.Start(getETLHost())
}

func getTime() (retTime string) {
	currentTime := time.Now().Local()
	year := strconv.Itoa(currentTime.Year())
	month := strconv.Itoa(int(currentTime.Month()))
	day := strconv.Itoa(currentTime.Day())
	hour := strconv.Itoa(currentTime.Hour())
	minute := strconv.Itoa(currentTime.Minute())
	second := strconv.Itoa(currentTime.Second())
	retTime = (year + "-" + month + "-" + day + "  " + hour + ":" + minute + ":" + second)
	return
}

func delaySecond(n time.Duration) {
	for _ = range time.Tick(n * time.Second) {
		executeTask()
	}
}

func getETLHost() (url string) {
	content, err := ioutil.ReadFile("config.config")
	if err != nil {
		fmt.Println(err.Error())
		url = "err"
	} else {
		var settings map[string]interface{}
		settings = make(map[string]interface{})
		err = json.Unmarshal(content, &settings)
		if err != nil || settings["ETLHost"] == nil {
			fmt.Println(err.Error())
			url = "err"
		} else {
			url = settings["ETLHost"].(string)
		}
	}

	return
}

func draw() {
	service.Log("Starting Json Stack Server for Duo V6.....")
	service.Log("\n")
	service.Log("\n")
	service.Log("______               ___                  _____ _             _ ")
	service.Log("|  _  \\             |_  |                /  ___| |           | |   ")
	service.Log("| | | |_   _  ___     | | ___  ___  _ __ \\ `--.| |_ __ _  ___| | __")
	service.Log("| | | | | | |/ _ \\    | |/ __|/ _ \\| '_ \\ `--. \\ __/ _` |/ __| |/ /")
	service.Log("| |/ /| |_| | (_) /\\__/ /\\__ \\ (_) | | | /\\__/ / || (_| | (__|   < ")
	service.Log("|___/  \\__,_|\\___/\\____/ |___/\\___/|_| |_\\____/ \\__\\__,_|\\___|_|\\_\\")
	service.Log("\n")
	service.Log("\n")
}
