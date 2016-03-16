package repositories

import (
	"duov6.com/consoleworker/structs"
)

type AbstractRepository interface {
	GetWorkerName(request structs.ServiceRequest) string
	ProcessWorker(request structs.ServiceRequest) structs.ServiceResponse
}
