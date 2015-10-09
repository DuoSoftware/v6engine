package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	fmt.Println("Staring Elastic Data Wiper!")

	//Read configurations
	configContent, _ := ioutil.ReadFile("configuration.config")
	var configItemArray map[string]interface{}
	configItemArray = make(map[string]interface{})
	_ = json.Unmarshal(configContent, &configItemArray)

	//Read Delete List
	deleteContent, _ := ioutil.ReadFile("deleteList.config")
	var deleteItemArray map[string]interface{}
	deleteItemArray = make(map[string]interface{})
	_ = json.Unmarshal(deleteContent, &deleteItemArray)

	elasticPath := configItemArray["path"].(string)

	for indexName, classValues := range deleteItemArray {
		for _, className := range classValues.([]interface{}) {
			status := delete(elasticPath, indexName, className.(string))
			fmt.Print("Deleting Index : " + indexName + " Type : " + className.(string) + " -> ")
			if status {
				fmt.Println("SUCCESS!")
			} else {
				fmt.Println("FAILED!")
			}
		}
	}

}

func delete(path string, namespace string, class string) (status bool) {
	status = false
	var url string

	if class == "*" {
		url = "http://" + path + "/" + namespace + "/"
	} else {
		url = "http://" + path + "/" + namespace + "/" + class + "/"
	}

	req, err := http.NewRequest("DELETE", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		status = false
	} else {
		status = true
	}
	defer resp.Body.Close()

	return status
}
