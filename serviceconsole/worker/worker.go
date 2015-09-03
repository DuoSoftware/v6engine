package main

import (
	"duov6.com/serviceconsole/scheduler/core"
	"fmt"
)

type Worker struct {
}

func (w *Worker) Start() {
	cebadapter.Attach("ProcessWorker", func(s bool){
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

	downloader := core.Downloader{}
	fmt.Println("worker start ")
	downloader.Start()
}

func main() {
	worker := Worker{}
	worker.Start()
}
