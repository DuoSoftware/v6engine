package endpoints

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
	"duov6.com/objectstore/storageengines"
)

type Dispatcher struct {
}

func (d *Dispatcher) Dispatch(request *messaging.ObjectRequest) repositories.RepositoryResponse {

	var storageEngine storageengines.AbstractStorageEngine // request.StoreConfiguration.StorageEngine

	switch request.Configuration.StorageEngine {
	case "REPLICATED":
		storageEngine = storageengines.ReplicatedStorageEngine{}
	case "SINGLE":
		storageEngine = storageengines.SingleStorageEngine{}
	}

	var outResponse repositories.RepositoryResponse = storageEngine.Store(request)

	return outResponse
}
