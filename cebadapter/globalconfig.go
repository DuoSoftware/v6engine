package cebadapter

import (
"duov6.com/ceb"
"duov6.com/config"
"encoding/json"
"fmt"
)

var globalConfigs map[string][]interface{}
var funcQueue map[string]func(data []interface{})

func GetLatestGlobalConfig(name string, fn func(data []interface{})) {
	if globalConfigs == nil {
		globalConfigs = make(map[string][]interface{})
	}

	if globalConfigs[name] == nil {

		if (checkLocalCache(name)){
			fmt.Println("Store Configuration loaded from LOCAL CACHE!!!!!!!")
			fn(globalConfigs[name])
			return
		}

		if funcQueue == nil {
			funcQueue = make(map[string]func(data []interface{}))
		}
		funcQueue[name] = fn

		sendData := make(map[string]interface{})
		sendData["class"] = name

		ceb.GetClient().ExecuteCommand("getglobalconfig", sendData)
	} else {
		fn(globalConfigs[name])
	}
}

func checkLocalCache(name string) bool {
	isLoadedFromCache := false

	bytes,err := config.Get("globalConfig." + name)
	if err == nil {
		isLoadedFromCache = true
		configData := make([]interface{},0)
		err = json.Unmarshal(bytes, &configData)
		globalConfigs[name] = configData
	}

	return isLoadedFromCache
}


func setLocalCache(name string, data []interface{}){
	dataset, err := json.Marshal(data)

	if err ==nil {
		plainString := string(dataset[:len(dataset)])
		config.Save("globalConfig." + name, plainString)
	}else{
		fmt.Println(err.Error())
	}
}

func GetGlobalConfig(name string) (data []interface{}) {

	data = globalConfigs[name]
	return
}

func SetGlobalConfig(key string, value []interface{}) {
	globalConfigs[key] = value

	setLocalCache(key, value)

	if funcQueue[key] != nil {
		funcQueue[key](value)
		delete(funcQueue, key)

	}
}
