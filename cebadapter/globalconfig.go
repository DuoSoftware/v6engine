package cebadapter

import ("duov6.com/ceb")

var globalConfigs map[string][]interface{}
var funcQueue map[string]func(data []interface{})

func GetLatestGlobalConfig(name string, fn func(data []interface{})) {
	if globalConfigs == nil {
		globalConfigs = make(map[string][]interface{})
	}

	if globalConfigs[name] == nil {
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

func GetGlobalConfig(name string) (data []interface{}) {

	data = globalConfigs[name]
	return
}

func SetGlobalConfig(key string, value []interface{}) {
	globalConfigs[key] = value

	if funcQueue[key] != nil {
		funcQueue[key](value)
		delete(funcQueue, key)

	}
}
