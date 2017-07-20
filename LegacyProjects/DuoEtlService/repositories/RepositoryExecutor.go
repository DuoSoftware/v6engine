package repositories

import (
	"duov6.com/DuoEtlService/logger"
	"duov6.com/DuoEtlService/messaging"
	"fmt"
)

func Execute(request *messaging.ETLRequest) messaging.ETLResponse {
	var response messaging.ETLResponse
	logger.Log("Running Executor ->>>>>>")
	fmt.Println(request)
	logger.Log("\n")
	response.IsSuccess = true
	response.Message = "Executer Done!"
	return response
}
