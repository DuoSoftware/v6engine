package messaging

import (
	"duov6.com/common"
	"duov6.com/duonotifier/configuration"
	"fmt"
)

type NotifierRequest struct {
	NotifyMethod  string
	Controls      RequestNotifyControls
	Parameters    map[string]interface{}
	Configuration configuration.NotifierConfiguration
}

func (s *NotifierRequest) Log(message string) {
	fmt.Println(message)
	common.PublishLog("DuoNotifierLog.log", message)
}
