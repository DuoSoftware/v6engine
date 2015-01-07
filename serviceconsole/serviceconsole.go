package main

import (
	"duov6.com/serviceconsole/configuration"
	"duov6.com/serviceconsole/messaging"
	"duov6.com/serviceconsole/processmanager"
	"fmt"
	//"reflect"
	//"strconv"
)

type serviceconsole struct {
	Request *messaging.ServiceRequest
}

func (service *serviceconsole) Begin() {

	var tempConf = configuration.MockServiceConfigurationDownloader{}
	var storedServiceConfiguration = configuration.StoreServiceConfiguration{}
	storedServiceConfiguration = tempConf.DownloadConfiguration()
	service.Request = &messaging.ServiceRequest{}
	service.Request.OperationCode = "ExecuteWorker"
	service.Request.Configuration = storedServiceConfiguration
	exe := processmanager.WorkersExecutor{}
	response := exe.Execute(service.Request)
	if response.IsSuccess == true {
		fmt.Println("Successfully Executed!")
	} else {
		fmt.Println("Execution Failed!")
	}
}

func main() {
	var service = serviceconsole{}
	service.Begin()
}
