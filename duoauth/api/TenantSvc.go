package api

import (
	// "duov6.com/common"
	// notifier "duov6.com/duonotifier/client"
	// "duov6.com/objectstore/client"
	// "duov6.com/session"
	"duov6.com/term"
	"encoding/json"
	// "fmt"
	"github.com/SiyaDlamini/gorest"
	// "strconv"
	// "strings"
)

type TenantSvc struct {
	gorest.RestService
	getAllTenants        gorest.EndPoint `method:"GET" path:"/tenants" output:"AuthResponse"`
	getTenant            gorest.EndPoint `method:"GET" path:"/tenants/{tid:string}" output:"AuthResponse"`
	createTenant         gorest.EndPoint `method:"POST" path:"/tenants" postdata:"Tenant"`
	updateTenant         gorest.EndPoint `method:"PUT" path:"/tenants" postdata:"Tenant"`
	deleteTenant         gorest.EndPoint `method:"DELETE" path:"/tenants/{tid:string}"`
	getTenantUsers       gorest.EndPoint `method:"GET" path:"/tenants/{tid:string}/users" output:"AuthResponse"`
	getUserTenants       gorest.EndPoint `method:"GET" path:"/tenants/{userid:string}/getall" output:"AuthResponse"`
	deleteUserFromTenant gorest.EndPoint `method:"DELETE" path:"/tenants/{tid:string}/removeuser/{Email:string}"`
	getUserDefaultTenant gorest.EndPoint `method:"GET" path:"/tenants/{userid:string}/getdefault" output:"AuthResponse"`
	setUserDefaultTenant gorest.EndPoint `method:"GET" path:"/tenants/{userid:string}/setdefault/{tid:string}" output:"AuthResponse"`
}

func (T TenantSvc) GetAllTenants() AuthResponse {
	term.Write("Executing Method : Get All Tenants", term.Blank)
	response := AuthResponse{}
	//id_token := T.Context.Request().Header.Get("Securitytoken")
	return response
}

func (T TenantSvc) GetTenant(tid string) AuthResponse {
	term.Write("Executing Method : Get Tenant Info", term.Blank)
	response := AuthResponse{}
	return response
}

func (T TenantSvc) CreateTenant(tenant Tenant) {
	term.Write("Executing Method : Create a tenant.", term.Blank)
	response := AuthResponse{}
	b, _ := json.Marshal(response)
	T.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
}

func (T TenantSvc) UpdateTenant(tenant Tenant) {
	term.Write("Executing Method : Update Tenant.", term.Blank)
	response := AuthResponse{}
	b, _ := json.Marshal(response)
	T.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
}

func (T TenantSvc) DeleteTenant(tid string) {
	term.Write("Executing Method : Delete Tenant.", term.Blank)
	response := AuthResponse{}
	b, _ := json.Marshal(response)
	T.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
}

func (T TenantSvc) GetTenantUsers(tid string) AuthResponse {
	term.Write("Executing Method : Get Tenant Users", term.Blank)
	response := AuthResponse{}
	return response
}

func (T TenantSvc) GetUserTenants(userid string) AuthResponse {
	term.Write("Executing Method : Get Tenant Users", term.Blank)
	response := AuthResponse{}
	return response
}

func (T TenantSvc) DeleteUserFromTenant(tid, Email string) {
	term.Write("Executing Method : Delete Tenant.", term.Blank)
	response := AuthResponse{}
	b, _ := json.Marshal(response)
	T.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
}

func (T TenantSvc) GetUserDefaultTenant(userid string) AuthResponse {
	term.Write("Executing Method : Get users default tenant", term.Blank)
	response := AuthResponse{}
	return response
}

func (T TenantSvc) SetUserDefaultTenant(userid, tid string) AuthResponse {
	term.Write("Executing Method : Set users default tenant", term.Blank)
	response := AuthResponse{}
	return response
}
