package main

import (
	"duov6.com/cebadapter"
	"duov6.com/duonotifier/configuration"
	"duov6.com/duonotifier/endpoints"
	"duov6.com/duonotifier/messaging"
	"duov6.com/duonotifier/repositories"
	"fmt"
	"strings"
)

func main() {
	initializeCEBConfig()
	httpServer := endpoints.HTTPService{}
	go httpServer.Start()
}

func Send(securityToken string, notifyMethod string, parameters map[string]interface{}) messaging.NotifierResponse {
	initializeCEBConfig()
	//Creating Request to send
	var Request *messaging.NotifierRequest
	//Load configurations from CEB to request
	var response messaging.NotifierResponse
	response = initialize(Request, securityToken, notifyMethod, parameters)
	return response
}

func initializeCEBConfig() {
	inititalizeObjectStoreConfig()
	initializeDuoNotifierConfig()
}

func initializeDuoNotifierConfig() {
	forever := make(chan bool)
	cebadapter.Attach("DuoNotifier", func(s bool) {
		cebadapter.GetLatestGlobalConfig("DuoNotifier", func(data []interface{}) {
			fmt.Println("DuoNotifier Configuration Successfully Loaded...")
			fmt.Println(data)
			if data != nil {
				forever <- false
			}
		})
		fmt.Println("Successfully registered DuoNotifier in CEB")
	})

	<-forever
}

func inititalizeObjectStoreConfig() {
	forever := make(chan bool)
	cebadapter.Attach("ObjectStore", func(s bool) {
		cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
			fmt.Println("Store Configuration Successfully Loaded...")
			fmt.Println(data)
			agent := cebadapter.GetAgent()
			agent.Client.OnEvent("globalConfigChanged.StoreConfig", func(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
				cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
					fmt.Println("Store Configuration Successfully Updated...")
				})
			})
		})
		fmt.Println("Successfully registered ObjectStore in CEB")
	})

	<-forever
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
