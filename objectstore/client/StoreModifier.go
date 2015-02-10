package client

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/processors"
	"duov6.com/objectstore/repositories"
	"fmt"
	"github.com/fatih/structs"
	//"strconv"
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
	bodyMap := structs.Map(obj)
	m.Request.Body.Object = bodyMap

	return m
}

func (m *StoreModifier) AndStoreMany(objs []interface{}) *StoreModifier {
	m.Request.Controls.Multiplicity = "multiple"

	var interfaceList []map[string]interface{}
	interfaceList = make([]map[string]interface{}, len(objs))

	for index, element := range objs {
		interfaceList[index] = structs.Map(element)
	}

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

func (m *StoreModifier) FileOk() {
	if m.Request.Controls.Multiplicity == "single" {
		m.Request.Controls.Id = m.Request.Body.Object[m.Request.Body.Parameters.KeyProperty].(string)
	}

	dispatcher := processors.Dispatcher{}

	response := dispatcher.Dispatch(m.Request)

	fmt.Println(response.IsSuccess)
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
