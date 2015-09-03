package messaging

import (
	"duov6.com/common"
	"duov6.com/serviceconsole/configuration"
)

type ServiceRequest struct {
	RefID             string
	RefType           string
	OperationCode     string
	ScheduleTimeStamp string
	ControlParameters map[string]string
	Parameters        map[string]string
	Body              []byte
	Configuration     configuration.StoreServiceConfiguration

	IsLogEnabled bool
	MessageStack []string
}

func (s *ServiceRequest) Log(message string) {
	s.MessageStack = append(s.MessageStack, message)
	common.PublishLog("ServiceConsole.log", message)
}
