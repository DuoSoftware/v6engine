package commands

import (
	"duov6.com/fws"
	"fmt"
)

func GetInfo(from string, name string, data map[string]interface{}, resources map[string]interface{}) {

	fmt.Println(data)

	client := resources["client"].(*fws.FWSClient)

	agentInfo := AgentInfo{}
	agentInfo.CommandMaps = client.CommandMaps
	agentInfo.StatMetadata = client.StatMetadata
	agentInfo.ConfigMetadata = client.ConfigMetadata

	if agent.ListnerName != "" {
		agent.Client.ClientCommand(agent.ListnerName, "agent", "test", agentInfo)
	}

}

type AgentInfo struct {
	CommandMaps    []fws.CommandMap     `json:"commandMaps"`
	StatMetadata   []fws.StatMetadata   `json:"statMetadata"`
	ConfigMetadata []fws.ConfigMetadata `json:"configMetadata"`
}
