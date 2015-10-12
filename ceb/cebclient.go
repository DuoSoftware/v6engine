package ceb

import (
	"bufio"
	"duov6.com/ceb/messaging"
	"encoding/json"
	"encoding/base64"
	"fmt"
	"net"
)

/*
   { to: 'Agent 1',
     command: 'matricMessage',
     data: { state: 'on', from: 'admin' },
     persistIfOffline: false,
     alwaysPersist: false },
*/

type CEBClient struct {
	outgoing chan string
	reader   	*bufio.Reader
	writer   	*bufio.Writer
	connection 	*net.Conn
	listener func(s messaging.CEBTCPCommand)

	agentName      string
	events         map[string]func(from string, name string, data map[string]interface{}, resources map[string]interface{})
	commands       map[string]func(from string, name string, data map[string]interface{}, resources map[string]interface{})
	CommandMaps    []CommandMap
	StatMetadata   []StatMetadata
	ConfigMetadata []ConfigMetadata
	CanMonitorOutput bool
	Resources      	map[string]interface{}
	ListenerName   string
}

var curentClient *CEBClient

func GetClient() *CEBClient {
	return curentClient
}

func NewCEBClient(host string, callback func(s bool)) (client *CEBClient, e error) {
	connection, e := net.Dial("tcp", host)

	if e == nil {
		writer := bufio.NewWriter(connection)
		reader := bufio.NewReader(connection)

		c := &CEBClient{
			outgoing: make(chan string),
			reader:   reader,
			writer:   writer,
			connection: &connection,
		}

		c.events = make(map[string]func(from string, name string, data map[string]interface{}, resources map[string]interface{}))
		c.commands = make(map[string]func(from string, name string, data map[string]interface{}, resources map[string]interface{}))
		c.Resources = make(map[string]interface{})
		c.CommandMaps = make([]CommandMap, 0)
		c.StatMetadata = make([]StatMetadata, 0)
		c.ConfigMetadata = make([]ConfigMetadata, 0)
		c.ListenerName = ""

		curentClient = c
		client = c

		client.listener = func(s messaging.CEBTCPCommand) {

			data := s.Data.(map[string]interface{})

			if data["message"] !=nil{
				if data["message"] == "Successfully Registered!!!" {
					callback(true);
				}
			}


			if data["type"] !=nil {
				if data["type"] == "command" {

					if data["name"] != nil {

						var fwsCommandName = data["name"].(string)

						if fwsCommandName == "agentCommand" {
							var subData = data["data"].(map[string]interface{})

							var fromUser string

							if subData["from"] != nil {
								fromUser = subData["from"].(string)
							} else {
								fromUser = "UNKNOWN"
							}
							var commandName = subData["command"].(string)
							var commandData map[string]interface{}

							if subData["data"] != nil {
								commandData = subData["data"].(map[string]interface{})
							} else {
								commandData = make(map[string]interface{})
							}

							c.executeCommand(fromUser, commandName, commandData, c.Resources)
						}
					}
				}else {
					var eventData = data["data"].(map[string]interface{});
					c.executeEvent(eventData["userName"].(string), data["name"].(string), eventData, c.Resources);
				}
			}
		}

		client.listen()
	}

	return
}

func (client *CEBClient) executeCommand(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
	if client.commands[name] != nil {
		client.commands[name](from, name, data, resources)
	}
}

func (client *CEBClient) executeEvent(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
	if client.events[name] != nil {
		client.events[name](from, name, data, resources)
	}
}
func (client *CEBClient) read() {

	for {
		if client != nil {
			if client.reader != nil {
				base64Line, _ := client.reader.ReadBytes('|')
				if (len(base64Line) !=0){
					line,err := base64.StdEncoding.DecodeString(string(base64Line[:len(base64Line)-1]))
					commandObject := messaging.CEBTCPCommand{}
					err = json.Unmarshal(line[:len(line)], &commandObject)
					if err != nil {
						fmt.Println("Error : " + err.Error())
					} else {
						client.listener(commandObject)
					}
				}else{
					//assume that the connection is disconnected
				}

			}
		}
	}
}

func (client *CEBClient) write() {
	for data := range client.outgoing {
		
		encData := base64.StdEncoding.EncodeToString([]byte(data))

		fmt.Println("AGENT SEND : ")
		client.writer.WriteString(encData)
		client.writer.Flush()
	}
}

func (client *CEBClient) listen() {
	go client.read()
	go client.write()
}

