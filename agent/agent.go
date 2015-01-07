package main

import (
	"duov6.com/agent/commands"
	"duov6.com/agent/core"
	"duov6.com/fws"
	"fmt"
)

func main() {
	a := &core.Agent{}

	c, err := fws.NewFWSClient("192.168.2.42:5000")

	if err == nil {

		c.Resources["agent"] = a
		a.Client = c

		if c == nil {
			fmt.Println("CLIENT IS NIL")
		}

		a.Client.Subscribe("command", "switch", commands.AgentSwitch)

		//a.Client.Subscribe("command", "monitorLog", commands.MonitorLog)

		a.Client.Register("Agent 2", "1234")

		forever := make(chan bool)
		<-forever
	} else {
		fmt.Println("TCP Connection Error : " + err.Error())
	}

}
