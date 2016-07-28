package repositories

import (
	"duov6.com/common"
	"duov6.com/objectstore/keygenerator"
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"fmt"
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
		controlObject.Version = common.GetGUID()
		controlObject.Namespace = request.Controls.Namespace
		controlObject.Class = request.Controls.Class
		controlObject.Tenant = "123"
		controlObject.LastUdated = currentTime

		request.Body.Object["__osHeaders"] = controlObject
	} else {
		for _, obj := range request.Body.Objects {
			controlObject := messaging.ControlHeaders{}
			controlObject.Version = common.GetGUID()
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
	key := namespace + "." + class + "." + common.GetGUID()
	return key
}

func getDomainClassAttributesKey(request *messaging.ObjectRequest) (key string) {
	key = request.Controls.Namespace + ".domainClassAttributes." + request.Controls.Class
	return
}

func getStringByObject(obj interface{}) string {
	//fmt.Println("********************************************************")
	value := ""
	result, err := json.Marshal(obj)

	if err == nil {
		//	fmt.Println("Successfully Created String Object for interface object")
		value = string(result)
	} else {
		fmt.Println(err.Error())
		value = "{}"
	}
	//fmt.Println("********************************************************")
	return value
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

func CheckRedisAvailability(request *messaging.ObjectRequest) (status bool) {
	status = true
	if request.Configuration.ServerConfiguration["REDIS"] == nil {
		status = false
	}
	//remove this when going Live
	//status = true
	return
}

func GetRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
	isGUIDKey := false
	isAutoIncrementId := false //else MANUAL key from the user

	//multiple requests
	if (obj[request.Body.Parameters.KeyProperty].(string) == "-999") || (request.Body.Parameters.AutoIncrement == true) {
		isAutoIncrementId = true
	}

	if (obj[request.Body.Parameters.KeyProperty].(string) == "-888") || (request.Body.Parameters.GUIDKey == true) {
		isGUIDKey = true
	}

	if isGUIDKey {
		returnID = common.GetGUID()
	} else if isAutoIncrementId {
		if CheckRedisAvailability(request) {
			returnID = keygenerator.GetIncrementID(request, "COMMON", 0)
		} else {
			request.Log("Debug : WARNING! : Returning GUID since REDIS not available and not concurrent safe!")
			returnID = common.GetGUID()
		}
	} else {
		returnID = obj[request.Body.Parameters.KeyProperty].(string)
	}
	return
}
