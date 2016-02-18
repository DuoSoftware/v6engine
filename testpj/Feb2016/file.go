package main

import (
	"duov6.com/cebadapter"
	"duov6.com/objectstore/client"
	"duov6.com/term"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func main() {
	initialize()
	bytes, _ := client.Go("ignore", "roshitha123duocom.space.test.12thdoor.com", "productimagesNew").GetOne().ByUniqueKey("bamil bulk.jpg").Ok() // fetech user autherized

	dd := record{}
	json.Unmarshal(bytes, &dd)

	ioutil.WriteFile("ddds.jpg", dd.Body, 0666)
	fmt.Println("huehuehue")
}

func initialize() {
	cebadapter.Attach("DuoAuth", func(s bool) {
		cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
			term.Write("Store Configuration Successfully Loaded...", term.Information)

			agent := cebadapter.GetAgent()

			agent.Client.OnEvent("globalConfigChanged.StoreConfig", func(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
				cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
					term.Write("Store Configuration Successfully Updated...", term.Information)
				})
			})
		})
		term.Write("Successfully registered in CEB", term.Information)
	})

}

type record struct {
	Body     []byte
	FileName string
	Id       string
}
