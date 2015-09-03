package client

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/processors"
	//"duov6.com/objectstore/repositories"
	"github.com/fatih/structs"
	//"strconv"
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
	m.Request.Controls.Operation = "delete"
	m.Request.Controls.Multiplicity = "single"
	bodyMap := structs.Map(obj)
	key := bodyMap[m.Request.Body.Parameters.KeyProperty].(string)
	m.Request.Controls.Id = key
	return m
}

func (m *DeleteModifier) WithKeyField(field string) *DeleteModifier {
	m.Request.Body.Parameters = messaging.ObjectParameters{}
	m.Request.Body.Parameters.KeyProperty = field
	return m
}

func (m *DeleteModifier) Ok() {
	dispatcher := processors.Dispatcher{}
	//var repResponse repositories.RepositoryResponse = dispatcher.Dispatch(m.Request)
	dispatcher.Dispatch(m.Request)
	return
}

func NewDeleteModifier(request *messaging.ObjectRequest) *DeleteModifier {
	modifier := DeleteModifier{Request: request}
	modifier.Request = request
	return &modifier
}
