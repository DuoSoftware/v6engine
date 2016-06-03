package client

import (
	"bytes"
	"duov6.com/duonotifier/messaging"
	"duov6.com/objectstore/client"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func Notify(securityToken, EmailTemplateId, SmsTemplateId, recieverEmail string, defaultParams map[string]string, customParams map[string]string) messaging.NotifierResponse {
	var response messaging.NotifierResponse

	isSms := false
	isEmail := false

	if EmailTemplateId != "" {
		isEmail = true
	}

	if SmsTemplateId != "" {
		isSms = true
	}

	var JSON_Document string
	var SMS_JSON_Document string
	var EMAIL_JSON_Document string

	domain := GetTenantNameFromEmail(recieverEmail)

	if isEmail {
		subject := GetEmailSubject(EmailTemplateId, domain)
		if subject != "NIL" {
			EMAIL_JSON_Document = getEmailJsonDoc(subject, domain, EmailTemplateId, defaultParams, customParams, recieverEmail)
		} else {
			isEmail = false
		}
	}

	if isSms {
		number := GetPhoneNumber(recieverEmail, domain)
		if number != "" {
			SMS_JSON_Document = getSMSJsonDoc(domain, SmsTemplateId, defaultParams, customParams, number)
		}
	}

	if isSms && isEmail {
		//both
		JSON_Document = Get_SMS_EMAIL_JSON_Document_(EMAIL_JSON_Document, SMS_JSON_Document)
	} else if isSms && !isEmail {
		//sms only
		JSON_Document = SMS_JSON_Document
	} else if !isSms && isEmail {
		//email only
		JSON_Document = EMAIL_JSON_Document
	} else if !isSms && !isEmail {
		JSON_Document = ""
		response.IsSuccess = false
		response.Message = "Notification sending Failed!"
		return response
	}

	url := "http://" + gethost() + ":3500/command/notification"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(JSON_Document)))
	req.Header.Set("securityToken", securityToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		response.IsSuccess = false
		response.Message = "Notification sending Failed!"
	} else {
		response.IsSuccess = true
		response.Message = "Notification sending Successful!"
	}

	return response
}

func Send(securityToken, subject, domain, class, templateId string, defaultParams map[string]string, customParams map[string]string, recieverEmail string) messaging.NotifierResponse {
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

func SendSMS(securityToken, domain, class, templateId string, defaultParams map[string]string, customParams map[string]string, recieverNumber string) messaging.NotifierResponse {
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

func getFrom() (url string) {
	content, err := ioutil.ReadFile("settings.config")
	if err == nil {
		object := make(map[string]interface{})
		_ = json.Unmarshal(content, &object)
		url = object["From"].(string)
	} else {
		url = "DuoWorld.com <mail-noreply@duoworld.com>"
	}
	return
}

func getEmailJsonDoc(subject, domain, templateId string, defaultParams map[string]string, customParams map[string]string, recieverEmail string) (json string) {
	json = "{\"type\":\"email\",\"to\":\"" + recieverEmail + "\",\"subject\":\"" + subject + "\",\"from\":\"" + getFrom() + "\",\"Namespace\": \"" + domain + "\",\"TemplateID\": \"" + templateId + "\",\"DefaultParams\": {" + getStringByMap(defaultParams) + "},\"CustomParams\": {" + getStringByMap(customParams) + "}}"
	return
}

func getSMSJsonDoc(domain, templateId string, defaultParams map[string]string, customParams map[string]string, recieverNumber string) (json string) {
	json = "{\"type\":\"sms\",\"number\":\"" + recieverNumber + "\",\"Namespace\": \"" + domain + "\",\"TemplateID\": \"" + templateId + "\",\"DefaultParams\": {" + getStringByMap(defaultParams) + "},\"CustomParams\": {" + getStringByMap(customParams) + "}}"
	return
}

func Get_SMS_EMAIL_JSON_Document_(emailDoc, smsDoc string) (json string) {
	json = "{\"type\": \"email,sms\"," + emailDoc + "," + smsDoc + "}"
	fmt.Println(json)
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

func GetTenantNameFromEmail(email string) (tenantName string) {
	tenantName = strings.Replace(email, "@", "", -1)
	tenantName = strings.Replace(email, ".", "", -1)
	return
}

func GetEmailSubject(EmailTemplateId, tenant string) (subject string) {
	bytes, _ := client.Go("securityToken", tenant, "templates").GetOne().ByUniqueKey(EmailTemplateId).Ok()
	if bytes == nil || len(bytes) <= 4 {
		bytes, _ = client.Go("securityToken", "com.duosoftware.com", "templates").GetOne().ByUniqueKey(EmailTemplateId).Ok()
	}

	if len(bytes) <= 4 {
		subject = "NIL"
	} else {
		data := make(map[string]interface{})
		err := json.Unmarshal(bytes, &bytes)
		if err != nil {
			fmt.Println(err.Error())
			subject = "NIL"
		} else {
			fmt.Println(data)
			subject = data["Title"].(string)
		}
	}

	return
}

func GetPhoneNumber(recieverEmail, tenant string) (phone string) {
	bytes, _ := client.Go("securityToken", tenant, "profile").GetOne().BySearching("Email:yoyo@yoyo.com").Ok()
	if bytes == nil || len(bytes) <= 4 {
		fmt.Println("No Phone Number Found!")
		phone = ""
	} else {
		fmt.Println("Record Found in Profile....")
		var data []map[string]interface{}
		err := json.Unmarshal(bytes, &data)
		if err != nil {
			fmt.Println(err.Error())
			phone = ""
		} else {
			fmt.Println(data)
			phone = data[0]["PhoneNumber"].(string)
		}
	}
	return
}
