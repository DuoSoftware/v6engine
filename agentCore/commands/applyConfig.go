package commands

import (
	"duov6.com/config"
	"duov6.com/ceb"
	"fmt"
)

func ApplyConfig(from string, name string, data map[string]interface{}, resources map[string]interface{}) {

	var fileName = data["fileName"].(string)
	var jsonData = data["data"].(interface{})
	var mapData = data["data"].(map[string]interface{})

	fmt.Println("Config applied!!! ", fileName)

	client := resources["client"].(*ceb.CEBClient)
	client.UpdateCommandMetadata(fileName, mapData)
	config.Add(jsonData, fileName)

	//jsonBytes, _ := json.Marshal(jsonData)
	//jsonString := string(jsonBytes[:len(jsonBytes)])
	//config.Save(fileName, jsonString)

/*	client := resources["client"].(*fws.FWSClient)

	agentInfo := AgentInfo{}
	agentInfo.CommandMaps = client.CommandMaps
	agentInfo.StatMetadata = client.StatMetadata
	agentInfo.ConfigMetadata = client.ConfigMetadata

	if agent.ListnerName != "" {
		agent.Client.ClientCommand(agent.ListnerName, "agent", "test", agentInfo)
	}
*/

}