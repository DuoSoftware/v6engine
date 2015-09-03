package commands

import (
	"duov6.com/agentCore/core"
	"fmt"
)

func ServerMonitor(from string, name string, data map[string]interface{}, resources map[string]interface{}) {

	fmt.Println(data)

	agent = resources["agent"].(*core.Agent)

	var attrib = data["state"].(string)

	if attrib == "on" {
		agent.Client.ListenerName = from
		fmt.Println("Turning on Monitoring : " + from)
		agent.IsAgentEnabled = true
	} else if attrib == "off" {
		fmt.Println("Turning off Monitoring : " + from)
		agent.IsAgentEnabled = false
	}
}
