package main

import (
	"duov6.com/common"
	"encoding/json"
	"fmt"
)

func main() {
	fmt.Println(bb("testuserone.developer.duoworld.com", "02ae443edd17d9eb9e0bd1c4078d233e"))
}

func bb(domain, securityToken string) (status bool) {
	url := "http://" + domain + "/apis/ratingservice/process/" + domain + "/user/1/tenant"

	headers := make(map[string]string)
	headers["securityToken"] = securityToken

	err, bodyBytes := common.HTTP_GET(url, headers, false)

	responseMap := make(map[string]interface{})

	if err != nil {
		json.Unmarshal(([]byte(err.Error())), &responseMap)
		status = false
	} else {
		json.Unmarshal(bodyBytes, &responseMap)
		status = true
	}

	fmt.Println(responseMap)

	return
}
