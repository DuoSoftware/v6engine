package documentcache

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func verifyDirectory() {
	_, err := os.Stat("/tmp/doc_cache")
	if err != nil {
		path := "/tmp/doc_cache"
		os.Mkdir(path, 0777)
		fmt.Println("Seems Fresh! Cache Directory Created at : " + path)
	}
}

func writeToFile(key string, ttl int, body interface{}) (status bool) {
	var storeObject map[string]interface{}
	storeObject = make(map[string]interface{})
	storeObject["data"] = body
	storeObject["ttl"] = getExpiryTime(ttl)
	result, err := json.Marshal(storeObject)
	if err != nil {
		fmt.Println(err.Error())
		return
	} else {
		err := ioutil.WriteFile(("/tmp/doc_cache/" + getFileName(key)), result, 0666)
		if err != nil {
			fmt.Println("Error : " + err.Error())
			status = false
		} else {
			status = true
		}
	}

	return
}

func readFromFile(key string) (body interface{}, status bool) {
	filePath := "/tmp/doc_cache/" + getFileName(key)
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error : " + err.Error())
		body = nil
		status = false
	} else {
		var fileContent map[string]interface{}
		fileContent = make(map[string]interface{})
		err = json.Unmarshal(content, &fileContent)
		if err != nil {
			fmt.Println("Error : " + err.Error())
			body = nil
			status = false
		} else {
			fileTime := getTimeFromString(fileContent["ttl"].(string))
			fileStatus := checkDataValidity(fileTime)
			if fileStatus {
				body = fileContent["data"]
				status = true
			} else {
				deleteFile(filePath)
				body = nil
				status = false
			}
		}
	}

	return
}

func getFileName(key string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(key)))
}

func getExpiryTime(ttl int) time.Time {
	additionalTime := time.Duration(ttl) * time.Minute
	nowTime := time.Now().Local().Add(additionalTime)
	return nowTime
}

func checkDataValidity(fileTime time.Time) (status bool) {
	nowTime := time.Now().Local()
	timeDifference := nowTime.Sub(fileTime)
	if timeDifference > 0 {
		status = false
	} else {
		status = true
	}
	return
}

func getTimeFromString(timestring string) (timestamp time.Time) {
	const template = "2006-01-02T15:04:05Z07:00"
	timestamp, err := time.Parse(template, timestring)
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}

func deleteFile(path string) {
	err := os.Remove(path)
	if err != nil {
		fmt.Println("Error : " + err.Error())
	}
}
