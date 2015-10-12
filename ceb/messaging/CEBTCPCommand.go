package messaging

type CEBTCPCommand struct {
	Command string      `json:"command"`
	Data    interface{} `json:"data"`
}