func (client *CEBClient) Msg(message string) {
	client.outgoing <- message
}

func (client *CEBClient) Register(userName string, securityToken string) {

	client.agentName = userName
	tcpComm := messaging.CEBTCPCommand{}
	tcpComm.Command = "register"
	regComm := messaging.RegisterTCPCommand{}
	regComm.SecurityToken = securityToken
	regComm.UserName = userName
	regComm.ResourceClass = "server"

	tcpComm.Data = regComm

	byteSet, err := json.Marshal(&tcpComm)

	if err != nil {
		fmt.Println(err.Error())
	}

	message := string(byteSet[:len(byteSet)])

	client.outgoing <- string(message)
}

func (client *CEBClient) ClientCommand(to string, class string, typ string, data interface{}) {
	//client.outgoing <- message

	tcpComm := messaging.CEBTCPCommand{}
	tcpComm.Command = "command"

	fwsCommand := messaging.CommandTCPCommand{}
	fwsCommand.Name = "commandforward"
	fwsCommand.Type = "command"

	forwardComm := messaging.CommandForwardTCPParamters{}

	forwardComm.To = to
	forwardComm.Command = "agentResponse"
	forwardComm.PersistIfOffline = false
	forwardComm.AlwaysPersist = false

	servMonitorCommand := messaging.ServerMonitorCommand{}
	servMonitorCommand.Class = class
	servMonitorCommand.Type = typ
	servMonitorCommand.Data = data

	forwardComm.Data = servMonitorCommand

	fwsCommand.Data = forwardComm
	tcpComm.Data = fwsCommand

	/*

		To               string      `json:"to"`
		Command          string      `json:"command"`
		Data             interface{} `json:"data"`
		PersistIfOffline bool        `json:"persistIfOffline"`
		AlwaysPersist    bool        `json:"alwaysPersist"`
	*/

	byteSet, err := json.Marshal(&tcpComm)

	if err != nil {
		fmt.Println(err.Error())
	}

	message := string(byteSet[:len(byteSet)])

	client.outgoing <- string(message)

}

func (client *CEBClient) ExecuteCommand(name string, data map[string]interface{}) {
	invoke(client, name, "command", data)
}

func (client *CEBClient) TriggerEvent(name string, data map[string]interface{}) {
	invoke(client, name, "event", data)
}

func invoke(client *CEBClient, name string, typ string, data map[string]interface{}) {
	//client.outgoing <- message

	tcpComm := messaging.CEBTCPCommand{}
	tcpComm.Command = "command"

	fwsCommand := messaging.CommandTCPCommand{}
	fwsCommand.Name = name
	fwsCommand.Type = typ
	data["from"] = client.agentName
	fwsCommand.Data = data

	tcpComm.Data = fwsCommand

	byteSet, err := json.Marshal(&tcpComm)

	if err != nil {
		fmt.Println(err.Error())
	}

	message := string(byteSet[:len(byteSet)])

	client.outgoing <- string(message)

}

func (client *CEBClient) AddListener(f func(s messaging.CEBTCPCommand)) {
	client.listener = f
}

func (client *CEBClient) OnCommand(name string, proc func(from string, name string, data map[string]interface{}, resources map[string]interface{})) {
	client.commands[name] = proc
}

func (client *CEBClient) OnEvent(name string, proc func(from string, name string, data map[string]interface{}, resources map[string]interface{})) {
	client.events[name] = proc
	subscribeMap := make(map[string]interface{})
	subscribeMap["userName"] = client.agentName;
	invoke(client, name, "event-subscribe", subscribeMap)
}

func (client *CEBClient) AddCommandMetadata(m CommandMap) {
	client.CommandMaps = append(client.CommandMaps, m)
}

func (client *CEBClient) UpdateCommandMetadata(name string, newSettings map[string]interface{}) {
	for _,command := range client.ConfigMetadata {
		if command.Code == name {
			for k,v := range newSettings{
				command.Parameters[k] = v;
			}
			break
		}
  	}
}

func (client *CEBClient) AddStatMetadata(m StatMetadata) {
	client.StatMetadata = append(client.StatMetadata, m)
}

func (client *CEBClient) AddConfigMetadata(m ConfigMetadata) {
	client.ConfigMetadata = append(client.ConfigMetadata, m)
}

func (client *CEBClient) UpdateStats(statInterface interface{}){
	client.ClientCommand(client.ListenerName, "stat", "test", statInterface)
}
