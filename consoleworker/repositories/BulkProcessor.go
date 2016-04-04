package repositories

import (
	duocommon "duov6.com/common"
	"duov6.com/consoleworker/common"
	"duov6.com/consoleworker/objectstore"
	"duov6.com/consoleworker/structs"
	"errors"
	"fmt"
	"strings"
	"time"
)

type BulkProcessor struct {
}

func (repository BulkProcessor) GetWorkerName(request structs.ServiceRequest) string {
	return "BulkProcessor"
}

func (repository BulkProcessor) ProcessWorker(request structs.ServiceRequest) structs.ServiceResponse {
	response := structs.ServiceResponse{}

	fileName := request.Parameters["FileName"].(string)
	data := ReadDataFromObjectStore("tasks.serviceconsole.payload", strings.Replace(fileName, ".xlsx", "", -1))

	if len(data) == 0 {
		response.Err = errors.New("No Objects Found!")
		return response
	}

	serviceRequests := createServiceRequests(data, request)

	queueDispatchRequest := structs.QueueDispatchRequest{}
	queueDispatchRequest.Objects = serviceRequests

	queueDispather := QueueDispatcher{}
	queueDispather.ProcessWorker(queueDispatchRequest)

	return response
}

func ReadDataFromObjectStore(namespace string, class string) (data []map[string]interface{}) {
	time.Sleep(time.Second * 5) //just to wait till it insert to elastic since im testing on elastic. remove this on actual cloud host
	data, err := objectstore.GetAll(namespace, class)
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}

func createServiceRequests(data []map[string]interface{}, request structs.ServiceRequest) (serviceRequests []structs.ServiceRequest) {
	serviceRequests = make([]structs.ServiceRequest, len(data))

	for x := 0; x < len(data); x++ {
		serviceRequest := structs.ServiceRequest{}
		serviceRequest.RefId = duocommon.GetGUID()
		serviceRequest.RefType = request.RefType
		serviceRequest.OperationCode = "SmoothFlow"
		serviceRequest.TimeStamp = common.GetTime()
		serviceRequest.TimeStampReadable = common.GetTimeReadable()
		serviceRequest.Parameters = data[x]
		serviceRequest.Body = []byte("Request Originated from Bulk Processor in DuoWorker @ " + serviceRequest.TimeStamp)
		serviceRequests[x] = serviceRequest
	}

	return
}
