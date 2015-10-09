package client

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/processors"
	"duov6.com/objectstore/repositories"
	"fmt"
	"github.com/fatih/structs"
	//"strconv"
	"github.com/twinj/uuid"
	"reflect"
	"strconv"
	"time"
)

type StoreModifier struct {
	Request *messaging.ObjectRequest
}

func (m *StoreModifier) WithKeyField(field string) *StoreModifier {
	m.Request.Body.Parameters = messaging.ObjectParameters{}
	m.Request.Body.Parameters.KeyProperty = field
	return m
}

func (m *StoreModifier) AndStoreOne(obj interface{}) *StoreModifier {

	m.Request.Controls.Multiplicity = "single"
	v := reflect.ValueOf(obj)
	k := v.Kind()

	var bodyMap map[string]interface{}

	if k != reflect.Map {
		bodyMap = structs.Map(obj)
	} else {
		bodyMap = v.Interface().(map[string]interface{})
	}
	//fmt.Println("SDFASDFASDF")
	//fmt.Println("CONVERTED " , bodyMap)

	m.Request.Body.Object = bodyMap
	controlObject := messaging.ControlHeaders{}
	controlObject.Version = uuid.NewV1().String()
	controlObject.Namespace = m.Request.Controls.Namespace
	controlObject.Class = m.Request.Controls.Class
	controlObject.Tenant = "123"
	controlObject.LastUdated = getTime()
	m.Request.Body.Object["__osHeaders"] = controlObject
	return m
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

func (m *StoreModifier) AndStoreMany(objs []interface{}) *StoreModifier {
	m.Request.Controls.Multiplicity = "multiple"

	s := reflect.ValueOf(objs)
	var interfaceList []map[string]interface{}
	interfaceList = make([]map[string]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		//newMap := structs.Map(s.Index(i).Interface())
		obj := s.Index(i).Interface()
		v := reflect.ValueOf(obj)
		k := v.Kind()
		fmt.Println("KIND : ", k)
		var newMap map[string]interface{}

		if k != reflect.Map {
			newMap = structs.Map(obj)
		} else {
			newMap = obj.(map[string]interface{})
		}

		interfaceList[i] = newMap
	}

	//for index, element := range objs {
	//	interfaceList[index] = structs.Map(element)
	//}

	m.Request.Body.Objects = interfaceList
	return m
}

func (m *StoreModifier) AndStoreMapInterface(objs []map[string]interface{}) *StoreModifier {
	m.Request.Controls.Multiplicity = "multiple"
	m.Request.Body.Objects = objs
	return m
}

func (m *StoreModifier) Ok() {
	if m.Request.Controls.Multiplicity == "single" {
		m.Request.Controls.Id = m.Request.Body.Object[m.Request.Body.Parameters.KeyProperty].(string)
	}

	dispatcher := processors.Dispatcher{}

	repositories.FillControlHeaders(m.Request)

	response := dispatcher.Dispatch(m.Request)

	fmt.Println(response.IsSuccess)
}

func (m *StoreModifier) FileOk() []map[string]interface{} {
	if m.Request.Controls.Multiplicity == "single" {
		m.Request.Controls.Id = m.Request.Body.Object[m.Request.Body.Parameters.KeyProperty].(string)
	}

	dispatcher := processors.Dispatcher{}

	response := dispatcher.Dispatch(m.Request)

	fmt.Println(response.IsSuccess)
	return response.Data
}
func NewStoreModifier(request *messaging.ObjectRequest) *StoreModifier {
	modifier := StoreModifier{Request: request}
	modifier.Request = request
	modifier.Request.Controls.Operation = "insert"
	modifier.Request.Body = messaging.RequestBody{}
	return &modifier
}

func NewStoreModifierWithOperation(request *messaging.ObjectRequest, operation string) *StoreModifier {
	modifier := StoreModifier{Request: request}
	modifier.Request = request
	modifier.Request.Controls.Operation = operation
	modifier.Request.Body = messaging.RequestBody{}
	return &modifier
}
