package Transaction

import (
	"duov6.com/objectstore/cache"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	Command   = 0
	Operation = 1
)

func ExecuteCommand(request *messaging.ObjectRequest) repositories.RepositoryResponse {

	var response repositories.RepositoryResponse

	tranactionCommand := strings.ToUpper(request.Body.Transaction.Type)
	var err error
	switch tranactionCommand {
	case "BEGIN":
		TransactionID := NewTransaction(request)
		if TransactionID != "" {
			response.Transaction.TransactionID = TransactionID
		} else {
			err = errors.New("Error Creating New Transaction!")
		}
		break
	case "COMMIT":
		err = CommitTransaction(request)
		break
	case "ROLLBACK":
		err = RollbackTransaction(request)
		break
	}

	if err != nil {
		response.IsSuccess = false
		response.Message = err.Error()
	} else {
		response.IsSuccess = true
		response.Message = "Successfully Executed Transaction COMMAND!"
	}

	return response
}

func ExecuteOperation(request *messaging.ObjectRequest) repositories.RepositoryResponse {
	var response repositories.RepositoryResponse
	var err error
	err = AppendTransaction(request)
	if err != nil {
		response.IsSuccess = false
		response.Message = err.Error()
	} else {
		response.IsSuccess = true
		response.Message = "Successfully Executed Transaction OPERATION!"
	}
	return response
}

func NewTransaction(request *messaging.ObjectRequest) (transactionID string) {
	transactionID = request.Body.Parameters.TransactionID
	transactionStruct := request.Body.Transaction

	if transactionID == "" && strings.EqualFold(transactionStruct.Type, "BEGIN") {
		transactionID = GetTransactionID()

		metadata := make(map[string]interface{})
		metadata["TimeStamp"] = time.Now().Format("2006-01-02 15:04:05")
		metadata["TransactionID"] = transactionID

		bucketValue, _ := json.Marshal(metadata)
		err := cache.RPush(request, GetBucketName(transactionID), string(bucketValue), cache.Transaction)
		if err != nil {
			request.Log(err.Error())
		}
	} else {
		transactionID = ""
	}

	return
}

func RollbackTransaction(request *messaging.ObjectRequest) (err error) {
	//delete transaction key
	TransactionID := request.Body.Transaction.Parameters["TransactionID"].(string)
	status := cache.DeleteKey(request, GetBucketName(TransactionID), cache.Transaction)
	if !status {
		err = errors.New("Couldn't Expire Transaction!")
	}
	return
}

func AppendTransaction(request *messaging.ObjectRequest) (err error) {
	transactionID := request.Body.Parameters.TransactionID
	transactionStruct := request.Body.Transaction

	if transactionID != "" && transactionStruct.Type == "" {
		bucketValue, _ := json.Marshal(request)
		err = cache.RPush(request, GetBucketName(transactionID), string(bucketValue), cache.Transaction)
	} else {
		err = errors.New("No TransactionID Found!")
	}
	return
}

func CommitTransaction(request *messaging.ObjectRequest) (err error) {
	TLog(request, request.Body.Transaction.Parameters["TransactionID"].(string))
	err = Execute(request)
	if err == nil {
		TransactionID := request.Body.Transaction.Parameters["TransactionID"].(string)
		_ = cache.DeleteKey(request, GetBucketName(TransactionID), cache.Transaction)
		_ = cache.DeleteKey(request, GetSuccessBucketName(TransactionID), cache.Transaction)
		_ = cache.DeleteKey(request, GetInvertBucketName(TransactionID), cache.Transaction)
	} else {
		//Produce Commit Logs!
	}
	return
}
