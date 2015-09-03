package applib

import (
	//"duov6.com/authlib"
	"duov6.com/gorest"
	"duov6.com/term"
	//"duov"
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
	OAuthURIs     string
	AppICON       string
	ScreenShoots  []string
	OtherData     map[string]interface{}
}

type AppSvc struct {
	gorest.RestService
	get gorest.EndPoint `method:"GET" path:"/Application/Get/{ApplicationID:string}" output:"Application"`
	add gorest.EndPoint `method:"POST" path:"/Application/Add/" postdata:"Application"`
}

func (app AppSvc) Get(ApplicationID string) (a Application) {
	scurityToken := app.Context.Request().Header.Get("SecurityToken")
	term.Write("Get App for SecurityToken "+scurityToken, term.Debug)
	var h Apphanler
	a, err := h.Get(ApplicationID, scurityToken)
	if err != "" {

		app.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(err))
		return
	}

	return
}

func (app AppSvc) Add(a Application) {
	scurityToken := app.Context.Request().Header.Get("SecurityToken")
	term.Write("Add App for SecurityToken "+scurityToken, term.Debug)

	var h Apphanler

	a, err := h.Add(a, scurityToken)
	if err != "" {
		app.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(err))
		return
	}
	b, _ := json.Marshal(a)
	app.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
	return
	//return a
}
