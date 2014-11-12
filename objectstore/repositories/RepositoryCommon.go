package repositories

import (
	"duov6.com/objectstore/messaging"
)

func getDefaultNotImplemented() RepositoryResponse {
	return RepositoryResponse{IsSuccess: false, IsImplemented: false, Message: "Operation Not Implemented"}
}

func FillControlHeaders(request *messaging.ObjectRequest) {

	if request.Controls.Multiplicity == "single" {
		controlObject := messaging.ControlHeaders{}
		controlObject.Version = "xxx-xxx-xxx-xxx"
		controlObject.Namespace = request.Controls.Namespace
		controlObject.Class = request.Controls.Class
		controlObject.Tenant = "123"
		controlObject.LastUdated = "xx/xx/xxxx xx:xx:xx"

		request.Body.Object["__osHeaders"] = controlObject
	} else {
		for _, obj := range request.Body.Objects {
			controlObject := messaging.ControlHeaders{}
			controlObject.Version = "xxx-xxx-xxx-xxx"
			controlObject.Namespace = request.Controls.Namespace
			controlObject.Class = request.Controls.Class
			controlObject.Tenant = "123"
			controlObject.LastUdated = "xx/xx/xxxx xx:xx:xx"

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
