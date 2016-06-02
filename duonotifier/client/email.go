package client

import (
	"bytes"
	"duov6.com/duonotifier/messaging"
	"encoding/json"
	//"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func Send(securityToken string, subject string, domain string, class string, templateId string, defaultParams map[string]string, customParams map[string]string, recieverEmail string) messaging.NotifierResponse {
	var response messaging.NotifierResponse
	JSON_Document := getEmailJsonDoc(subject, domain, templateId, defaultParams, customParams, recieverEmail)

	url := "http://" + gethost() + ":3500/command/notification"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(JSON_Document)))
	req.Header.Set("securityToken", securityToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		//fmt.Println(err.Error())
		response.IsSuccess = false
		response.Message = "Email sending Failed!"
	} else {
		response.IsSuccess = true
		response.Message = "Email sending Successful!"
	}
	return response
}

func SendSMS(securityToken string, domain string, class string, templateId string, defaultParams map[string]string, customParams map[string]string, recieverNumber string) messaging.NotifierResponse {
	var response messaging.NotifierResponse
	JSON_Document := getSMSJsonDoc(domain, templateId, defaultParams, customParams, recieverNumber)

	url := "http://" + gethost() + ":3500/command/notification"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(JSON_Document)))
	req.Header.Set("securityToken", securityToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		//fmt.Println(err.Error())
		response.IsSuccess = false
		response.Message = "SMS sending Failed!"
	} else {
		response.IsSuccess = true
		response.Message = "SMS sending Successful!"
	}
	return response
}

func gethost() (url string) {
	content, _ := ioutil.ReadFile("agent.config")
	object := make(map[string]interface{})
	_ = json.Unmarshal(content, &object)
	tokens := strings.Split(strings.TrimSpace(object["cebUrl"].(string)), ":")
	url = tokens[0]
	return
}

func getEmailJsonDoc(subject string, domain string, templateId string, defaultParams map[string]string, customParams map[string]string, recieverEmail string) (json string) {
	json = "{\"type\":\"email\",\"to\":\"" + recieverEmail + "\",\"subject\":\"" + subject + "\",\"from\":\"DuoWorld.com <mail-noreply@duoworld.com>\",\"Namespace\": \"" + domain + "\",\"TemplateID\": \"" + templateId + "\",\"DefaultParams\": {" + getStringByMap(defaultParams) + "},\"CustomParams\": {" + getStringByMap(customParams) + "}}"
	return
}

func getSMSJsonDoc(domain string, templateId string, defaultParams map[string]string, customParams map[string]string, recieverNumber string) (json string) {
	json = "{\"type\":\"sms\",\"number\":\"" + recieverNumber + "\",\"Namespace\": \"" + domain + "\",\"TemplateID\": \"" + templateId + "\",\"DefaultParams\": {" + getStringByMap(defaultParams) + "},\"CustomParams\": {" + getStringByMap(customParams) + "}}"
	return
}

func getStringByMap(object map[string]string) (output string) {

	if len(object) == 0 || object == nil {
		output = ""
	} else {
		index := 0
		for key, value := range object {
			if index == (len(object) - 1) {
				output += "\"" + key + "\":\"" + value + "\""
			} else {
				output += "\"" + key + "\":\"" + value + "\","
			}
			index += 1
		}
	}
	return
}
