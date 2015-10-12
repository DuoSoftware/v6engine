package messaging

type ServerMonitorCommand struct {
	Class string      `json:"class"`
	Type  string      `json:"type"`
	Data  interface{} `json:"data"`
}
