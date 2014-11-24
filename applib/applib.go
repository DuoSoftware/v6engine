package applib

import (
	"code.google.com/p/gorest"
	//"duov6.com/authlib"
	///"duov6.com/term"
	"encoding/json"
)

func NewApphanler() AppSvc {
	var apphdl AppSvc
	return apphdl
}

type Application struct {
	ApplicationID string
	SecretKey     string
	Name          string
	Description   string
	AppType       string
	AppUri        string
	OtherData     map[string]interface{}
}

type AppSvc struct {
	gorest.RestService
	get gorest.EndPoint `method:"GET" path:"/Application/Get/{ApplicationID:string}" output:"Application"`
	add gorest.EndPoint `method:"POST" path:"/Application/Add/" postdata:"Application"`
}

func (app AppSvc) Get(ApplicationID string) (a Application) {
	var h Apphanler
	a, err := h.Get(ApplicationID)
	if err != "" {

		app.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(err))
		return
	}

	return
}

func (app AppSvc) Add(a Application) {
	var h Apphanler
	a, err := h.Add(a)
	if err != "" {
		app.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(err))
		return
	}
	b, _ := json.Marshal(a)
	app.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
	return
	//return a
}
