package client

import (
	"duov6.com/duonotifier/messaging"
	"duov6.com/objectstore/client"
	"encoding/json"
	//"fmt"
	"strings"
)

func GetTemplate(request messaging.TemplateRequest) (response interface{}) {
	//response = request
	data := make(map[string]interface{})
	isDefault := false
	bytes, _ := client.Go("securityToken", request.Namespace, "templates").GetOne().ByUniqueKey(request.TemplateID).Ok()
	if len(bytes) <= 4 {
		isDefault = true
		bytes, _ = client.Go("securityToken", "com.duosoftware.com", "templates").GetOne().ByUniqueKey(request.TemplateID).Ok()
	}
	_ = json.Unmarshal(bytes, &data)
	result := getFormatted(request, data, isDefault)
	response = result
	return
}

//template code... structure it as soon as possible.

func getFormatted(request messaging.TemplateRequest, data map[string]interface{}, isDefault bool) map[string]interface{} {

	keyWordMap := make(map[string]string)

	dataCopy := make(map[string]interface{})

	for key, value := range data {
		dataCopy[key] = value
	}

	if isDefault {
		for key, value := range request.DefaultParams {
			keyWordMap[key] = value
		}
	} else {
		for key, value := range request.CustomParams {
			keyWordMap[key] = value
		}
	}

	for key, value := range dataCopy {

		if key != "TemplateID" && !strings.Contains(key, "osHeaders") {
			stringVal := value.(string)
			//fmt.Println(stringVal)
			for param, paramVal := range keyWordMap {
				stringVal = strings.Replace(stringVal, param, paramVal, -1)
			}
			dataCopy[key] = stringVal
		}
	}

	return dataCopy
}
