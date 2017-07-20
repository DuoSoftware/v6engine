package main

import (
	"bytes"
	"duov6.com/common"
	//"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	//"strings"
)

func main() {
	url := "https://xmplod.auth0.com/tokeninfo"
	fmt.Println("URL:>", url)
	//token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJodHRwczovL3htcGxvZC5hdXRoMC5jb20vIiwic3ViIjoiYXV0aDB8NTc0ODIxMDE2NDMyM2UyNTA1Njc1NjNmIiwiYXVkIjoiSTFydHZYRndZdXFCYnEwajdXRjg5UXlmaUlKQU9aMkYiLCJleHAiOjE0NjQ3MDM3NTAsImlhdCI6MTQ2NDY2Nzc1MH0.YoanmIQOipvnbrHTucSw-wC8p2KS5xYqasUHsgSNa0E"
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkbW4iOiJkdW93b3JsZC5jb20iLCJlbWwiOiJ4aWtlQG1zd29yay5ydSIsImlzcyI6ImR1b3dvcmxkLmNvbSIsInNjb3BlIjp7fSwic3QiOiJhNGJlZTUwYTM1ZGY3N2M0MmQxZThjZjdiZGI2OTAxOSIsInVpZCI6IjNlNDQ0ZGZjNGVmMTAwODhiZTY0YzlkM2JjMjMxYThkIn0=.P4vyv/cTibvkkNnIFPrd0BxyQJEMDO5tJN6bUI4bMHg="
	var jsonStr = []byte(`{"id_token":"` + token + `"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		//panic(err)
		fmt.Println(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	o := make(map[string]interface{})
	err1 := json.Unmarshal(body, &o)
	if err1 != nil {
		fmt.Println(err1)
	} else {
		fmt.Println(o)
	}
	/*
		array := strings.Split(token, ".")
		str := array[1]
		fmt.Println(str)

		data, _ := base64.StdEncoding.DecodeString(str)

		fmt.Println(string(data))
		jwt := make(map[string]interface{})
		strJwt := string(data)
		//fmt.Println(strings.LastIndex(strJwt, "}") + 1)
		//fmt.Println(len(strJwt))
		if len(strJwt) != (strings.LastIndex(strJwt, "}") + 1) {
			strJwt += "}"
		}
		err1 = json.Unmarshal([]byte(strJwt), &jwt)
		if err1 != nil {
			fmt.Println(err1)
		} else {
			fmt.Println(jwt)
		}*/
	fmt.Println(common.JwtUnload(token))
}
