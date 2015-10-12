package repositories

import (
	//"duov6.com/DuoEtlService/logger"
	"duov6.com/DuoEtlService/messaging"
)

type KibanaRepository struct {
}

func (repo KibanaRepository) GetETLName() string {
	return "KIBANA"
}

func (repo KibanaRepository) ExecuteETLService(request *messaging.ETLRequest) messaging.ETLResponse {
	var response messaging.ETLResponse
	response.IsSuccess = true
	response.Message = "Kibana Successfully Executed!"
	return response
}
