package Transaction

import (
	"duov6.com/objectstore/cache"
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	Command   = 0
	Operation = 1
)

func NewTransaction(request *messaging.ObjectRequest) (err error) {
	transactionID := request.Body.Parameters.TransactionID
	transactionStruct := request.Body.Transaction

	if transactionID == "" && strings.EqualFold(transactionStruct.Type, "BEGIN") {
		transactionID = GetTransactionID()

		metadata := make(map[string]interface{})
		metadata["TimeStamp"] = time.Now().Format("2006-01-02 15:04:05")
		metadata["TransactionID"] = transactionID

		bucketValue, _ := json.Marshal(metadata)
		err = cache.RPush(request, GetBucketName(transactionID), string(bucketValue))
	} else {
		err = errors.New("No Transaction Command Found!")
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

func CreateBlockEntry(request *messaging.ObjectRequest, TransactionID string) (err error) {
	//entry := GetBlockEntryName(request, TransactionID)
	//err = cache.StoreValue(request, entry, request)
	return
}

func DeleteBlockEntry(request *messaging.ObjectRequest, TransactionID string) (err error) {
	//entry := GetBlockEntryName(request)
	//err = cache.DeleteKey(request, entry)
	return
}

func CheckBlockEntry(request *messaging.ObjectRequest, TransactionID string) (status bool) {
	//entry := GetBlockEntryName(request, TransactionID)
	return
}
