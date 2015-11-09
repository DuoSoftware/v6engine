package cebadapter
//I don't  want to be a lazy coder :`(
//Finally I am not lazy :)
import (
	"duov6.com/agentCore"
	"duov6.com/agentCore/core"
	"duov6.com/agentCore/commands"
	"fmt"
)

var agent *core.Agent

func Attach(serverClass string, callback func(s bool)) {

	
	err := agentCore.New(serverClass, func(s bool){
		if (s== true){
			agentCore.GetInstance().Client.OnEvent ("userstatechanged",commands.GoOffline)
		}
		callback(s);
	});

	if err == nil{
		agentCore.GetInstance().Client.OnCommand("globalconfigrecieved", GlobalConfigRecieved)
	}else{
		fmt.Println("Error Creating Client!!!");
	}

}

func GetAgent() (agent *core.Agent){
	return agentCore.GetInstance()
}
