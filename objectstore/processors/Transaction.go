package processors

import (
	"duov6.com/common"
	"duov6.com/objectstore/Transaction"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
	"duov6.com/objectstore/storageengines"
)

type TransactionDispatcher struct {
}

func (t *TransactionDispatcher) DispatchTransaction(request *messaging.ObjectRequest) repositories.RepositoryResponse {
	var outResponse repositories.RepositoryResponse

	//Fake Logic!
	var storageEngine storageengines.AbstractStorageEngine
	storageEngine = storageengines.ReplicatedStorageEngine{}
	outResponse = storageEngine.Store(request)
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
