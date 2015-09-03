package processmanager

import (
	"duov6.com/serviceconsole/messaging"
	"fmt"
	"log"
)

type QueuedObjectStoreWorker struct {
}

func (queue QueuedObjectStoreWorker) GetWorkerName() string {
	return "QueuedObjectStoreWorker"
}

func (queue QueuedObjectStoreWorker) ExecuteWorker(request *messaging.ServiceRequest) messaging.ServiceResponse {
	fmt.Println("Not Implemented in Queued ObjectStore for RabbitMQ. Use REPLICATED Object Store")
	var temp = messaging.ServiceResponse{}
	if request.Body != nil {
		log.Printf("Received a message: %s", request.Body)
	}
	temp.IsSuccess = false
	return temp
}
