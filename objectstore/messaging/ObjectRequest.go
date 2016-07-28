package messaging

import (
	"duov6.com/objectstore/configuration"
	"duov6.com/term"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
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

	if reflect.TypeOf(value).String() == "string" {
		message = value.(string)
	} else {
		byteArray, err := json.Marshal(value)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		message = string(byteArray)
	}

	o.MessageStack = append(o.MessageStack, message)

	if o.IsLogEnabled || strings.Contains(strings.ToLower(message), "error") || strings.Contains(strings.ToLower(message), "info") {

		lowerCasedMsg := strings.ToLower(message)

		if strings.Contains(lowerCasedMsg, "error") {
			term.Write(value, term.Error)
		} else if strings.Contains(lowerCasedMsg, "debug") {
			term.Write(value, term.Debug)
		} else {
			term.Write(value, term.Information)
		}

	}
}
