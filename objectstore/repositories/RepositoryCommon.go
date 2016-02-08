package repositories

import (
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"github.com/twinj/uuid"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func getDefaultNotImplemented() RepositoryResponse {
	return RepositoryResponse{IsSuccess: false, IsImplemented: false, Message: "Operation Not Implemented"}
}

func getDataType(item interface{}) (datatype string) {
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

func FillControlHeaders(request *messaging.ObjectRequest) {
	//currentTime := time.Now().Local().String()
	currentTime := getTime()
	if request.Controls.Multiplicity == "single" {
		controlObject := messaging.ControlHeaders{}
		controlObject.Version = uuid.NewV1().String()
		controlObject.Namespace = request.Controls.Namespace
		controlObject.Class = request.Controls.Class
		controlObject.Tenant = "123"
		controlObject.LastUdated = currentTime

		request.Body.Object["__osHeaders"] = controlObject
	} else {
		for _, obj := range request.Body.Objects {
			controlObject := messaging.ControlHeaders{}
			controlObject.Version = uuid.NewV1().String()
			controlObject.Namespace = request.Controls.Namespace
			controlObject.Class = request.Controls.Class
			controlObject.Tenant = "123"
			controlObject.LastUdated = currentTime
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

func getNoSqlKey(request *messaging.ObjectRequest) string {
	key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id
	return key
}

func getNoSqlKeyById(request *messaging.ObjectRequest, obj map[string]interface{}) string {
	key := request.Controls.Namespace + "." + request.Controls.Class + "." + obj[request.Body.Parameters.KeyProperty].(string)
	return key
}

func getNoSqlKeyByGUID(request *messaging.ObjectRequest) string {
	namespace := request.Controls.Namespace
	class := request.Controls.Class
	key := namespace + "." + class + "." + uuid.NewV1().String()
	return key
}

func getStringByObject(obj interface{}) string {

	result, err := json.Marshal(obj)

	if err == nil {
		return string(result)
	} else {
		return "{}"
	}
}

func getByteByValue(input interface{}) (value []byte) {
	value, _ = json.Marshal(input)
	return
}

func getSQLnamespace(request *messaging.ObjectRequest) string {
	return (strings.Replace(request.Controls.Namespace, ".", "", -1))
}

func ConvertOsheaders(input messaging.ControlHeaders) string {
	myStr := "{\"Class\":\"" + input.Class + "\",\"LastUdated\":\"2" + input.LastUdated + "\",\"Namespace\":\"" + input.Namespace + "\",\"Tenant\":\"" + input.Tenant + "\",\"Version\":\"" + input.Version + "\"}"
	return myStr
}

func getEmptyByteObject() (returnByte []byte) {
	var empty []map[string]interface{}
	empty = make([]map[string]interface{}, 0)
	returnByte, _ = json.Marshal(empty)
	return
}

func checkEmptyByteArray(input []byte) (status bool) {
	status = false
	if len(input) == 4 || len(input) == 2 || len(input) < 2 || input == nil {
		status = true
	}
	return
}
