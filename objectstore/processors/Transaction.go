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
	var storageEngine storageengines.AbstractStorageEngine
	storageEngine = storageengines.ReplicatedStorageEngine{}
	var outResponse repositories.RepositoryResponse = storageEngine.Store(request)
	return outResponse
}

// func (t *Transaction) GetRequestType() string {
// 	if transactionID != "" {
// 		return "operation"
// 	} else if transactionStruct.Type != "" {
// 		return "command"
// 	}
// }

func (t *Transaction) GetTransactionID() string {
	return uuid.NewV1().String()
}
