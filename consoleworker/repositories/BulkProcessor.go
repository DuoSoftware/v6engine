package repositories

import (
	"duov6.com/consoleworker/structs"
	"fmt"
)

type BulkProcessor struct {
}

func (repository BulkProcessor) GetWorkerName(request structs.ServiceRequest) string {
	return "BulkProcessor"
}

func (repository BulkProcessor) ProcessWorker(request structs.ServiceRequest) structs.ServiceResponse {
	response := structs.ServiceResponse{}
	fmt.Println(request)
	return response
}
