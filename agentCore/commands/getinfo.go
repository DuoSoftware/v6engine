package commands

import (
	"duov6.com/ceb"
)

func GetInfo(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
	client := resources["client"].(*ceb.CEBClient)

	agentInfo := AgentInfo{}
	agentInfo.CommandMaps = client.CommandMaps
	agentInfo.StatMetadata = client.StatMetadata
	agentInfo.ConfigMetadata = client.ConfigMetadata
	agentInfo.CanMonitorOutput = client.CanMonitorOutput

	if agent.Client.ListenerName != "" {
		agent.Client.ClientCommand(agent.Client.ListenerName, "agent", "test", agentInfo)
	}

}

type AgentInfo struct {
	CommandMaps    		[]ceb.CommandMap     	`json:"commandMaps"`
	StatMetadata   		[]ceb.StatMetadata   	`json:"statMetadata"`
	ConfigMetadata 		[]ceb.ConfigMetadata 	`json:"configMetadata"`
	CanMonitorOutput 	bool					`json:"canMonitorOutput"`
}
