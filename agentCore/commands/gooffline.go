package commands

import (
	"duov6.com/agentCore/core"
	"fmt"
)

func GoOffline(from string, name string, data map[string]interface{}, resources map[string]interface{}) {

	if (data["state"] == "offline"){
		agent := resources["agent"].(*core.Agent)
		if (agent.Client.ListenerName == from && agent.IsAgentEnabled){
			fmt.Println ("Tenantwatch user went offline!!!!!")
			agent.IsAgentEnabled = false;
			agent.Client.ListenerName = "";
		}
	}
}
