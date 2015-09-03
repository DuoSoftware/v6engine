package commands

import (
	//"bitbucket.org/bertimus9/systemstat"
	"duov6.com/agentCore/core"
	"fmt"
	"time"
	"math/rand"
)

var agent *core.Agent

func AgentSwitch(from string, name string, data map[string]interface{}, resources map[string]interface{}) {

	fmt.Println(data)

	agent = resources["agent"].(*core.Agent)
	var attrib = data["state"].(string)
	isStat := false

	if data["enableStats"] != nil {
		isStat = data["enableStats"].(bool)
	}

	if attrib == "on" {
		agent.Client.ListenerName = from
		fmt.Println("Turning on Monitoring : " + from)
		if (len(agent.Client.StatMetadata) >0 ){
			agent.IsAgentEnabled = isStat
			if isStat == true{
				go StartTimer()	
			}			
		}else{
			fmt.Println("No stats defined for this agent")
		}

		
	} else if attrib == "off" {
		fmt.Println("Turning off Monitoring : " + from)
		agent.IsAgentEnabled = false
		agent.Client.ListenerName = "";
	}
}

func StartTimer() {
	c := time.Tick(2 * time.Second)
	for now := range c {
		_ = now
		
		if agent.IsAgentEnabled {
			matrix := Matrices{}

			/*
				matrix.TotalMemory = systemstat.GetMemSample().MemTotal
				matrix.BufferSize = systemstat.GetMemSample().Buffers
				matrix.UsedMemory = systemstat.GetMemSample().MemUsed
				matrix.TotalSwapMemory = systemstat.GetMemSample().SwapTotal
				matrix.UsedSwapMemory = systemstat.GetMemSample().SwapUsed
				matrix.FeeSwapMemory = systemstat.GetMemSample().SwapFree
				matrix.Freememory = systemstat.GetMemSample().MemFree
				matrix.SystemTime = systemstat.GetUptime().Time
				matrix.SystemupTime = systemstat.GetUptime().Uptime
			*/

			matrix.TotalMemory = rand.Intn(20)        //systemstat.GetMemSample().MemTotal
			matrix.BufferSize = rand.Intn(20)         //systemstat.GetMemSample().Buffers
			matrix.UsedMemory = rand.Intn(20)         //systemstat.GetMemSample().MemUsed
			matrix.TotalSwapMemory = rand.Intn(20)    //systemstat.GetMemSample().SwapTotal
			matrix.UsedSwapMemory = rand.Intn(20)     //systemstat.GetMemSample().SwapUsed
			matrix.FeeSwapMemory = rand.Intn(20)      //systemstat.GetMemSample().SwapFree
			matrix.Freememory = rand.Intn(20)         //systemstat.GetMemSample().MemFree
			matrix.SystemTime = time.Now() //systemstat.GetUptime().Time
			matrix.SystemupTime = 10       //systemstat.GetUptime().Uptime1

			agent.Client.ClientCommand(agent.Client.ListenerName, "stat", "test", matrix)
		}else{
			break
		}

	}
}
