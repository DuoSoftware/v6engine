package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	fmt.Println("START")
	for x := 0; x < 1000000; x++ {
		fmt.Println(x)
		insert()
	}
	fmt.Println("END")

}

func insert() {
	//securityToken := "123"
	//domain := "loadtest2.jay.com"
	//class := "test2"

	// data := make(map[string]interface{})
	// data["GUAccountID"] = "-888"
	// data["accountNo"] = "1243789076432144231"

	//JSON_Document := CreateJSON(data, "GUAccountID")
	JSON_Document := `{
	    "AppCode":"appcode",
	    "ProcessCode":"BULKACCCREATE",
	    "SessionID":"789",
	    "SecurityToken":"3f082f3f1370889d5b15839c5185de7e",
	    "Log":"log",
	    "Namespace":"45.55.83.253",
	    "JSONData":"{\"InSessionID\":\"789\",\"InSecurityToken\":\"343434\",\"InLog\":\"log\",\"InNamespace\":\"45.55.83.253\",\"InAccountNo\":\"78\"}"
	}`
	//url := "http://" + "192.168.1.194:3000" + "/" + domain + "/" + class
	url := "http://localhost:8093/processengine/InvokeFlow/appcode/BULKACCCREATE/789"
	//url := "http://192.168.1.194:8787/smoothflow/Invoke"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(JSON_Document)))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		var objArray interface{}
		json.Unmarshal(body, &objArray)
		fmt.Println(objArray)
		fmt.Println("YAY!")
	}
	defer resp.Body.Close()
}

func CreateJSON(input map[string]interface{}, primaryKeyFrield string) (output string) {
	output += "{\"Object\":{"

	index := 0
	for key, value := range input {
		if index == len(input)-1 {
			output += "\"" + key + "\":\"" + value.(string) + "\""
		} else {
			output += "\"" + key + "\":\"" + value.(string) + "\","
		}
		index += 1
	}

	output += "}, \"Parameters\":{\"KeyProperty\":\"" + primaryKeyFrield + "\"}}"

	return
}
