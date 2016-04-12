package processors

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
	"duov6.com/objectstore/storageengines"
	"github.com/twinj/uuid"
	//"duov6.com/term"
	//"fmt"
)

type Transaction struct {
}

func (t *Transaction) ProcessTransaction(request *messaging.ObjectRequest) repositories.RepositoryResponse {
	var outResponse repositories.RepositoryResponse

	//Fake Logic!
	var storageEngine storageengines.AbstractStorageEngine
	storageEngine = storageengines.ReplicatedStorageEngine{}
	outResponse = storageEngine.Store(request)
	return outResponse
}

func (t *Transaction) GetRequestType(request *messaging.ObjectRequest) (reqType string) {
	if request.Body.Parameters.TransactionID != "" {
		reqType = "operation"
	} else if request.Body.Transaction.Type != "" {
		reqType = "command"
	}
	return
}

func (t *Transaction) GetTransactionID() string {
	return uuid.NewV1().String()
}
