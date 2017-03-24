package main

import (
	"duov6.com/common"
	"encoding/json"
	"fmt"
	//"strings"
)

func main() {
	url := "https://www.googleapis.com/oauth2/v1/userinfo?alt=json&access_token=ya29.Ci9rA_Yz08nshMvTwy1wB882Q7DD6qUdFBTKr-9qX8IFOhDI55cbK98mZ-fb5QujnA"
	err, body := common.HTTP_GET(url, nil, false)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		data := make(map[string]interface{})
		_ = json.Unmarshal(body, &data)
		fmt.Println(data["id"].(string))
		fmt.Println(data["email"].(string))
	}
}
