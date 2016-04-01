package main

import (
	"duov6.com/cebadapter"
	"duov6.com/serviceconsole/scheduler/core"
	"duov6.com/term"
	"fmt"
)

type Scheduler struct {
}

func (s *Scheduler) Start() {
	// cebadapter.Attach("ProcessScheduler", func(s bool) {
	// 	cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
	// 		fmt.Println("Store Configuration Successfully Loaded...")

	// 		agent := cebadapter.GetAgent()

	// 		agent.Client.OnEvent("globalConfigChanged.StoreConfig", func(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
	// 			cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
	// 				fmt.Println(data)
	// 				fmt.Println("Store Configuration Successfully Updated...")
	// 			})
	// 		})
	// 	})
	// 	fmt.Println("Successfully registered in CEB")
	// })
	forever := make(chan bool)
	cebadapter.Attach("ObjectStore", func(s bool) {
		cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
			fmt.Println("Store Configuration Successfully Loaded...")
			fmt.Println(data)
			if data != nil {
				forever <- false
			}
			agent := cebadapter.GetAgent()

			agent.Client.OnEvent("globalConfigChanged.StoreConfig", func(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
				cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
					fmt.Println("Store Configuration Successfully Updated...")
				})
			})
		})
		fmt.Println("Successfully registered in CEB")
	})
	<-forever

	downloader := core.Downloader{}
	term.Write("Starting Serviec Console Scheduler...", term.Debug)
	downloader.Start()

}

func main() {
	scheduler := Scheduler{}
	scheduler.Start()
}
