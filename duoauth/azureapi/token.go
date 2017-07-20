package azureapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var AccessToken string
var AccessTokenTime time.Time

func GetGraphApiToken() (token string, err error) {
	if AccessToken == "" {
		err = FetchAccessToken()
	} else if time.Now().Sub(AccessTokenTime).Seconds() > 3500 {
		err = FetchAccessToken()
	}

	if err == nil {
		token = AccessToken
	}

	return
}

func FetchAccessToken() (err error) {
	url := "https://login.microsoftonline.com/smoothflowio.onmicrosoft.com/oauth2/token?api-version=1.6"

	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	postBody := make(map[string]interface{})
	postBody["grant_type"] = "client_credentials"
	postBody["client_id"] = "9f957153-8e69-40f8-8b68-2572bec86910"
	postBody["code"] = "AQABAAIAAABnfiG-mA6NTae7CdWW7QfdChQ8gzEYw1uCXFo2MG9ZtVyOfBmJts_HRD3P0OwVya7FNTa3A0luh1Uv0yoKx-rxyKmquNxgqmKcFUdrqsOpkWsqMN7wEo2KLPmikVdT02RYQ6lbDYX8N9TAJJ48FofGTmcwW_wZIspbczjT9FxaUFD7qJWmwqaypCZbdRoR6DKy4dFi7EL-eyF-qea81qlQP9WxP5IWu3bEupm7w2P4eyY-BiCgUvLrj5z_Ff7ZJcnxQt1YoZnbt08cyX4q_V_0OH1zSBdTdOVskW6sNbKK6-58qH1bWfDFlbwJb6RzY7hveQuiRUv03vJQJuLtJYZWm4TF22UE2u_HA-cEltE7KpamATBpgDNU-L-YsfO7wpQjdi8wxMdY0mzTIYI5N2tWU56Qfn2aRiRrl0A1WM38weSAr6b4Dz9cy0D_m-QoO0m_QIPuRGoqrUcugPZInDA58VvkkFkbwn1MFStgGxvy8-sPfAJZuRkh026Xgr5wthGlQ8zYK7w23jCGQloo7jMYwUc35Pe0cL26_NEphubqm8MbrY8nQ6W-aKk23xejyUgGamBBZHioc82CvNHgAaLOGGFfaHFu2cfAVUL7P41dum8t7CGCgQwSbu9VneVVkl8xhjJDkNQQEUgaiDciZt62wx_MUfe3cWo0LXuM1_CWuyAA&session_state=62a7e3cb-8254-423a-9ec7-c67be94af47c"
	postBody["redirect_uri"] = "http://localhost/reply"
	postBody["resource"] = "https://graph.windows.net"
	postBody["client_secret"] = "/i3L6iNPkHi8KrHHxxHHDN5WfvUNco39xq8mO5rPssk="

	err, body := HTTP_FORM_POST(url, postBody)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		data := make(map[string]interface{})
		err = json.Unmarshal(body, &data)
		AccessToken = data["access_token"].(string)
		AccessTokenTime = time.Now()
	}
	return
}

func HTTP_FORM_POST(urll string, postBody map[string]interface{}) (err error, body []byte) {

	form := url.Values{}
	for key, value := range postBody {
		form.Add(key, value.(string))
	}

	req, err := http.NewRequest("POST", urll, strings.NewReader(form.Encode()))

	req.PostForm = form

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err = errors.New("Connection Failed!")
	} else {
		defer resp.Body.Close()
		body, _ = ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			err = errors.New(string(body))
		}
	}
	return
}
