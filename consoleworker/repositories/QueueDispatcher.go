package repositories

import (
	"duov6.com/consoleworker/common"
	"duov6.com/consoleworker/structs"
	"errors"
	"fmt"
	"strings"
)

type QueueDispatcher struct {
}

func (repository QueueDispatcher) GetWorkerName(request structs.ServiceRequest) string {
	return "Queue Dispatcher Repository"
}

func (repository QueueDispatcher) ProcessWorker(request structs.QueueDispatchRequest) structs.ServiceResponse {
	response := structs.ServiceResponse{}

	configs := common.GetConfigurations()

	namespace, class := getNamespaceAndClass(request.Objects[0].RefType)

	dispatcherURL := configs["SVC_DISPATCHER_URL"].(string) + namespace + "/" + class

	isError := false

	for _, serviceRequest := range request.Objects {
		err := common.PostHTTPRequest(dispatcherURL, serviceRequest)
		if err != nil {
			fmt.Println(err.Error())
			isError = true
		}
	}

	if isError {
		response.Err = errors.New("Error pushing all requests to Queue Manager.. Check Requests..")
	} else {
		response.Err = nil
	}

	return response
}

func getNamespaceAndClass(input string) (namespace string, class string) {
	tokens := strings.Split(input, ".")
	class = tokens[(len(tokens) - 1)]
	namespace = ""

	for x := 0; x < (len(tokens) - 1); x++ {
		namespace += tokens[x] + "."
	}

	namespace = strings.TrimSuffix(namespace, ".")

	return
}
