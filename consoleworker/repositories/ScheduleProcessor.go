package repositories

import (
	"duov6.com/consoleworker/structs"
	"fmt"
)

type ScheduleProcessor struct {
}

func (repository ScheduleProcessor) GetWorkerName(request structs.ServiceRequest) string {
	return "Schedule Processor Repository"
}

func (repository ScheduleProcessor) ProcessWorker(request structs.ServiceRequest) structs.ServiceResponse {
	response := structs.ServiceResponse{}
	fmt.Println("Not Planned what to do. Just scheduling things! :P ")
	return response
}
