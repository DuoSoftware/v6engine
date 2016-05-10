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
		break
	case "ROLLBACK":
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
		err := cache.RPush(request, GetBucketName(transactionID), string(bucketValue))
		if err != nil {
			request.Log(err.Error())
		}
	} else {
		transactionID = ""
	}

	return
}

func AppendTransaction(request *messaging.ObjectRequest) (err error) {
	transactionID := request.Body.Parameters.TransactionID
	transactionStruct := request.Body.Transaction

	if transactionID != "" && transactionStruct.Type == "" {
		bucketValue, _ := json.Marshal(request)
		err = cache.RPush(request, GetBucketName(transactionID), string(bucketValue))
	} else {
		err = errors.New("No TransactionID Found!")
	}
	return
}

//Block List Implementation. Next release

func AppendBlockEntry(request *messaging.ObjectRequest, TransactionID string) (err error) {
	entry := GetBlockEntryName(request, TransactionID)
	if cache.ExistsKeyValue(request, entry) {
		var data []interface{}
		byteValue := cache.GetValue(request, entry)
		err = json.Unmarshal(byteValue, &data)
		data = append(data, request)
		byteValue = nil
		byteValue, err = json.Marshal(data)
		err = cache.StoreValue(request, entry, string(byteValue))
	} else {
		var data []interface{}
		data = append(data, request)
		byteValue, _ := json.Marshal(data)
		err = cache.StoreValue(request, entry, string(byteValue))
	}
	return
}

func DeleteBlockEntry(request *messaging.ObjectRequest, TransactionID string) (err error) {
	entry := GetBlockEntryName(request, TransactionID)
	status := cache.DeleteKey(request, entry)
	if !status {
		err = errors.New("Delete Failed!")
	}
	return
}

func VerifyBlockSafe(request *messaging.ObjectRequest) (status bool) {
	// 	TransactionID := request.Body.Parameters.TransactionID
	// 	entry := GetBlockEntryName(request, TransactionID)
	// 	if cache.ExistsKeyValue(request, entry) {
	// 		var data []interface{}
	// 		byteValue := cache.GetValue(request, entry)
	// 		err := json.Unmarshal(byteValue, &data)

	// 		var objects []map[string]interface{}
	// 		if request.Body.Object != nil {
	// 			objects = make([]map[string]interface{}, 1)
	// 			objects[0] = request.Body.Object
	// 		} else {
	// 			objects = make([]map[string]interface{}, len(request.Body.Objects))
	// 			objects = request.Body.Objects
	// 		}

	// 		for _, singularObject := range objects {
	// 			for _, singleData := range data {
	// 				if singleData.Body.Object != nil {
	// 					if (singularObject[request.Body.Parameters.KeyProperty].(string) == singleData[request.Body.Parameters.KeyProperty].(string)) && (TransactionID != singleData.Body.Parameters.TransactionID) {
	// 						status = false
	// 						return
	// 					}
	// 				} else {
	// 					for _, subObjects := range singleData.Body.Objects {
	// 						if (singularObject[request.Body.Parameters.KeyProperty].(string) == subObjects[request.Body.Parameters.KeyProperty].(string)) && (TransactionID != singleData.Body.Parameters.TransactionID) {
	// 							status = false
	// 							return
	// 						}
	// 					}
	// 				}
	// 			}
	// 		}

	// 	} else {
	// 		status = false
	// 	}
	status = true
	return
}
