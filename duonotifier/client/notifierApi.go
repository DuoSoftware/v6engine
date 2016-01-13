package client

import (
	"duov6.com/duonotifier/configuration"
	"duov6.com/duonotifier/messaging"
	"duov6.com/duonotifier/repositories"
	"duov6.com/objectstore/client"
	"encoding/json"
	"fmt"
	"strings"
)

func Send(securityToken string, notifyMethod string, parameters map[string]interface{}) messaging.NotifierResponse {
	//Creating Request to send
	var Request *messaging.NotifierRequest
	//Load configurations from CEB to request
	var response messaging.NotifierResponse
	response = initialize(Request, securityToken, notifyMethod, parameters)
	return response
}

func initialize(Request *messaging.NotifierRequest, securityToken string, notifyMethod string, parameters map[string]interface{}) messaging.NotifierResponse {
	var notifyConfigs = configuration.ConfigurationManager{}.Get()
	var notifierConfiguration = configuration.NotifierConfiguration{}
	notifierConfiguration = notifyConfigs

	//read Namespace and Class to get information
	tenentData := strings.Split(notifierConfiguration.NotifyId, ".")
	namespace := tenentData[0]
	for x := 1; x < (len(tenentData) - 1); x++ {
		namespace += "." + tenentData[x]
	}
	class := tenentData[(len(tenentData) - 1)]

	var requestControls messaging.RequestNotifyControls
	requestControls.SecurityToken = securityToken
	requestControls.Namespace = namespace
	requestControls.Class = class

	Request = &messaging.NotifierRequest{}
	Request.NotifyMethod = notifyMethod
	Request.Parameters = parameters
	Request.Configuration = notifierConfiguration
	Request.Controls = requestControls

	//Execute the Repository
	var response messaging.NotifierResponse
	response = repositories.Execute(Request)
	return response

}

func GetTemplate(request messaging.TemplateRequest) (response interface{}) {
	//response = request
	data := make(map[string]interface{})
	isDefault := false
	bytes, _ := client.Go("securityToken", request.Namespace, "templates").GetOne().ByUniqueKey(request.TemplateID).Ok()
	if len(bytes) <= 4 {
		isDefault = true
		bytes, _ = client.Go("securityToken", "com.duosoftware.com", "templates").GetOne().ByUniqueKey(request.TemplateID).Ok()
	}
	_ = json.Unmarshal(bytes, &data)
	result := getFormatted(request, data, isDefault)
	response = result
	return
}

//template code... structure it as soon as possible.

func getFormatted(request messaging.TemplateRequest, data map[string]interface{}, isDefault bool) map[string]interface{} {

	keyWordMap := make(map[string]string)

	dataCopy := make(map[string]interface{})

	for key, value := range data {
		dataCopy[key] = value
	}

	if isDefault {
		for key, value := range request.DefaultParams {
			keyWordMap[key] = value
		}
	} else {
		for key, value := range request.CustomParams {
			keyWordMap[key] = value
		}
	}

	for key, value := range dataCopy {

		if key != "TemplateID" {
			stringVal := value.(string)
			fmt.Println(stringVal)
			for param, paramVal := range keyWordMap {
				stringVal = strings.Replace(stringVal, param, paramVal, -1)
			}
			dataCopy[key] = stringVal
		}
	}

	return dataCopy
}
