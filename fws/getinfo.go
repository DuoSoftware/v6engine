package fws

import (
	"fmt"
)

func GetInfo(from string, name string, data map[string]interface{}, resources map[string]interface{}) {

	fmt.Println(data)

	client := resources["client"].(*FWSClient)

	agentInfo := AgentInfo{}
	agentInfo.CommandMaps = client.CommandMaps
	agentInfo.StatMetadata = client.StatMetadata
	agentInfo.ConfigMetadata = client.ConfigMetadata

	if client.ListenerName != "" {
		client.ClientCommand(client.ListenerName, "agent", "test", agentInfo)
	}

}

type AgentInfo struct {
	CommandMaps    []CommandMap     `json:"commandMaps"`
	StatMetadata   []StatMetadata   `json:"statMetadata"`
	ConfigMetadata []ConfigMetadata `json:"configMetadata"`
}
