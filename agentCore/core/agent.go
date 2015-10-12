package core

import (
	"duov6.com/ceb"
)

type Agent struct {
	IsAgentEnabled bool
	Client         *ceb.CEBClient
}
