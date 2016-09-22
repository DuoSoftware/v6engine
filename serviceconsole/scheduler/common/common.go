package common

import (
	"encoding/json"
	"io/ioutil"
)

func GetSettings() (object map[string]string) {
	content, _ := ioutil.ReadFile("settings.config")
	object = make(map[string]string)
	_ = json.Unmarshal(content, &object)
	return
}
