package repositories

import (
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
	return response
}
