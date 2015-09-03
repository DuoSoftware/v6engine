package processmanager

import (
	"duov6.com/serviceconsole/messaging"
	"fmt"
	"log"
)

type ExcelWorker struct {
}

func (worker ExcelWorker) GetWorkerName() string {
	return "ExcelWorker"
}

func (worker ExcelWorker) ExecuteWorker(request *messaging.ServiceRequest) messaging.ServiceResponse {
	fmt.Println("Not Implemented in Excel Worker in RabbitMQ. Use FileServer.FileManager.Store(*FileRequest)")
	var temp = messaging.ServiceResponse{}
	if request.Body != nil {
		log.Printf("Received a message: %s", request.Body)
	}
	temp.IsSuccess = false
	return temp
}
