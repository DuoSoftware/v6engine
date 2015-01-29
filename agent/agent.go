package main

import (
	"duov6.com/agent/commands"
	"duov6.com/agent/core"
	"duov6.com/agent/dcommands"
	"duov6.com/common"
	"duov6.com/config"
	"duov6.com/fws"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {

	data, err := ioutil.ReadFile("./agent.config")

	if err == nil {
		var settings map[string]interface{}
		settings = make(map[string]interface{})
		err := json.Unmarshal(data, &settings)

		if err == nil {

			a := &core.Agent{}

			c, err := fws.NewFWSClient(settings["cebUrl"].(string))
			if err == nil {
				c.Resources["agent"] = a
				a.Client = c

				if c == nil {
					fmt.Println("CLIENT IS NIL")
				}

				a.Client.Subscribe("command", "switch", commands.AgentSwitch)
				a.Client.Subscribe("command", "getinfo", commands.GetInfo)

				registerCommands(a)
				registerConfig(a)
				registerStats(a)

				agentClass := settings["class"].(string)

				a.Client.Register(agentClass+"@"+common.GetLocalHostName(), "1234")

				forever := make(chan bool)
				<-forever
			} else {
				fmt.Println("TCP Connection Error : " + err.Error())
			}

		} else {
			fmt.Println("Configuration file is not in correct JSON format : " + err.Error())
		}

	} else {
		fmt.Println("Error accessing configuration file ./agent.config : " + err.Error())
	}

}

func registerCommands(a *core.Agent) {
	maps, err := dcommands.GetAllMaps()
	if err == nil {

		for _, m := range maps {

			a.Client.AddCommandMetadata(m)
			a.Client.Subscribe("command", m.Code, commands.ShellScript)
		}
	} else {
		fmt.Println("Error reading custom scripts.")
	}

}

func registerConfig(a *core.Agent) {
	configs := config.GetConfigs()

	for _, c := range configs {
		nc := strings.Replace(c, ".config", "", -1)
		m, err := config.GetMap(nc)
		if err == nil {
			md := fws.ConfigMetadata{}
			md.Name = nc
			md.Code = nc
			md.Parameters = m
			a.Client.AddConfigMetadata(md)
		} else {
			fmt.Println(err.Error())
		}

	}
}

func registerStats(a *core.Agent) {
	m_TotalMemory := fws.StatMetadata{}
	m_TotalMemory.Name = "Total Memory"
	m_TotalMemory.Type = "line"
	m_TotalMemory.XAxis = "TotalMemory"
	m_TotalMemory.YAxis = "SystemTime"
	m_TotalMemory.MaxX = 15
	a.Client.AddStatMetadata(m_TotalMemory)

	m_UsedMemory := fws.StatMetadata{}
	m_UsedMemory.Name = "Total Memory"
	m_UsedMemory.Type = "line"
	m_UsedMemory.XAxis = "UsedMemory"
	m_UsedMemory.YAxis = "SystemTime"
	m_UsedMemory.MaxX = 15
	a.Client.AddStatMetadata(m_TotalMemory)

	m_Freememory := fws.StatMetadata{}
	m_Freememory.Name = "Total Memory"
	m_Freememory.Type = "line"
	m_Freememory.XAxis = "Freememory"
	m_Freememory.YAxis = "SystemTime"
	m_Freememory.MaxX = 15
	a.Client.AddStatMetadata(m_Freememory)

	m_BufferSize := fws.StatMetadata{}
	m_BufferSize.Name = "Total Memory"
	m_BufferSize.Type = "line"
	m_BufferSize.XAxis = "BufferSize"
	m_BufferSize.YAxis = "SystemTime"
	m_BufferSize.MaxX = 15
	a.Client.AddStatMetadata(m_BufferSize)

	m_TotalSwapMemory := fws.StatMetadata{}
	m_TotalSwapMemory.Name = "Total Memory"
	m_TotalSwapMemory.Type = "line"
	m_TotalSwapMemory.XAxis = "TotalSwapMemory"
	m_TotalSwapMemory.YAxis = "SystemTime"
	m_TotalSwapMemory.MaxX = 15
	a.Client.AddStatMetadata(m_TotalSwapMemory)

	m_UsedSwapMemory := fws.StatMetadata{}
	m_UsedSwapMemory.Name = "Total Memory"
	m_UsedSwapMemory.Type = "line"
	m_UsedSwapMemory.XAxis = "UsedSwapMemory"
	m_UsedSwapMemory.YAxis = "SystemTime"
	m_UsedSwapMemory.MaxX = 15
	a.Client.AddStatMetadata(m_UsedSwapMemory)

	m_FeeSwapMemory := fws.StatMetadata{}
	m_FeeSwapMemory.Name = "Total Memory"
	m_FeeSwapMemory.Type = "line"
	m_FeeSwapMemory.XAxis = "FeeSwapMemory"
	m_FeeSwapMemory.YAxis = "SystemTime"
	m_FeeSwapMemory.MaxX = 15
	a.Client.AddStatMetadata(m_FeeSwapMemory)
}
