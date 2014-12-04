package repositories

import (
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"github.com/twinj/uuid"
	"time"
)

func getDefaultNotImplemented() RepositoryResponse {
	return RepositoryResponse{IsSuccess: false, IsImplemented: false, Message: "Operation Not Implemented"}
}

func FillControlHeaders(request *messaging.ObjectRequest) {
	currentTime := time.Now().Local().String()
	if request.Controls.Multiplicity == "single" {
		controlObject := messaging.ControlHeaders{}
		controlObject.Version = uuid.NewV1().String()
		controlObject.Namespace = request.Controls.Namespace
		controlObject.Class = request.Controls.Class
		controlObject.Tenant = "123"
		controlObject.LastUdated = string(currentTime)

		request.Body.Object["__osHeaders"] = controlObject
	} else {
		for _, obj := range request.Body.Objects {
			controlObject := messaging.ControlHeaders{}
			controlObject.Version = uuid.NewV1().String()
			controlObject.Namespace = request.Controls.Namespace
			controlObject.Class = request.Controls.Class
			controlObject.Tenant = "123"
			controlObject.LastUdated = string(currentTime)

			obj["__osHeaders"] = controlObject
		}
	}
}

func getNoSqlKey(request *messaging.ObjectRequest) string {
	key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id
	return key
}

func getNoSqlKeyById(request *messaging.ObjectRequest, obj map[string]interface{}) string {
	key := request.Controls.Namespace + "." + request.Controls.Class + "." + obj[request.Body.Parameters.KeyProperty].(string)
	return key
}

func getStringByObject(request *messaging.ObjectRequest, obj map[string]interface{}) string {

	result, err := json.Marshal(obj)

	if err == nil {
		return string(result)
	} else {
		return "{}"
	}
}
