package repositories

import (
	"duov6.com/consoleworker/structs"
	"fmt"
)

func Execute(request structs.ServiceRequest) structs.ServiceResponse {
	result := structs.ServiceResponse{}
	var repository AbstractRepository
	repository = Create(request.OperationCode)
	fmt.Println("Executing : " + repository.GetWorkerName(request))
	result = repository.ProcessWorker(request)
	return result
}
