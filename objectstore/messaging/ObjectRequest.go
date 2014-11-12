package messaging

import (
	"duov6.com/objectstore/configuration"
)

type ObjectRequest struct {
	Controls      RequestControls
	Body          RequestBody
	Configuration configuration.StoreConfiguration
	Extras        map[string]interface{}
}
