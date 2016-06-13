package client

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/processors"
	"duov6.com/objectstore/repositories"
	"fmt"
	"github.com/fatih/structs"
	//"strconv"
	"duov6.com/common"
	"errors"
	"reflect"
	"strconv"
	"strings"
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

	m.Request.Body.Object = bodyMap
	m.FillControlHeaders(m.Request)
	return m
}

func (m *StoreModifier) AndStoreMany(objs []interface{}) *StoreModifier {
	m.Request.Controls.Multiplicity = "multiple"

	var interfaceList []map[string]interface{}
	if strings.Contains(reflect.TypeOf(objs).String(), "map") {
		interfaceList = make([]map[string]interface{}, len(objs))
		for _, obj := range objs {
			interfaceList = append(interfaceList, obj.(map[string]interface{}))
		}
	} else {
		s := reflect.ValueOf(objs)

		interfaceList = make([]map[string]interface{}, s.Len())

		for i := 0; i < s.Len(); i++ {
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
	}
	m.Request.Body.Objects = interfaceList
	m.FillControlHeaders(m.Request)
	return m
}

func (m *StoreModifier) AndStoreManyObjects(objs []map[string]interface{}) *StoreModifier {
	m.Request.Controls.Multiplicity = "multiple"
	m.Request.Body.Objects = objs
	return m
}

func (m *StoreModifier) AndStoreMapInterface(objs []map[string]interface{}) *StoreModifier {
	m.Request.Controls.Multiplicity = "multiple"
	m.Request.Body.Objects = objs
	return m
}

func (m *StoreModifier) Ok() (err error) {
	if m.Request.Controls.Multiplicity == "single" {
		m.Request.Controls.Id = m.Request.Body.Object[m.Request.Body.Parameters.KeyProperty].(string)
	}

	dispatcher := processors.Dispatcher{}

	repositories.FillControlHeaders(m.Request)

	response := dispatcher.Dispatch(m.Request)

	if !response.IsSuccess {
		if response.Message == "" {
			err = errors.New("Error Storing Object! : Undefined Error!")
		} else {
			err = errors.New(response.Message)
		}
	}

	return
}

func (m *StoreModifier) FileOk() repositories.RepositoryResponse {
	if m.Request.Controls.Multiplicity == "single" {
		m.Request.Controls.Id = m.Request.Body.Object[m.Request.Body.Parameters.KeyProperty].(string)
	}

	dispatcher := processors.Dispatcher{}

	response := dispatcher.Dispatch(m.Request)

	return response
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

func (m *StoreModifier) FillControlHeaders(request *messaging.ObjectRequest) {
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
