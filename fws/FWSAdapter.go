package fws

import (
	"duov6.com/common"
	//"duov6.com/term"
	"fmt"
)

func Attach(serverClass string) {

	client, err := NewFWSClient("192.168.2.42:5000")

	if err == nil {

		if client == nil {
			fmt.Println("CLIENT IS NIL")
		}

		client.Subscribe("command", "switch", Switch)

		client.Register(serverClass+"@"+common.GetLocalHostName(), "1234")

		//forever := make(chan bool)
		//<-forever
	} else {
		fmt.Println("TCP Connection Error : " + err.Error())
	}

}
