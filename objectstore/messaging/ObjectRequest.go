package messaging

import (
	//"duov6.com/common"
	"duov6.com/objectstore/configuration"
	//"fmt"
	"duov6.com/term"
	"encoding/json"
	"reflect"
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
		if reflect.TypeOf(value).String() == "string" {
			message = value.(string)
		} else {
			byteArray, _ := json.Marshal(value)
			message = string(byteArray)
		}
		term.Write(value, term.Error)
		o.MessageStack = append(o.MessageStack, message)
		//common.PublishLog("ObjectStoreLog.log", message)
	}
}
