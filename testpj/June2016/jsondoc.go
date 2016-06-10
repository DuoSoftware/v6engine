package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	t1()
}

func t1() {
	//dd := "{\"Id\":\"-999\", \"Name\":\"Shehan\"}"
	dd := "[{\"Id\":\"-999\", \"Name\":\"Shehan\"},{\"Id\":\"-888\", \"Name\":\"Prasad\"}]"
	if string(dd[0]) == "[" {
		fmt.Println("FUCK ME")
	}
	var x []map[string]interface{}
	err := json.Unmarshal([]byte(dd), &x)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(x)
	}
}

func t2() {
	//create the insert object
	var parameterObj InsParameters
	parameterObj.KeyProperty = "Id"

	// creating the object structure for the parameters for objectstore insert
	var insertObj = new(InsertTemplate)
	objectttt := make(map[string]interface{})
	objectttt["Id"] = "-999"
	objectttt["Name"] = "Shehan"
	insertObj.Object = objectttt
	insertObj.Parameters = parameterObj

	// converting the struct to JSON
	convertedObj, err := json.Marshal(insertObj)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(string(convertedObj))

	Invoke(string(convertedObj))
}

type InsertTemplate struct {
	Object     map[string]interface{}   `json:"Object"`
	Objects    []map[string]interface{} `json:"Objects"`
	Parameters InsParameters
}

type InsParameters struct {
	KeyProperty string `json:"KeyProperty"`
}

func Invoke(data string) {

	securityToken := "ignore"
	domain := "nadanada"
	class := "hueheuheu1"

	url := "http://localhost:3000/" + domain + "/" + class + "?securityToken=" + securityToken

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	req.Header.Set("securityToken", securityToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil && resp.StatusCode != 200 {
		fmt.Println(err.Error())
	} else {
		fmt.Println(resp)
	}
	defer resp.Body.Close()
}
