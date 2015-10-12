package main

import (
	"duov6.com/agentCore"
	"duov6.com/agentCore/commands"
	"fmt"
)

func main() {

	err :=  agentCore.New("", func(s bool){
		fmt.Println("Successfully Registered Agent!!!!");
		agentCore.GetInstance().Client.OnEvent ("userstatechanged",commands.GoOffline)
	});
	
	if err ==nil{
		forever := make(chan bool)
		<-forever		
	}

}
