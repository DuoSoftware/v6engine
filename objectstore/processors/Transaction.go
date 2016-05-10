package processors

import (
	"duov6.com/common"
	"duov6.com/objectstore/Transaction"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
	//"duov6.com/objectstore/storageengines"
)

type TransactionDispatcher struct {
}

func (t *TransactionDispatcher) DispatchTransaction(request *messaging.ObjectRequest) repositories.RepositoryResponse {
	var outResponse repositories.RepositoryResponse

	//Fake Logic!
	// var storageEngine storageengines.AbstractStorageEngine
	// storageEngine = storageengines.ReplicatedStorageEngine{}
	// outResponse = storageEngine.Store(request)
	// return outResponse

	outResponse = t.ExecuteTransaction(request)
	return outResponse
}

func (t *TransactionDispatcher) GetRequestType(request *messaging.ObjectRequest) (reqType int) {
	if request.Body.Parameters.TransactionID != "" {
		reqType = Transaction.Operation
	} else if request.Body.Transaction.Type != "" {
		reqType = Transaction.Command
	}
	return
}

func (t *TransactionDispatcher) GetTransactionID() string {
	return common.GetGUID()
}

func (t *TransactionDispatcher) ExecuteTransaction(request *messaging.ObjectRequest) repositories.RepositoryResponse {
	var outResponse repositories.RepositoryResponse

	requestType := t.GetRequestType(request)
	if requestType == Transaction.Operation {
		outResponse = Transaction.ExecuteOperation(request)
	} else if requestType == Transaction.Command {
		outResponse = Transaction.ExecuteCommand(request)
	}

	return outResponse
}
