package fws

import (
	"duov6.com/common"
	"duov6.com/config"

	//"duov6.com/term"
	"fmt"
)

func Attach(serverClass string) {

	agentConfig, err := config.GetMap("agent")

	if err == nil {
		client, err := NewFWSClient(agentConfig["cebUrl"].(string))

		if err == nil {

			if client == nil {
				fmt.Println("CLIENT IS NIL")
			}

			client.Subscribe("command", "switch", Switch)
			client.Subscribe("command", "globalconfigrecieved", GlobalConfigRecieved)

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
