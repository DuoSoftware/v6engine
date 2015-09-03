package duonotifier

import (
	"duov6.com/cebadapter"
	"duov6.com/duonotifier/configuration"
	"duov6.com/duonotifier/messaging"
	"duov6.com/duonotifier/repositories"
	"strings"
)

func Send(securityToken string, notifyMethod string, parameters map[string]interface{}) messaging.NotifierResponse {
	//Creating Request to send
	var Request *messaging.NotifierRequest
	//Loading CEB
	forever := make(chan bool)
	cebadapter.Attach("DuoNotifier", func(s bool) {
		cebadapter.GetLatestGlobalConfig("DuoNotifier", func(data []interface{}) {
			Request.Log("DuoNotifier Configuration Successfully Loaded...")
			if data != nil {
				forever <- false
			}
		})
		Request.Log("Successfully registered in CEB")
	})

	<-forever

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
	namespace := tenentData[0] + "." + tenentData[1] + "." + tenentData[2]
	class := tenentData[3]

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
	response = execute(Request)
	return response

}

func execute(Request *messaging.NotifierRequest) messaging.NotifierResponse {
	var response messaging.NotifierResponse
	abstractRepository := repositories.Create(Request.NotifyMethod)
	Request.Log("Executing Abstract Repository : " + abstractRepository.GetNotifierName())
	response = abstractRepository.ExecuteNotifier(Request)
	return response
}
