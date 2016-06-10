package main

import (
	"duov6.com/cebadapter"
	//"duov6.com/common"
	client "duov6.com/objectstore/goclient"
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

	// var Activ ActivationEmail
	// Activ.GUUserID = "Gg@gmail.com"
	// Activ.Token = common.RandText(10)

	Activ := make(map[string]interface{})
	Activ["GUUserID"] = "1234"
	Activ["Token"] = "huehuehue"
	// gg := make([]ActivationEmail, 1)
	// gg[0] = Activ

	gg := make([]map[string]interface{}, 1)
	gg[0] = Activ

	b := make([]interface{}, len(gg))
	for i := range gg {
		b[i] = gg[i]
	}

	//client.Go("ignore", "com.duosoftware.auth", "activation").StoreObject().WithKeyField("Token").AndStoreOne(Activ).Ok()
	client.Go("ignore", "com.duosoftware.auth", "activation").DeleteObject().WithKeyField("Token").AndDeleteMany(b).Ok()
}

type ActivationEmail struct {
	GUUserID string // GUUserID
	Token    string // Token for the email actiavte form
}
