package processors

import (
	"duov6.com/common"
	"duov6.com/objectstore/cache"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
	"duov6.com/objectstore/storageengines"
	"encoding/json"
	"fmt"
	"strconv"
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
		storageEngine = storageengines.ReplicatedStorageEngine{}
	case "SINGLE":
		storageEngine = storageengines.SingleStorageEngine{}
	}

	var outResponse repositories.RepositoryResponse = storageEngine.Store(request)

	if request.IsLogEnabled || !outResponse.IsSuccess {
		url := "/" + request.Controls.Namespace + "/" + request.Controls.Class
		fileBody := "------------------- Default Request -------------------\r\n"
		fileBody += "URL : " + url + "\r\n"
		requestInBytes, _ := json.Marshal(request.Body)
		fileBody += "Request Body : " + string(requestInBytes) + "\r\n"
		for index, element := range request.MessageStack {
			fileBody += "S-" + strconv.Itoa(index) + " : " + element + "\r\n"
		}

		common.PublishLog("ObjectStoreLog.log", fileBody)

		if d.CheckRedisAvailability(request) && !outResponse.IsSuccess {
			if request.Controls.Operation == "insert" || request.Controls.Operation == "update" {
				_ = cache.StoreKeyValue(request, d.GetKeyNameForLog(request), fileBody, cache.Log)
			}
		}

	}

	return outResponse
}

func (d *Dispatcher) GetKeyNameForLog(request *messaging.ObjectRequest) (val string) {
	if request.Controls.Multiplicity == "single" {
		val = "ErrorSinglePostLog." + request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id
	} else {
		val = "ErrorMultiplePostLog." + request.Controls.Namespace + "." + request.Controls.Class + "." + common.GetGUID()
	}
	fmt.Println(val)
	return
}

func (d *Dispatcher) CheckRedisAvailability(request *messaging.ObjectRequest) (status bool) {
	status = true
	if request.Configuration.ServerConfiguration["REDIS"] == nil {
		status = false
	}
	return
}
