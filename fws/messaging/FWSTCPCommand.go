package messaging

type FWSTCPCommand struct {
	Command string      `json:"command"`
	Data    interface{} `json:"data"`
}
