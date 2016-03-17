package repositories

import (
	"duov6.com/consoleworker/common"
	"duov6.com/consoleworker/structs"
	"errors"
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

	configs := common.GetConfigurations()

	smoothFlowUrl := configs["SVC_SMOOTHFLOW_URL"].(string)

	if request.Parameters["JSONData"] != nil {

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
	} else {
		response.Err = errors.New("Required Fields such as JSONData Not Found to Execute SmoothFlow Worker!")
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
