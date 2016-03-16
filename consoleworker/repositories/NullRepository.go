package repositories

import (
	"duov6.com/consoleworker/structs"
	"errors"
)

type NullRepository struct {
}

func (repository NullRepository) GetWorkerName(request structs.ServiceRequest) string {
	return "Null Repository"
}

func (repository NullRepository) ProcessWorker(request structs.ServiceRequest) structs.ServiceResponse {
	response := structs.ServiceResponse{}
	response.Err = errors.New("Repository Not Found! Available Repositories are BulkProcessor/SmoothFlow only. Check request!")
	return response
}
