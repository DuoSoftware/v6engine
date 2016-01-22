package main

import (
	"duov6.com/gorest"
	"fmt"
	"net/http"
)

type Data struct {
	Object map[string]interface{}
}

type Prasad struct {
	gorest.RestService
	registerTenantUser gorest.EndPoint `method:"POST" path:"/RegisterTenantUser/{SecurityToken:string}/{Code:string}/{ApplicationID:string}/{AppSecret:string}" postdata:"Data"`
	//registerTenantUser gorest.EndPoint `method:"POST" path:"/RegisterTenantUser/" postdata:"Data"`
}

func (A Prasad) RegisterTenantUser(dd Data, SecurityToken string, Code string, ApplicationID string, AppSecret string) {
	fmt.Println(SecurityToken)
	fmt.Println(Code)
	fmt.Println(ApplicationID)
	fmt.Println(AppSecret)
	fmt.Println(dd.Object)
	client.Go("token", "com.duosoftware.customer", "account").StoreObject().WithKeyField("Id").AndStoreOne(tmp).Ok()
}

// func (A Prasad) RegisterTenantUser(dd Data) {
// 	fmt.Println(dd)
// 	fmt.Println(dd.Aa)
// 	fmt.Println(dd.Cc)
// }

func main() {
	gorest.RegisterService(new(Prasad))
	http.Handle("/", gorest.Handle())
	http.ListenAndServe(":8787", nil)
}
