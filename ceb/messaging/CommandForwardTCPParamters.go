package messaging

type CommandForwardTCPParamters struct {
	To               string      `json:"to"`
	Command          string      `json:"command"`
	Data             interface{} `json:"data"`
	PersistIfOffline bool        `json:"persistIfOffline"`
	AlwaysPersist    bool        `json:"alwaysPersist"`
}
