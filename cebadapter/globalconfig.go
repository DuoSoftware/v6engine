package cebadapter

import (
	"bytes"
	"duov6.com/ceb"
	"duov6.com/config"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var globalConfigs map[string][]interface{}
var funcQueue map[string]func(data []interface{})

func GetLatestGlobalConfig(name string, fn func(data []interface{})) {
	if globalConfigs == nil {
		globalConfigs = make(map[string][]interface{})
	}

	if globalConfigs[name] == nil {

		if checkLocalCache(name) {
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

	bytes, err := config.Get("globalConfig." + name)
	if err == nil {
		isLoadedFromCache = true
		configData := make([]interface{}, 0)
		err = json.Unmarshal(bytes, &configData)
		globalConfigs[name] = configData
	}

	return isLoadedFromCache
}

func GetGlobalConfigFromREST(name string) bool {
	if globalConfigs == nil {
		globalConfigs = make(map[string][]interface{})
	}

	//get ceb url...
	var cebUrl string
	data, err := ioutil.ReadFile("./agent.config")

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	settings := make(map[string]interface{})
	_ = json.Unmarshal(data, &settings)
	cebUrl = settings["cebUrl"].(string)

	//Read Configs
	url := cebUrl + "/command/getglobalconfig/"
	postBody := `{"class":"` + name + `"}`

	url = strings.Replace(url, "5000", "3500", -1)
	url = "http://" + url
	fmt.Println(url)
	var body []byte
	err, body = GetConfigRest(url, []byte(postBody))
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	restData := make(map[string]interface{})
	_ = json.Unmarshal(body, &restData)

	configData := restData["data"].(map[string]interface{})["data"].(map[string]interface{})["config"].([]interface{})
	SetGlobalConfig(name, configData)
	return true
}

func setLocalCache(name string, data []interface{}) {
	dataset, err := json.Marshal(data)

	if err == nil {
		plainString := string(dataset[:len(dataset)])
		config.Save("globalConfig."+name, plainString)
	} else {
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

//helper functions
func GetConfigRest(url string, JSON_DATA []byte) (err error, body []byte) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(JSON_DATA))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("SecurityToken", "ignore")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err = errors.New("Connection Failed!")
	} else {
		defer resp.Body.Close()
		body, _ = ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			err = errors.New(string(body))
		}
	}
	return
}
