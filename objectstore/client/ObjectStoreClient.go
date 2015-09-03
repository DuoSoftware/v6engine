package client

import (
	"duov6.com/objectstore/configuration"
	"duov6.com/objectstore/messaging"
)

type ObjectStoreClient struct {
	Request *messaging.ObjectRequest
}

func (o *ObjectStoreClient) DeleteObject() *DeleteModifier {
	o.Request.Controls.Multiplicity = "single"
	return NewDeleteModifier(o.Request)
}

func (o *ObjectStoreClient) GetOne() *GetModifier {
	o.Request.Controls.Multiplicity = "single"
	return NewGetModifier(o.Request)
}

func (o *ObjectStoreClient) GetMany() *GetModifier {
	o.Request.Controls.Multiplicity = "multiple"
	return NewGetModifier(o.Request)
}

func (o *ObjectStoreClient) StoreObject() *StoreModifier {
	return NewStoreModifier(o.Request)
}

func (o *ObjectStoreClient) StoreObjectWithOperation(operation string) *StoreModifier {
	return NewStoreModifierWithOperation(o.Request, operation)
}

func Go(securityToken string, namespace string, class string) *ObjectStoreClient {
	client := ObjectStoreClient{}
	requestObject := getObjectRequest(securityToken, namespace, class)
	client.Request = &requestObject
	return &client
}

func getObjectRequest(headerToken string, headerNamespace string, headerClass string) (objectRequest messaging.ObjectRequest) {
	objectRequest.Controls = messaging.RequestControls{SecurityToken: headerToken, Namespace: headerNamespace, Class: headerClass}
	configObject := configuration.ConfigurationManager{}.Get(headerToken, headerNamespace, headerClass)
	objectRequest.Configuration = configObject

	objectRequest.IsLogEnabled = true
	var initialSlice []string
	initialSlice = make([]string, 0)
	objectRequest.MessageStack = initialSlice

	//objectRequest.IsLogEnabled = false

	var extraMap map[string]interface{}
	extraMap = make(map[string]interface{})
	objectRequest.Extras = extraMap
	return
}

//for pushing EXTRA parameters

func GoExtra(securityToken string, namespace string, class string, extra map[string]interface{}) *ObjectStoreClient {
	client := ObjectStoreClient{}
	requestObject := getObjectRequestExtra(securityToken, namespace, class, extra)
	client.Request = &requestObject
	return &client
}

func getObjectRequestExtra(headerToken string, headerNamespace string, headerClass string, extra map[string]interface{}) (objectRequest messaging.ObjectRequest) {
	objectRequest.Controls = messaging.RequestControls{SecurityToken: headerToken, Namespace: headerNamespace, Class: headerClass}
	configObject := configuration.ConfigurationManager{}.Get(headerToken, headerNamespace, headerClass)
	objectRequest.Configuration = configObject

	objectRequest.IsLogEnabled = true
	var initialSlice []string
	initialSlice = make([]string, 0)
	objectRequest.MessageStack = initialSlice

	//objectRequest.IsLogEnabled = false

	// var extraMap map[string]interface{}
	// extraMap = make(map[string]interface{})
	// objectRequest.Extras = extraMap
	objectRequest.Extras = extra
	return
}
