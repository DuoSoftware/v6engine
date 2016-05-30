package main

import (
	"duov6.com/cebadapter"
	"duov6.com/common"
	"duov6.com/objectstore/client"
	"fmt"
)

func main() {
	initialize()
	send()
}

func initialize() {
	forever := make(chan bool)
	cebadapter.Attach("ObjectStore", func(s bool) {
		cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
			fmt.Println("Store Configuration Successfully Loaded...")
			fmt.Println(data)
			agent := cebadapter.GetAgent()
			forever <- false
			agent.Client.OnEvent("globalConfigChanged.StoreConfig", func(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
				cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
					fmt.Println("Store Configuration Successfully Updated...")
				})
			})
		})
		fmt.Println("Successfully registered in CEB")
	})

	<-forever
}

func send() {

	var Activ ActivationEmail
	Activ.GUUserID = "Gg@gmail.com"
	Activ.Token = common.RandText(10)

	client.Go("ignore", "com.duosoftware.auth", "activation").StoreObject().WithKeyField("Token").AndStoreOne(Activ).Ok()

}

type ActivationEmail struct {
	GUUserID string // GUUserID
	Token    string // Token for the email actiavte form
}
