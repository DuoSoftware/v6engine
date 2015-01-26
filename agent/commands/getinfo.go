package commands

import (
	"duov6.com/agent/core"
	"duov6.com/fws"
	"fmt"
)

func GetInfo(from string, name string, data map[string]interface{}, resources map[string]interface{}) {

	fmt.Println(data)

	agent := resources["agent"].(*core.Agent)

	agentInfo := AgentInfo{}
	agentInfo.CommandMaps = agent.Client.CommandMaps
	agentInfo.StatMetadata = agent.Client.StatMetadata
	agentInfo.ConfigMetadata = agent.Client.ConfigMetadata

	agent.Client.ClientCommand(agent.ListnerName, "agent", "test", agentInfo)
}

type AgentInfo struct {
	CommandMaps    []fws.CommandMap     `json:"commandMaps"`
	StatMetadata   []fws.StatMetadata   `json:"statMetadata"`
	ConfigMetadata []fws.ConfigMetadata `json:"configMetadata"`
}
