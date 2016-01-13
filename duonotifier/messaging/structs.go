package messaging

import (
	"duov6.com/common"
	"duov6.com/duonotifier/configuration"
	"fmt"
)

type TemplateRequest struct {
	TemplateID    string
	DefaultParams map[string]string
	CustomParams  map[string]string
	Namespace     string
}

type NotifierResponse struct {
	IsSuccess bool
	Message   string
}

type RequestNotifyControls struct {
	SecurityToken string
	Namespace     string
	Class         string
}

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

type ServiceRequest struct {
	SecurityToken string
	NotifyMethod  string
	Parameters    map[string]interface{}
}
