package processors

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
	"duov6.com/objectstore/storageengines"
	"fmt"
	"strconv"
)

type Dispatcher struct {
}

func (d *Dispatcher) Dispatch(request *messaging.ObjectRequest) repositories.RepositoryResponse {

	var storageEngine storageengines.AbstractStorageEngine // request.StoreConfiguration.StorageEngine

	switch request.Configuration.StorageEngine {
	case "REPLICATED":
		request.Log("Starting replicated storage engine")
		storageEngine = storageengines.ReplicatedStorageEngine{}
	case "SINGLE":
		storageEngine = storageengines.SingleStorageEngine{}
	case "QUEUED":
		storageEngine = storageengines.QueuedStorageEngine{}
	}

	var outResponse repositories.RepositoryResponse = storageEngine.Store(request)

	if request.IsLogEnabled {
		for index, element := range request.MessageStack {
			fmt.Println("S-" + strconv.Itoa(index) + " : " + element)
		}
	}

	return outResponse
}
