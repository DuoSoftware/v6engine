package api

import (
	"duov6.com/common"
	// notifier "duov6.com/duonotifier/client"
	// "duov6.com/objectstore/client"
	// "duov6.com/session"
	"duov6.com/duoauth/azureapi"
	"duov6.com/term"
	"encoding/json"
	"fmt"
	"github.com/SiyaDlamini/gorest"
	"net/url"
	// "strconv"
	// "strings"
	"errors"
)

type TenantSvc struct {
	gorest.RestService
	getAllTenants        gorest.EndPoint `method:"GET" path:"/tenants" output:"AuthResponse"`
	getTenant            gorest.EndPoint `method:"GET" path:"/tenants/{tid:string}" output:"AuthResponse"`
	createTenant         gorest.EndPoint `method:"POST" path:"/tenants" postdata:"Tenant"`
	updateTenant         gorest.EndPoint `method:"PUT" path:"/tenants" postdata:"Tenant"`
	deleteTenant         gorest.EndPoint `method:"DELETE" path:"/tenants/{tid:string}"`
	getTenantUsers       gorest.EndPoint `method:"GET" path:"/tenants/{tid:string}/users" output:"AuthResponse"`
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

	access_token, err := azureapi.GetGraphApiToken()
	if err == nil {
		fmt.Println("1")
		//token is good. proceed.
		graphUrl := "https://graph.windows.net/smoothflowio.onmicrosoft.com/groups?api-version=1.6&$filter=" + url.QueryEscape("displayName eq '"+tid+"'")
		headers := make(map[string]string)
		headers["Authorization"] = "Bearer " + access_token
		headers["Content-Type"] = "application/json"

		var body []byte
		err, body = common.HTTP_GET(graphUrl, headers, false)
		if err == nil {
			data := make(map[string]interface{})
			_ = json.Unmarshal(body, &data)

			if len(data["value"].([]interface{})) > 0 {
				//tenant found.
				descriptionString := (((data["value"].([]interface{}))[0]).(map[string]interface{}))["description"].(string)
				tenant := Tenant{}
				if err = json.Unmarshal([]byte(descriptionString), &tenant); err == nil {
					tenant.TenantID = tid
					response.Status = true
					response.Message = "Successfully retrieved tenant information."
					response.Data = tenant
				}
			} else {
				//tenant not found
				err = errors.New("Tenant not found.")
			}
		}
	}

	if err != nil {
		response.Status = false
		response.Message = err.Error()
	}

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
