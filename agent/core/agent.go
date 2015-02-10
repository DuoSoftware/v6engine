package core

import (
	"duov6.com/fws"
)

type Agent struct {
	IsAgentEnabled bool
	Client         *fws.FWSClient
	ListnerName    string
}
