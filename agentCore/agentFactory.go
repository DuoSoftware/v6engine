package agentCore

import (
	"duov6.com/agentCore/commands"
	"duov6.com/agentCore/core"
	"duov6.com/agentCore/dcommands"
	"duov6.com/common"
	"duov6.com/config"
	"duov6.com/ceb"
	"duov6.com/term"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)


var agentInstance *core.Agent;
//initializes a new agent
func New(agentClass string, callback func(s bool))(err error){

	data, err := ioutil.ReadFile("./agent.config")

	if err == nil {
		var settings map[string]interface{}
		settings = make(map[string]interface{})
		err := json.Unmarshal(data, &settings)

		if err == nil {

			a := &core.Agent{}
			agentInstance = a;
			c, err := ceb.NewCEBClient(settings["cebUrl"].(string), func (s bool){
				if (a.Client.CanMonitorOutput){
					term.AddPlugin(AgentLogger{})
				}
				callback(s)	
			})
			if err == nil {
				c.Resources["agent"] = a
				c.Resources["client"] = c

				a.Client = c

				if c == nil {
					fmt.Println("CLIENT IS NIL")
				}

				a.Client.OnCommand("switch", commands.AgentSwitch)
				a.Client.OnCommand("getinfo", commands.GetInfo)
				a.Client.OnCommand("applyconfig", commands.ApplyConfig)
				

				registerCommands(a)
				registerConfig(a)

				if (settings["showResourceStats"]!=nil){
					if (settings["showResourceStats"].(bool) == true){
						registerStats(a)
					}
				}

				if (settings["canMonitorOutput"]!=nil){
					if (settings["canMonitorOutput"].(bool) == true){
						a.Client.CanMonitorOutput = true;
					}
				}

				if agentClass == ""{
					agentClass = settings["class"].(string)	
				}
				

				a.Client.Register(agentClass+"@"+common.GetLocalHostName(), "1234")

			} else {
				fmt.Println("TCP Connection Error : " + err.Error())
			}

		} else {
			fmt.Println("Configuration file is not in correct JSON format : " + err.Error())
		}

	} else {
		fmt.Println("Error accessing configuration file ./agent.config : " + err.Error())
	}
	
	return 
}

func GetInstance()(a *core.Agent){
	return agentInstance;
}

func registerCommands(a *core.Agent) {
	maps, err := dcommands.GetAllMaps()
	if err == nil {

		for _, m := range maps {

			a.Client.AddCommandMetadata(m)
			a.Client.OnCommand(m.Code, commands.ShellScript)
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
			md := ceb.ConfigMetadata{}
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
	m_TotalMemory := ceb.StatMetadata{}
	m_TotalMemory.Name = "Total Memory"
	m_TotalMemory.Type = "line"
	m_TotalMemory.XAxis = "TotalMemory"
	m_TotalMemory.YAxis = "SystemTime"
	m_TotalMemory.MaxX = 15
	a.Client.AddStatMetadata(m_TotalMemory)

	m_UsedMemory := ceb.StatMetadata{}
	m_UsedMemory.Name = "Total Memory"
	m_UsedMemory.Type = "line"
	m_UsedMemory.XAxis = "UsedMemory"
	m_UsedMemory.YAxis = "SystemTime"
	m_UsedMemory.MaxX = 15
	a.Client.AddStatMetadata(m_TotalMemory)

	m_Freememory := ceb.StatMetadata{}
	m_Freememory.Name = "Total Memory"
	m_Freememory.Type = "line"
	m_Freememory.XAxis = "Freememory"
	m_Freememory.YAxis = "SystemTime"
	m_Freememory.MaxX = 15
	a.Client.AddStatMetadata(m_Freememory)

	m_BufferSize := ceb.StatMetadata{}
	m_BufferSize.Name = "Total Memory"
	m_BufferSize.Type = "line"
	m_BufferSize.XAxis = "BufferSize"
	m_BufferSize.YAxis = "SystemTime"
	m_BufferSize.MaxX = 15
	a.Client.AddStatMetadata(m_BufferSize)

	m_TotalSwapMemory := ceb.StatMetadata{}
	m_TotalSwapMemory.Name = "Total Memory"
	m_TotalSwapMemory.Type = "line"
	m_TotalSwapMemory.XAxis = "TotalSwapMemory"
	m_TotalSwapMemory.YAxis = "SystemTime"
	m_TotalSwapMemory.MaxX = 15
	a.Client.AddStatMetadata(m_TotalSwapMemory)

	m_UsedSwapMemory := ceb.StatMetadata{}
	m_UsedSwapMemory.Name = "Total Memory"
	m_UsedSwapMemory.Type = "line"
	m_UsedSwapMemory.XAxis = "UsedSwapMemory"
	m_UsedSwapMemory.YAxis = "SystemTime"
	m_UsedSwapMemory.MaxX = 15
	a.Client.AddStatMetadata(m_UsedSwapMemory)

	m_FeeSwapMemory := ceb.StatMetadata{}
	m_FeeSwapMemory.Name = "Total Memory"
	m_FeeSwapMemory.Type = "line"
	m_FeeSwapMemory.XAxis = "FeeSwapMemory"
	m_FeeSwapMemory.YAxis = "SystemTime"
	m_FeeSwapMemory.MaxX = 15
	a.Client.AddStatMetadata(m_FeeSwapMemory)
}
