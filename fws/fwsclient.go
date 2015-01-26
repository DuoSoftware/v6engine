package fws

import (
	"bufio"
	"duov6.com/fws/messaging"
	"encoding/json"
	"fmt"
	"net"
)

type RegisterTCPCommand struct {
	UserName      string `json:"userName"`
	SecurityToken string `json:"securityToken"`
}

/*
   { to: 'Agent 1',
     command: 'matricMessage',
     data: { state: 'on', from: 'admin' },
     persistIfOffline: false,
     alwaysPersist: false },
*/

type FWSClient struct {
	outgoing chan string
	reader   *bufio.Reader
	writer   *bufio.Writer
	listener func(s messaging.FWSTCPCommand)

	events         map[string]func(from string, name string, data map[string]interface{}, resources map[string]interface{})
	commands       map[string]func(from string, name string, data map[string]interface{}, resources map[string]interface{})
	CommandMaps    []CommandMap
	StatMetadata   []StatMetadata
	ConfigMetadata []ConfigMetadata
	Resources      map[string]interface{}
}

var curentClient *FWSClient

func GetClient() *FWSClient {
	return curentClient
}

func NewFWSClient(host string) (client *FWSClient, e error) {
	connection, e := net.Dial("tcp", host)

	if e == nil {
		writer := bufio.NewWriter(connection)
		reader := bufio.NewReader(connection)

		c := &FWSClient{
			outgoing: make(chan string),
			reader:   reader,
			writer:   writer,
		}

		c.events = make(map[string]func(from string, name string, data map[string]interface{}, resources map[string]interface{}))
		c.commands = make(map[string]func(from string, name string, data map[string]interface{}, resources map[string]interface{}))
		c.Resources = make(map[string]interface{})
		c.CommandMaps = make([]CommandMap, 0)
		c.StatMetadata = make([]StatMetadata, 0)
		c.ConfigMetadata = make([]ConfigMetadata, 0)

		curentClient = c
		client = c

		client.listener = func(s messaging.FWSTCPCommand) {

			data := s.Data.(map[string]interface{})

			if data["name"] != nil {

				var fwsCommandName = data["name"].(string)

				if fwsCommandName == "agentCommand" {
					var subData = data["data"].(map[string]interface{})

					fromUser := subData["from"].(string)

					var commandName = subData["command"].(string)
					var commandData = subData["data"].(map[string]interface{})

					c.execute("command", fromUser, commandName, commandData, c.Resources)
					/*
						if commandName == "switch" {


						}
					*/
				}

			}

		}

		client.listen()
	}

	return
}

func (client *FWSClient) execute(typ string, from string, name string, data map[string]interface{}, resources map[string]interface{}) {
	if typ == "command" {
		client.commands[name](from, name, data, resources)
	} else {
		client.events[name](from, name, data, resources)
	}
}

func (client *FWSClient) read() {

	for {
		if client != nil {
			if client.reader != nil {
				line, _ := client.reader.ReadBytes('|')

				commandObject := messaging.FWSTCPCommand{}

				err := json.Unmarshal(line[:len(line)-1], &commandObject)
				if err != nil {
					fmt.Println("Error : " + err.Error())
				} else {
					client.listener(commandObject)
				}
			}
		}
	}
}

func (client *FWSClient) write() {
	for data := range client.outgoing {
		fmt.Println("SEND : ", data)
		client.writer.WriteString(data)
		client.writer.Flush()
	}
}

func (client *FWSClient) listen() {
	go client.read()
	go client.write()
}

func (client *FWSClient) Msg(message string) {
	client.outgoing <- message
}

func (client *FWSClient) Register(userName string, securityToken string) {

	tcpComm := messaging.FWSTCPCommand{}
	tcpComm.Command = "register"
	regComm := messaging.RegisterTCPCommand{}
	regComm.SecurityToken = securityToken
	regComm.UserName = userName
	tcpComm.Data = regComm

	byteSet, err := json.Marshal(&tcpComm)

	if err != nil {
		fmt.Println(err.Error())
	}

	message := string(byteSet[:len(byteSet)])

	client.outgoing <- string(message)
}

func (client *FWSClient) ClientCommand(to string, class string, typ string, data interface{}) {
	//client.outgoing <- message

	tcpComm := messaging.FWSTCPCommand{}
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

func (client *FWSClient) AddListener(f func(s messaging.FWSTCPCommand)) {
	client.listener = f
}

func (client *FWSClient) Subscribe(typ string, name string, proc func(from string, name string, data map[string]interface{}, resources map[string]interface{})) {
	if typ == "command" {
		client.commands[name] = proc
	} else {
		client.events[name] = proc
	}
}

func (client *FWSClient) AddCommandMetadata(m CommandMap) {
	client.CommandMaps = append(client.CommandMaps, m)
}

func (client *FWSClient) AddStatMetadata(m StatMetadata) {
	client.StatMetadata = append(client.StatMetadata, m)
}

func (client *FWSClient) AddConfigMetadata(m ConfigMetadata) {
	client.ConfigMetadata = append(client.ConfigMetadata, m)
}
