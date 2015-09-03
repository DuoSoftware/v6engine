package processmanager

import (
	"duov6.com/serviceconsole/messaging"
	"fmt"
	"log"
)

type WorkFlowWorker2 struct {
}

func (worker WorkFlowWorker2) GetWorkerName() string {
	return "WorkFlowWorker2"
}

func (worker WorkFlowWorker2) ExecuteWorker(request *messaging.ServiceRequest) messaging.ServiceResponse {
	fmt.Println("Not Implemented in WorkFlow Reconnect in RabbitMQ.")
	var temp = messaging.ServiceResponse{}
	if request.Body != nil {
		log.Printf("Received a message: %s", request.Body)
	}
	temp.IsSuccess = false
	return temp
}
