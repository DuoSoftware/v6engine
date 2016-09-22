package Transaction

import (
	"duov6.com/objectstore/cache"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
	"duov6.com/objectstore/storageengines"
	"encoding/json"
	"errors"
	//"fmt"
	"strconv"
)

func Execute(request *messaging.ObjectRequest) (err error) {
	//Get Length of List
	TransactionID := request.Body.Transaction.Parameters["TransactionID"].(string)
	if cache.ExistsKeyValue(request, GetBucketName(TransactionID), cache.Transaction) && cache.GetListLength(request, GetBucketName(TransactionID), cache.Transaction) > 1 {
		err = StartProcess(request)
	} else {
		err = errors.New("Error : No such Transaction items found or either already rollbacked!")
		request.Log("Error : No such Transaction items found or either already rollbacked!")
	}
	return
}

func StartProcess(request *messaging.ObjectRequest) (err error) {
	//GetTask
	request.Log("Debug : Starting Executing Transactions Items.")
	TransactionID := request.Body.Transaction.Parameters["TransactionID"].(string)
	tasklength := cache.GetListLength(request, GetBucketName(TransactionID), cache.Transaction)

	var x int64
	//pop first element and throw away
	_, _ = cache.LPop(request, GetBucketName(TransactionID), cache.Transaction)

	for x = 0; x < tasklength-1; x++ {
		taskNo := strconv.Itoa(int(x))
		request.Log("Debug : Executing Element No : " + taskNo)
		pickedRequest, err2 := GetTask(request)
		if err2 != nil {
			//Rollback executed while executing last processs -> Execute rollback process
			err = StartRollBackProcess(request)
			if err != nil {
				err = errors.New("Debug : Successfully Rolledback because Rollback was triggered!")
			}
			return
		} else {
			//execute
			invertedRequests := GetInvertedRequests(pickedRequest)
			response := ProcessDispatcher(pickedRequest)
			//if success -> Push to success list, Create invert request and push to invert list
			if response.IsSuccess {
				request.Log("Debug : Successfully Executed Element No : " + taskNo)
				_ = PushToSuccessList(pickedRequest, TransactionID)
				_ = PushToInvertList(invertedRequests, TransactionID)
				//update log
				//UpdateLogStatus(int(x), TransactionID, "TRUE")
			} else { //if false -> Start rollback process
				//UpdateLogStatus(int(x), TransactionID, "FALSE")
				request.Log("Error : Error Executing Element No : " + taskNo + ". Starting Rollback!")
				err = StartRollBackProcess(request)
				if err != nil {
					err = errors.New("Debug : Successfully Rolledback because Rollback was triggered!")
				}
				return
			}
		}
	}
	return
}

func ProcessDispatcher(request *messaging.ObjectRequest) repositories.RepositoryResponse {
	var storageEngine storageengines.AbstractStorageEngine // request.StoreConfiguration.StorageEngine
	storageEngine = storageengines.ReplicatedStorageEngine{}
	var outResponse repositories.RepositoryResponse = storageEngine.Store(request)
	return outResponse
}

func GetTask(request *messaging.ObjectRequest) (retRequest *messaging.ObjectRequest, err error) {
	TransactionID := request.Body.Transaction.Parameters["TransactionID"].(string)
	var byteVal []byte
	byteVal, err = cache.LPop(request, GetBucketName(TransactionID), cache.Transaction)
	// if err != nil -> key has removed. RollBack has been called
	if err == nil {
		if len(byteVal) <= 4 {
			err = errors.New("Rollbacked!")
			return nil, err
		}
		err2 := json.Unmarshal(byteVal, &retRequest)
		if err2 != nil {
			request.Log(err2.Error())
		}
	} else {
		request.Log(err.Error())
	}
	return
}

func GetInvertedTask(request *messaging.ObjectRequest) (retRequest *messaging.ObjectRequest, err error) {
	TransactionID := request.Body.Transaction.Parameters["TransactionID"].(string)
	var byteVal []byte
	byteVal, err = cache.LPop(request, GetInvertBucketName(TransactionID), cache.Transaction)

	// if err != nil -> key has removed.. RollBack has been called
	if err == nil {
		err2 := json.Unmarshal(byteVal, &retRequest)
		if err2 != nil {
			request.Log(err2.Error())
		}
	}
	return
}

func PushToSuccessList(request *messaging.ObjectRequest, TransactionID string) (err error) {
	bucketValue, err := json.Marshal(request)
	err = cache.RPush(request, GetSuccessBucketName(TransactionID), string(bucketValue), cache.Transaction)
	return
}

func PushToInvertList(request []*messaging.ObjectRequest, TransactionID string) (err error) {
	for _, singleRequest := range request {
		bucketValue, _ := json.Marshal(singleRequest)
		err = cache.RPush(singleRequest, GetInvertBucketName(TransactionID), string(bucketValue), cache.Transaction)
	}
	return
}

func StartRollBackProcess(request *messaging.ObjectRequest) (err error) {
	TransactionID := request.Body.Transaction.Parameters["TransactionID"].(string)
	tasklength := cache.GetListLength(request, GetInvertBucketName(TransactionID), cache.Transaction)
	isAllSuccess := true

	var x int64

	for x = 0; x < tasklength; x++ {
		pickedRequest, _ := GetInvertedTask(request)
		response := ProcessDispatcher(pickedRequest)
		if !response.IsSuccess {
			isAllSuccess = false
		}
	}

	if !isAllSuccess {
		err = errors.New("Not All Rollbacks were successful!")
	}

	return
}
