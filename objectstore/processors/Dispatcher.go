package processors

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
	"duov6.com/objectstore/storageengines"
)

type Dispatcher struct {
}

func (d *Dispatcher) Dispatch(request *messaging.ObjectRequest) repositories.RepositoryResponse {

	transactionID := request.Body.Parameters.TransactionID
	transactionStruct := request.Body.Transaction

	var outResponse repositories.RepositoryResponse

	if transactionID != "" || transactionStruct.Type != "" {
		request.Log("Transaction Request")
		if repositories.CheckRedisAvailability(request) {
			var t TransactionDispatcher
			outResponse = t.DispatchTransaction(request)
		} else {
			outResponse.IsSuccess = false
			outResponse.Message = "REDIS not found! Please Config REDIS for ObjectStore for Transactions!"
		}
	} else {
		request.Log("Default Request")
		outResponse = d.ProcessDefaultDispatcher(request)
	}

	return outResponse
}

func (d *Dispatcher) ProcessDefaultDispatcher(request *messaging.ObjectRequest) repositories.RepositoryResponse {

	var storageEngine storageengines.AbstractStorageEngine // request.StoreConfiguration.StorageEngine

	switch request.Configuration.StorageEngine {
	case "REPLICATED":
		request.Log("Starting replicated storage engine")
		storageEngine = storageengines.ReplicatedStorageEngine{}
	case "SINGLE":
		storageEngine = storageengines.SingleStorageEngine{}
	}

	var outResponse repositories.RepositoryResponse = storageEngine.Store(request)

	//Commented here because need to fmt is when executing. Saving for future references.
	// if request.IsLogEnabled {
	// 	for index, element := range request.MessageStack {
	// 		request.Log("S-" + strconv.Itoa(index) + " : " + element)
	// 	}
	// }

	return outResponse
}
