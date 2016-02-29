package client

import (
	"duov6.com/documentcache"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/processors"
	"duov6.com/objectstore/repositories"
	"encoding/json"
	//"fmt"
)

type GetModifier struct {
	Request *messaging.ObjectRequest
}

func (m *GetModifier) ByUniqueKey(key string) *GetModifier {
	m.Request.Controls.Operation = "read-key"
	m.Request.Controls.Id = key
	return m
}

func (m *GetModifier) ByUniqueKeyCache(key string, ttl int) *GetModifier {
	m.Request.Controls.Operation = "read-key-cache"
	m.Request.Controls.Id = key

	var cacheData map[string]interface{}
	cacheData = make(map[string]interface{})
	cacheData["ttl"] = ttl
	m.Request.Body.Object = cacheData
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
	return m
}

func (m *GetModifier) All() *GetModifier {
	return m
}

func (m *GetModifier) Aggregate(key string) *GetModifier {
	return m
}

func (m *GetModifier) Ok() (output []byte, err string) {
	if m.Request.Controls.Operation != "read-key-cache" {
		dispatcher := processors.Dispatcher{}
		var repResponse repositories.RepositoryResponse = dispatcher.Dispatch(m.Request)
		output = repResponse.Body

	} else {
		//using Cache
		//first read from cache
		key := m.Request.Controls.Id

		cacheData := documentcache.Fetch(key)

		if cacheData != nil {
			cacheDataByte, _ := json.Marshal(cacheData)
			output = cacheDataByte
		} else {

			//if not in cache.... get from objectstore..
			dispatcher := processors.Dispatcher{}
			var repResponse repositories.RepositoryResponse = dispatcher.Dispatch(m.Request)
			output = repResponse.Body
			//write to cache and return
			var object interface{}
			_ = json.Unmarshal(repResponse.Body, &object)
			ttl := m.Request.Body.Object["ttl"].(int)
			_ = documentcache.Store(key, ttl, object)
		}
	}

	if len(output) == 2 {
		//	err = "ERROR"
		output = nil
	}

	return
}

func NewGetModifier(request *messaging.ObjectRequest) *GetModifier {
	modifier := GetModifier{Request: request}
	modifier.Request = request
	return &modifier
}
