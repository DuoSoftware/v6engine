package processors

import (
	"duov6.com/common"
	"duov6.com/objectstore/Transaction"
	"duov6.com/objectstore/cache"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
	"encoding/json"
	"strconv"
	//"duov6.com/objectstore/storageengines"
)

type TransactionDispatcher struct {
}

func (t *TransactionDispatcher) DispatchTransaction(request *messaging.ObjectRequest) repositories.RepositoryResponse {
	var outResponse repositories.RepositoryResponse
	outResponse = t.ExecuteTransaction(request)

	//.............................................
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

		if t.CheckRedisAvailability(request) && !outResponse.IsSuccess {
			if request.Controls.Operation == "insert" || request.Controls.Operation == "update" {
				_ = cache.StoreKeyValue(request, t.GetTNameForLog(request), fileBody, cache.Log)
			}
		}
	}

	if !request.IsLogEnabled {
		request.MessageStack = make([]string, 0)
	}

	//............................................
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

		//Enable Logs forcefully
		if !request.IsLogEnabled {
			request.IsLogEnabled = true
			var initialSlice []string
			initialSlice = make([]string, 0)
			request.MessageStack = initialSlice
		}

		outResponse = Transaction.ExecuteCommand(request)
	}

	return outResponse
}

func (t *TransactionDispatcher) GetTNameForLog(request *messaging.ObjectRequest) (val string) {
	val = "ErrorTransactionLog." + request.Body.Parameters.TransactionID
	return
}

func (t *TransactionDispatcher) CheckRedisAvailability(request *messaging.ObjectRequest) (status bool) {
	status = true
	if request.Configuration.ServerConfiguration["REDIS"] == nil {
		status = false
	}
	return
}
