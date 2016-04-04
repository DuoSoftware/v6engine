package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func GetConfigurations() (config map[string]interface{}) {
	content, err := ioutil.ReadFile("settings.config")
	err = json.Unmarshal(content, &config)
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}

func PostHTTPRequest(url string, data interface{}) (err error) {
	securityToken := "ignore"

	JSON_Document, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(JSON_Document))
	req.Header.Set("securityToken", securityToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	return
}

func GetTime() (retTime string) {
	currentTime := time.Now().Local()
	year := strconv.Itoa(currentTime.Year())
	month := strconv.Itoa(int(currentTime.Month()))
	day := strconv.Itoa(currentTime.Day())
	hour := strconv.Itoa(currentTime.Hour())
	minute := strconv.Itoa(currentTime.Minute())
	second := strconv.Itoa(currentTime.Second())

	retTime = (year + month + day + hour + minute + second)

	return
}

func GetTimeReadable() (retTime string) {
	currentTime := time.Now().Local()
	year := strconv.Itoa(currentTime.Year())
	month := strconv.Itoa(int(currentTime.Month()))
	day := strconv.Itoa(currentTime.Day())
	hour := strconv.Itoa(currentTime.Hour())
	minute := strconv.Itoa(currentTime.Minute())
	second := strconv.Itoa(currentTime.Second())

	retTime = (year + "-" + month + "-" + day + " " + hour + ":" + minute + ":" + second)

	return
}
