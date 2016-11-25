package client

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/processors"
	"duov6.com/objectstore/repositories"
)

type GetModifier struct {
	Request *messaging.ObjectRequest
}

func (m *GetModifier) ByUniqueKey(key string) *GetModifier {
	m.Request.Controls.Operation = "read-key"
	m.Request.Controls.Id = key
	return m
}

func (m *GetModifier) BySearching(keyword string) *GetModifier {
	m.Request.Controls.Operation = "read-keyword"
	m.Request.Body = messaging.RequestBody{}
	m.Request.Body.Query = messaging.Query{}
	m.Request.Body.Query.Parameters = keyword
	return m
}

func (m *GetModifier) ByQuerying(query string) *GetModifier {
	m.Request.Controls.Operation = "read-filter"
	m.Request.Body = messaging.RequestBody{}
	m.Request.Body.Query = messaging.Query{}
	m.Request.Body.Query.Parameters = query
	m.Request.Body.Query.Type = "Query"
	return m
}

func (m *GetModifier) All() *GetModifier {
	return m
}

func (m *GetModifier) Aggregate(key string) *GetModifier {
	return m
}

func (m *GetModifier) Ok() (output []byte, err string) {
	dispatcher := processors.Dispatcher{}
	var repResponse repositories.RepositoryResponse = dispatcher.Dispatch(m.Request)
	output = repResponse.Body

	if !repResponse.IsSuccess {
		err = repResponse.Message
	}
	if len(output) == 2 {
		output = nil
	}

	return
}

func NewGetModifier(request *messaging.ObjectRequest) *GetModifier {
	modifier := GetModifier{Request: request}
	modifier.Request = request
	return &modifier
}
