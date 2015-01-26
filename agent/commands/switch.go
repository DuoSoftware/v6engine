package commands

import (
	//"bitbucket.org/bertimus9/systemstat"
	"duov6.com/agent/core"
	"fmt"
	"time"
)

var agent *core.Agent

func AgentSwitch(from string, name string, data map[string]interface{}, resources map[string]interface{}) {

	fmt.Println(data)

	agent = resources["agent"].(*core.Agent)
	var attrib = data["state"].(string)
	var isStat = data["enableStats"].(bool)

	if attrib == "on" {
		agent.ListnerName = from
		fmt.Println("Turning on Monitoring : " + from)
		agent.IsAgentEnabled = isStat
		go StartTimer()
	} else if attrib == "off" {
		fmt.Println("Turning off Monitoring" + from)
		agent.IsAgentEnabled = false
	}
}

func StartTimer() {

	c := time.Tick(1 * time.Second)
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

			matrix.TotalMemory = 10        //systemstat.GetMemSample().MemTotal
			matrix.BufferSize = 10         //systemstat.GetMemSample().Buffers
			matrix.UsedMemory = 10         //systemstat.GetMemSample().MemUsed
			matrix.TotalSwapMemory = 10    //systemstat.GetMemSample().SwapTotal
			matrix.UsedSwapMemory = 10     //systemstat.GetMemSample().SwapUsed
			matrix.FeeSwapMemory = 10      //systemstat.GetMemSample().SwapFree
			matrix.Freememory = 10         //systemstat.GetMemSample().MemFree
			matrix.SystemTime = time.Now() //systemstat.GetUptime().Time
			matrix.SystemupTime = 10       //systemstat.GetUptime().Uptime1

			agent.Client.ClientCommand(agent.ListnerName, "stat", "test", matrix)
		}

	}
}
