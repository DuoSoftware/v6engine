package client

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/processors"
	//"duov6.com/objectstore/repositories"
	"github.com/fatih/structs"
	//"strconv"
	//"fmt"
	"errors"
	"reflect"
	"strings"
)

type DeleteModifier struct {
	Request *messaging.ObjectRequest
}

func (m *DeleteModifier) ByUniqueKey(key string) *DeleteModifier {
	m.Request.Controls.Operation = "delete"
	m.Request.Controls.Multiplicity = "single"
	m.Request.Controls.Id = key
	return m
}

func (m *DeleteModifier) AndDeleteObject(obj interface{}) *DeleteModifier {
	return m.AndDeleteOne(obj)
}

func (m *DeleteModifier) AndDeleteOne(obj interface{}) *DeleteModifier {
	m.Request.Controls.Operation = "delete"
	m.Request.Controls.Multiplicity = "single"

	bodyMap := make(map[string]interface{})

	if strings.Contains(reflect.TypeOf(obj).String(), "map") {
		bodyMap = obj.(map[string]interface{})
	} else {
		bodyMap = structs.Map(obj)
	}

	key := bodyMap[m.Request.Body.Parameters.KeyProperty].(string)
	m.Request.Body.Object = bodyMap
	m.Request.Controls.Id = key
	return m
}

func (m *DeleteModifier) AndDeleteMany(objs []interface{}) *DeleteModifier {
	m.Request.Controls.Operation = "delete"
	m.Request.Controls.Multiplicity = "single"

	s := reflect.ValueOf(objs)
	var interfaceList []map[string]interface{}
	interfaceList = make([]map[string]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		obj := s.Index(i).Interface()
		v := reflect.ValueOf(obj)
		k := v.Kind()
		var newMap map[string]interface{}

		if k != reflect.Map {
			newMap = structs.Map(obj)
		} else {
			newMap = obj.(map[string]interface{})
		}

		interfaceList[i] = newMap
	}
	m.Request.Body.Objects = interfaceList
	return m
}

func (m *DeleteModifier) WithKeyField(field string) *DeleteModifier {
	m.Request.Body.Parameters = messaging.ObjectParameters{}
	m.Request.Body.Parameters.KeyProperty = field
	return m
}

func (m *DeleteModifier) Ok() (err error) {
	dispatcher := processors.Dispatcher{}
	response := dispatcher.Dispatch(m.Request)

	if !response.IsSuccess {
		if response.Message == "" {
			err = errors.New("Error Deleting Object! : Undefined Error!")
		} else {
			err = errors.New(response.Message)
		}
	}
	return
}

func NewDeleteModifier(request *messaging.ObjectRequest) *DeleteModifier {
	modifier := DeleteModifier{Request: request}
	modifier.Request = request
	return &modifier
}
