package repositories

import (
	"duov6.com/consoleworker/common"
	"duov6.com/consoleworker/structs"
	"fmt"
)

type SmoothFlowProcessor struct {
}

func (repository SmoothFlowProcessor) GetWorkerName(request structs.ServiceRequest) string {
	return "SmoothFlowProcessor"
}

func (repository SmoothFlowProcessor) ProcessWorker(request structs.ServiceRequest) structs.ServiceResponse {
	response := structs.ServiceResponse{}
	fmt.Println(request)

	configs := common.GetConfigurations()

	smoothFlowUrl := configs["SVC_SMOOTHFLOW_URL"].(string)

	err := common.PostHTTPRequest(smoothFlowUrl, request.Parameters)
	if err != nil {
		fmt.Println(err.Error())
		response.Err = err
	} else {
		response.Err = nil
	}

	return response
}
