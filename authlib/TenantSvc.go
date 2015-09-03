package authlib

import (
	"duov6.com/gorest"
	"duov6.com/session"
	"encoding/json"
	//"fmt"
)

type TenantSvc struct {
	gorest.RestService
	autherized          gorest.EndPoint `method:"GET" path:"/tenant/Autherized/{TenantID:string}" output:"TenantAutherized"`
	getTenant           gorest.EndPoint `method:"GET" path:"/tenant/GetTenant/{TenantID:string}" output:"Tenant"`
	acceptRequest       gorest.EndPoint `method:"GET" path:"/tenant/AcceptRequest/{securityLevel:string}/{RequestToken:string}/{accept:bool}" output:"bool"`
	getTenants          gorest.EndPoint `method:"GET" path:"/tenant/GetTenants/{securityToken:string}" output:"[]TenantMinimum"`
	getSampleTenantForm gorest.EndPoint `method:"GET" path:"/tenant/GetSampleTenantForm/" output:"Tenant"`
	inviteUser          gorest.EndPoint `method:"POST" path:"/tenant/InviteUser/" postdata:"[]InviteUsers"`
	createTenant        gorest.EndPoint `method:"POST" path:"/tenant/CreateTenant/" postdata:"Tenant"`
	searchTenants       gorest.EndPoint `method:"GET" path:"/tenant/SearchTenants/{SearchString:string}/{pagesize:int}/{startPoint:int}" output:"[]Tenant"`
	subciribe           gorest.EndPoint `method:"GET" path:"/tenant/Subciribe/{TenantID:string}" output:"bool"`
}

func (T TenantSvc) CreateTenant(t Tenant) {
	//fmt.Println(T.Context.Request().Header["SecurityToken"])
	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		th := TenantHandler{}
		b, _ := json.Marshal(th.CreateTenant(t, user, false))
		T.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)

	}
}

func (T TenantSvc) Autherized(TenantID string) (outCrt TenantAutherized) {
	//fmt.Println(T.Context.Request().Header["SecurityToken"])
	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	//outCrt = TenantAutherize{}

	if error == "" {
		th := TenantHandler{}
		b, d := th.Autherized(TenantID, user)

		if b {
			outCrt = d
			return d
		} else {
			T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Tenant ID " + TenantID + " not Atherized"))
			return
		}
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("SecurityToken  not Autherized"))
		return
	}
}

func (T TenantSvc) GetTenant(TenantID string) Tenant {
	//fmt.Println(T.Context.Request().Header.Get("Securitytoken"))
	_, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	//outCrt = TenantAutherize{}

	if error == "" {
		th := TenantHandler{}
		return th.GetTenant(TenantID)
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("SecurityToken  not Autherized"))
		return Tenant{}
	}
}

func (T TenantSvc) InviteUser(users []InviteUsers) {
	//fmt.Println(T.Context.Request().Header.Get("Securitytoken"))
	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		th := TenantHandler{}
		th.AddUserToTenant(user, users)
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("SecurityToken  not Autherized"))
		return
	}
}

func (T TenantSvc) AcceptRequest(securityLevel, RequestToken string, accept bool) bool {
	//fmt.Println(T.Context.Request().Header.Get("Securitytoken"))
	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		th := TenantHandler{}
		return th.AcceptRequest(user, securityLevel, RequestToken, accept)
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("SecurityToken  not Autherized"))
		return false
	}
}

func (T TenantSvc) GetTenants(securityToken string) []TenantMinimum {
	//fmt.Println()
	tns := []TenantMinimum{}
	user, error := session.GetSession(securityToken, "Nil")
	//outCrt = TenantAutherize{}

	if error == "" {
		th := TenantHandler{}
		return th.GetTenantsForUser(user.UserID)
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("SecurityToken  not Autherized"))
	}
	return tns
}

func (T TenantSvc) SearchTenants(SearchString string, pageSize, startPoint int) []Tenant {
	th := TenantHandler{}

	return th.SearchTenants(SearchString, pageSize, startPoint)
}

func (T TenantSvc) GetSampleTenantForm() Tenant {
	var t Tenant
	t = Tenant{}
	t.Name = "Sample Tenant"
	t.OtherData = make(map[string]string)
	t.OtherData["CompanyName"] = "DuoSoftware Pvt Ltd"
	t.OtherData["SampleAttributs"] = "Values"
	t.Private = true
	t.Statistic = make(map[string]string)
	t.Statistic["NumberOfUsers"] = "10"
	t.Statistic["DataUp"] = "1GB"
	t.Statistic["DataDown"] = "1GB"
	t.Shell = "Shell"
	t.TenantID = "smapletenat.duoworld.info"
	return t
}

func (T TenantSvc) Subciribe(TenantID string) bool {
	//users :=InviteUsers[]{"Sss",sss,sss}

	//fmt.Println(T.Context.Request().Header.Get("Securitytoken"))
	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		th := TenantHandler{}
		return th.RequestToTenant(user, TenantID)
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("SecurityToken  not Autherized"))
		return
	}
}
