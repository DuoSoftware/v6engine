package messaging

import (
	"duov6.com/DuoEtlService/configuration"
)

type ETLRequest struct {
	Body          RequestBody
	Controls      RequestETLControls
	Parameters    map[string]interface{}
	Configuration configuration.ETLConfiguration
}
