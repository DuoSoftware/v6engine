package commands

import (
	"duov6.com/agent/core"
	"fmt"
	"time"
)

var agent *core.Agent

func AgentSwitch(from string, name string, data map[string]interface{}, resources map[string]interface{}) {

	fmt.Println(data)

	agent = resources["agent"].(*core.Agent)
	var attrib = data["state"].(string)

	if attrib == "on" {
		agent.ListnerName = from
		fmt.Println("Turning on Monitoring : " + from)
		agent.IsAgentEnabled = true
		go StartTimer()
	} else if attrib == "off" {
		fmt.Println("Turning off Monitoring" + from)
		agent.IsAgentEnabled = false
	}
}

func StartTimer() {

	c := time.Tick(1 * time.Second)
	for now := range c {
		_ = now

		if agent.IsAgentEnabled {
			matrix := Matrices{}
			matrix.Ok = "true"

			agent.Client.ClientCommand(agent.ListnerName, "matrics", "test", matrix)
		}

	}
}
