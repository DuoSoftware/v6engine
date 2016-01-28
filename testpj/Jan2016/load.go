package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("START")
	for x := 0; x < 1000; x++ {
		fmt.Println(x)
		insert()
	}
	fmt.Println("END")

}

func insert() {
	securityToken := "123"
	domain := "loadtest2.jay.com"
	class := "test2"

	data := make(map[string]interface{})
	data["GUAccountID"] = "-888"
	data["accountNo"] = "1243789076432144231"

	JSON_Document := CreateJSON(data, "GUAccountID")

	url := "http://" + "192.168.1.194:3000" + "/" + domain + "/" + class
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(JSON_Document)))
	req.Header.Set("securityToken", securityToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	} else {
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
