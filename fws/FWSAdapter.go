package fws

import (
	"duov6.com/common"
	"duov6.com/config"
	//"duov6.com/term"
	"fmt"
	"strings"
)

func Attach(serverClass string) {

	agentConfig, err := config.GetMap("agent")

	if err == nil {
		client, err := NewFWSClient(agentConfig["cebUrl"].(string))

		if err == nil {

			if client == nil {
				fmt.Println("CLIENT IS NIL")
			}

			client.Resources["client"] = client

			client.Subscribe("command", "switch", Switch)
			client.Subscribe("command", "globalconfigrecieved", GlobalConfigRecieved)
			client.Subscribe("command", "getinfo", GetInfo)

			registerConfig(client)

			client.Register(serverClass+"@"+common.GetLocalHostName(), "1234")

			//forever := make(chan bool)
			//<-forever
		} else {
			fmt.Println("TCP Connection Error : " + err.Error())
		}
	} else {
		fmt.Println("Error retrieving agent configuration : " + err.Error())
	}

}

func registerConfig(client *FWSClient) {
	configs := config.GetConfigs()

	for _, c := range configs {
		nc := strings.Replace(c, ".config", "", -1)
		m, err := config.GetMap(nc)
		if err == nil {
			md := ConfigMetadata{}
			md.Name = nc
			md.Code = nc
			md.Parameters = m
			client.AddConfigMetadata(md)
		} else {
			fmt.Println(err.Error())
		}

	}
}
