package processors

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
	"duov6.com/objectstore/storageengines"
	"duov6.com/term"
	//"fmt"
	"strconv"
)

type Dispatcher struct {
}

func (d *Dispatcher) Dispatch(request *messaging.ObjectRequest) repositories.RepositoryResponse {

	transactionID := request.Body.Parameters.TransactionID
	transactionStruct := request.Body.Transaction

	var outResponse repositories.RepositoryResponse

	if transactionID != "" || transactionStruct.Type != "" {
		term.Write("Transaction Request", term.Error)
		var t Transaction
		outResponse = t.ProcessTransaction(request)
	} else {
		term.Write("Default Request", term.Error)
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

	if request.IsLogEnabled {
		for index, element := range request.MessageStack {
			term.Write("S-"+strconv.Itoa(index)+" : "+element, term.Debug)
		}
	}

	return outResponse
}
