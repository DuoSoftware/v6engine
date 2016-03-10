package main

import (
	"duov6.com/cebadapter"
	"duov6.com/objectstore/client"
	"fmt"
	"strings"
)

func initialize() {
	forever := make(chan bool)
	cebadapter.Attach("ObjectStore", func(s bool) {
		cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
			fmt.Println("Store Configuration Successfully Loaded...")
			fmt.Println(data)
			if data != nil {
				forever <- false
				return
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
}

func main() {
	initialize()

	tmp := AuthCertificate{}
	tmp.Otherdata = make(map[string]string)
	tmp.UserID = "1"
	tmp.Username = "asdf"
	tmp.Name = "ggwp"
	tmp.Email = "gg@duo.com"
	tmp.SecurityToken = "-999"
	tmp.Domain = "duo"
	tmp.DataCaps = "none"
	tmp.ClientIP = "localhost"
	tmp.Otherdata["JWT"] = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkbW4iOiJnbWFpbC5jb20iLCJlbWwiOiJwcmFzYWRAZ21haWwuY29tIiwiaXNzIjoiZ21haWwuY29tIiwic2NvcGUiOnt9LCJzdCI6ImJiNGFiNjFhODI3NmI3YTRhZTE1ZTYyOGEyODFiNmM5IiwidWlkIjoiNWE4ZjZlYmI1Y2JjODU1NmE0Zjg4NTQ2ZDY5OTU3YmEifQ==.Tdnwv8Du9POvRYLfeOmyJ2ZiAmwosFTS7eh5svcrjiY="
	tmp.Otherdata["Scope"] = ""
	tmp.Otherdata["TenentsAccessible"] = strings.Replace("[{\"TenantID\":\"hh.bk.ebankslk.com\",\"Name\":\"hhh\"}]", "\"", "`", -1)
	//tmp.Otherdata["TenentsAccessible"] = `[{"TenantID":"hh.bk.ebankslk.com","Name":"hhh"}]`
	//tmp.Otherdata["TenentsAccessible"] = "jay"
	tmp.Otherdata["UserAgent"] = "Mozilla/5.0 (Windows NT 6.3; WOW64; rv:44.0) Gecko/20100101 Firefox/44.0"

	client.Go("token", "com.jay.test", "lol72").StoreObject().WithKeyField("SecurityToken").AndStoreOne(tmp).Ok()

}

type AuthCertificate struct {
	UserID, Username, Name, Email, SecurityToken, Domain, DataCaps, ClientIP string
	Otherdata                                                                map[string]string
}
