package repositories

import (
	"duov6.com/consoleworker/common"
	"duov6.com/consoleworker/structs"
	"fmt"
	"strings"
)

type SmoothFlowProcessor struct {
}

func (repository SmoothFlowProcessor) GetWorkerName(request structs.ServiceRequest) string {
	return "SmoothFlowProcessor"
}

func (repository SmoothFlowProcessor) ProcessWorker(request structs.ServiceRequest) structs.ServiceResponse {
	response := structs.ServiceResponse{}
	fmt.Println(request)

	configs := common.GetConfigurations()

	smoothFlowUrl := configs["SVC_SMOOTHFLOW_URL"].(string)

	json := JsonBuilder(request.Parameters["JSONData"].(map[string]interface{}))

	object := request.Parameters
	object["JSONData"] = json

	err := common.PostHTTPRequest(smoothFlowUrl, request.Parameters)
	if err != nil {
		fmt.Println(err.Error())
		response.Err = err
	} else {
		response.Err = nil
	}

	return response
}

func JsonBuilder(data map[string]interface{}) (json string) {
	json = ""

	for key, value := range data {
		json += "\"" + key + "\":\"" + value.(string) + "\", "
	}

	json = strings.TrimSuffix(json, ".")

	return
}
