package main

import (
	"duov6.com/cebadapter"
	"duov6.com/serviceconsole/configuration"
	"duov6.com/serviceconsole/messaging"
	"duov6.com/serviceconsole/processmanager"
	"fmt"
)

type serviceconsole struct {
	Request *messaging.ServiceRequest
}

func (service *serviceconsole) Begin() {

	var tempConf = configuration.ConfigurationManager{}.Get()
	var storedServiceConfiguration = configuration.StoreServiceConfiguration{}
	storedServiceConfiguration = tempConf

	fmt.Println(storedServiceConfiguration.ServerConfiguration)

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
	Draw()

	cebadapter.Attach("ServiceConsole", func(s bool){
		cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
			fmt.Println("Store Configuration Successfully Loaded...")

			agent := cebadapter.GetAgent();
			
			agent.Client.OnEvent("globalConfigChanged.StoreConfig", func(from string, name string, data map[string]interface{}, resources map[string]interface{}){
				cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
					fmt.Println("Store Configuration Successfully Updated...")
				});
			});
		})
		fmt.Println("Successfully registered in CEB")
	});

	x := make(chan bool)
	x <- true

}

func Draw() {

	fmt.Println("")
	fmt.Println("")
	fmt.Println("______             _____                 _          ")
	fmt.Println("|  _  \\           /  ___|               (_)         ")
	fmt.Println("| | | |_   _  ___ \\ `--.  ___ _ ____   ___  ___ ___ ")
	fmt.Println("| | | | | | |/ _ \\ `--. \\/ _ \\ '__\\ \\ / / |/ __/ _ \\")
	fmt.Println("| |/ /| |_| | (_) /\\__/ /  __/ |   \\ V /| | (_|  __/")
	fmt.Println("|___/  \\__,_|\\___/\\____/ \\___|_|    \\_/ |_|\\___\\___|")
	fmt.Println("")
	fmt.Println("")
}

/*
package main

import (
	"duov6.com/fws"
	"duov6.com/serviceconsole/configuration"
	"duov6.com/serviceconsole/messaging"
	"duov6.com/serviceconsole/processmanager"
	"fmt"
)

type serviceconsole struct {
	Request *messaging.ServiceRequest
}

func (service *serviceconsole) Begin() {

	Draw()
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
	service.Request.Log("Prasad")
}

func main() {

	fws.Attach("ServiceConsole")
	var service = serviceconsole{}
	service.Begin()
}

func Draw() {

	fmt.Println("")
	fmt.Println("")
	fmt.Println("______             _____                 _          ")
	fmt.Println("|  _  \\           /  ___|               (_)         ")
	fmt.Println("| | | |_   _  ___ \\ `--.  ___ _ ____   ___  ___ ___ ")
	fmt.Println("| | | | | | |/ _ \\ `--. \\/ _ \\ '__\\ \\ / / |/ __/ _ \\")
	fmt.Println("| |/ /| |_| | (_) /\\__/ /  __/ |   \\ V /| | (_|  __/")
	fmt.Println("|___/  \\__,_|\\___/\\____/ \\___|_|    \\_/ |_|\\___\\___|")
	fmt.Println("")
	fmt.Println("")
}


*/
