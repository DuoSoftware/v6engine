package messaging

import (
	//"duov6.com/common"
	"duov6.com/objectstore/configuration"
	"fmt"
)

type ObjectRequest struct {
	Controls      RequestControls
	Body          RequestBody
	Configuration configuration.StoreConfiguration
	Extras        map[string]interface{}

	IsLogEnabled bool
	MessageStack []string
}

func (o *ObjectRequest) Log(value interface{}) {
	var message string
	if o.IsLogEnabled {
		o.MessageStack = append(o.MessageStack, message)
		fmt.Println(value)
		//common.PublishLog("ObjectStoreLog.log", message)
	}
}
