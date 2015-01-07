package messaging

type RegisterTCPCommand struct {
	UserName      string `json:"userName"`
	SecurityToken string `json:"securityToken"`
}
