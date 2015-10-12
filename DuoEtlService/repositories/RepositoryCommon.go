package repositories

import (
	"duov6.com/DuoEtlService/messaging"
	"encoding/json"
	//"fmt"
	"github.com/twinj/uuid"
	"reflect"
	"strconv"
	"time"
)

func getDataType(item interface{}) (datatype string) {
	//fmt.Println(item)
	datatype = reflect.TypeOf(item).Name()
	if datatype == "bool" {
		datatype = "boolean"
	} else if datatype == "float64" {
		datatype = "real"
	} else if datatype == "" || datatype == "string" || datatype == "ControlHeaders" {
		datatype = "text"
	}
	return datatype
}

func fillControlHeaders(request *messaging.ETLRequest) {
	currentTime := getTime()
	if request.Body.Object != nil {
		controlObject := messaging.ControlHeaders{}
		controlObject.Version = uuid.NewV1().String()
		controlObject.Namespace = request.Controls.Namespace
		controlObject.Class = request.Controls.Class
		controlObject.Tenant = "123"
		controlObject.LastUpdated = currentTime
		request.Body.Object["__osHeaders"] = controlObject
	} else {
		for _, obj := range request.Body.Objects {
			controlObject := messaging.ControlHeaders{}
			controlObject.Version = uuid.NewV1().String()
			controlObject.Namespace = request.Controls.Namespace
			controlObject.Class = request.Controls.Class
			controlObject.Tenant = "123"
			controlObject.LastUpdated = currentTime
			obj["__osHeaders"] = controlObject
		}
	}
}

func getTime() (retTime string) {
	currentTime := time.Now().Local()
	year := strconv.Itoa(currentTime.Year())
	month := strconv.Itoa(int(currentTime.Month()))
	day := strconv.Itoa(currentTime.Day())
	hour := strconv.Itoa(currentTime.Hour())
	minute := strconv.Itoa(currentTime.Minute())
	second := strconv.Itoa(currentTime.Second())

	retTime = (year + "-" + month + "-" + day + "T" + hour + ":" + minute + ":" + second)

	return
}

func getStringByObject(obj interface{}) string {

	result, err := json.Marshal(obj)

	if err == nil {
		return string(result)
	} else {
		return "{}"
	}
}

func ConvertOsheaders(input messaging.ControlHeaders) string {
	myStr := "{\"Class\":\"" + input.Class + "\",\"LastUpdated\":\"2" + input.LastUpdated + "\",\"Namespace\":\"" + input.Namespace + "\",\"Tenant\":\"" + input.Tenant + "\",\"Version\":\"" + input.Version + "\"}"
	return myStr
}
