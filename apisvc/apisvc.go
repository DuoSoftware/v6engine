package apisvc

import (
	"duov6.com/api"
	"duov6.com/gorest"
	"encoding/json"
)

type ApiSvc struct {
	gorest.RestService
	get  gorest.EndPoint `method:"GET" path:"/API/Get/{apiname:string}" output:"string"`
	list gorest.EndPoint `method:"GET" path:"/API/List/" postdata:"[]string"`
}

func (A ApiSvc) Get(apiname string) (s string) {
	s = ""
	h := api.ApiHandler{}
	p := []api.Parameters{}
	h.NewDoc("AuthSvc", "6.0 Authendication Documentation sample")
	h.AddMethod(api.Method{"Login", "Method to login", "GET", "/Login/{username:string}/{password:string}/{domain:string}", `{"Name":"AuthSvc","Description":"New authedication machanisum","Methods":[{"Name":"Login","Description":"Loging Methded to login customer","Method":"/Login/{username:string}/{password:string}/{domain:string}","URI":"GET","OutPutBody":"output","OutPutType":"AuthCertificate","InParameters":[]}]}`, "AuthCertificate", p})
	b, _ := json.Marshal(h.Document)
	A.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
	return

}

func (A ApiSvc) List() []string {
	return []string{"Auth Service", "Sample Service", "Load Balance"}
}
