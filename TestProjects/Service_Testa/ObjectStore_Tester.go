package main

import (
	"duov6.com/cebadapter"
	//"duov6.com/common"
	client "duov6.com/objectstore/client"
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
	Activ["Token"] = "huehuehue2"

	Activ1 := make(map[string]interface{})
	Activ1["GUUserID"] = "1234"
	Activ1["Token"] = "huehuehue1"
	// gg := make([]ActivationEmail, 1)
	// gg[0] = Activ

	gg := make([]map[string]interface{}, 2)
	gg[0] = Activ
	gg[1] = Activ1

	b := make([]interface{}, len(gg))
	for i := range gg {
		b[i] = gg[i]
	}

	fmt.Println("999999")
	fmt.Println(b)
	//bytes, _ := client.Go("ignore", "roshitha123duocom.space.test.12thdoor.com", "productimagesNew").GetOne().ByUniqueKey("bamil bulk.jpg").Ok() // fetech user autherized

	client.Go("ignore", "wp", "gg").StoreObject().WithKeyField("Token").AndStoreMany(b).Ok()
	//client.Go("ignore", "wp", "gg").StoreObject().WithKeyField("Token").AndStoreOne(Activ).Ok()
	//client.Go("ignore", "com.duosoftware.auth", "activation").DeleteObject().WithKeyField("Token").AndDeleteMany(b).Ok()
}

type ActivationEmail struct {
	GUUserID string // GUUserID
	Token    string // Token for the email actiavte form
}
