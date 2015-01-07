package processmanager

import (
	ObjectStoreMessaging "duov6.com/objectstore/messaging"
	//"duov6.com/objectstore/repositories"
	//"duov6.com/objectstore/storageengines"
	"duov6.com/serviceconsole/messaging"
	"fmt"
	//"github.com/streadway/amqp"
	"encoding/json"
	//"log"
)

type QueuedObjectStoreWorker struct {
}

func (queue QueuedObjectStoreWorker) GetWorkerName() string {
	return "QueuedObjectStoreWorker"
}

func (queue QueuedObjectStoreWorker) ExecuteWorker(request *messaging.ServiceRequest) messaging.ServiceResponse {
	response := messaging.ServiceResponse{}

	if request.Body != nil {
		recievedRequest := ObjectStoreMessaging.ObjectRequest{}
		fmt.Println(".....................................................................")
		fmt.Println(recievedRequest)
		fmt.Println(".....................................................................")
		err := json.Unmarshal(request.Body, &recievedRequest)

		if err != nil {
			fmt.Println("Unmarshal sucked!")
		}

		response.IsSuccess = true
		fmt.Println(".....................................................................")
		fmt.Println(recievedRequest)
		recievedRequest.IsLogEnabled = true
		recievedRequest.Log("buhahaha")
		fmt.Println("Key Property : " + recievedRequest.Body.Parameters.KeyProperty)
		fmt.Println("Key Value : " + recievedRequest.Body.Parameters.KeyValue)
		//var myEngine = storageengines.ReplicatedStorageEngine{}
		//var outResponse repositories.RepositoryResponse = myEngine.Store(&recievedRequest)
		//	fmt.Println(outResponse)

	} else {
		response.IsSuccess = false
	}

	return response
}
