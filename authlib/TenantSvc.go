package authlib

import (
	"encoding/json"

	"duov6.com/common"
	notifier "duov6.com/duonotifier/client"
	"duov6.com/gorest"
	"duov6.com/session"
	"duov6.com/term"
	"fmt"
)

type TenantSvc struct {
	gorest.RestService
	autherized                gorest.EndPoint `method:"GET" path:"/tenant/Autherized/{TenantID:string}" output:"TenantAutherized"`
	getTenant                 gorest.EndPoint `method:"GET" path:"/tenant/GetTenant/{TenantID:string}" output:"Tenant"`
	acceptRequest             gorest.EndPoint `method:"GET" path:"/tenant/AcceptRequest/{email:string}/{RequestToken:string}" output:"bool"`
	getTenants                gorest.EndPoint `method:"GET" path:"/tenant/GetTenants/{securityToken:string}" output:"[]TenantMinimum"`
	getSampleTenantForm       gorest.EndPoint `method:"GET" path:"/tenant/GetSampleTenantForm/" output:"Tenant"`
	inviteUser                gorest.EndPoint `method:"POST" path:"/tenant/InviteUser/" postdata:"[]InviteUsers"`
	createTenant              gorest.EndPoint `method:"POST" path:"/tenant/CreateTenant/" postdata:"Tenant"`
	tenantUpgrade             gorest.EndPoint `method:"POST" path:"/tenant/TenantUpgrad/" postdata:"map[string]string"`
	searchTenants             gorest.EndPoint `method:"GET" path:"/tenant/SearchTenants/{SearchString:string}/{pagesize:int}/{startPoint:int}" output:"[]Tenant"`
	subciribe                 gorest.EndPoint `method:"GET" path:"/tenant/Subciribe/{TenantID:string}" output:"bool"`
	getUsers                  gorest.EndPoint `method:"GET" path:"/tenant/GetUsers/{TenantID:string}" output:"[]string"`
	addUser                   gorest.EndPoint `method:"GET" path:"/tenant/AddUser/{email:string}/{level:string}" output:"bool"`
	removeUser                gorest.EndPoint `method:"GET" path:"/tenant/RemoveUser/{email:string}" output:"bool"`
	tranferAdmin              gorest.EndPoint `method:"GET" path:"/tenant/TranferAdmin/{email:string}" output:"bool"`
	getPendingTenantRequest   gorest.EndPoint `method:"GET" path:"/tenant/GetPendingTenantRequest/" output:"[]PendingUserRequest"`
	getMyPendingTenantRequest gorest.EndPoint `method:"GET" path:"/tenant/GetMyPendingTenantRequest/" output:"[]PendingUserRequest"`
	getDefaultTenant          gorest.EndPoint `method:"GET" path:"/tenant/GetDefaultTenant/{UserID:string}" output:"Tenant"`
	setDefaultTenant          gorest.EndPoint `method:"GET" path:"/tenant/SetDefaultTenant/{UserID:string}/{TenantID:string}" output:"bool"`
	getTenantAdmin            gorest.EndPoint `method:"GET" path:"/tenant/GetTenantAdmin/{TenantID:string}" output:"[]InviteUsers"`
}

func (T TenantSvc) GetTenantAdmin(TenantID string) []InviteUsers {
	//Get Tenant Admin by TenantID
	term.Write("Executing Method : Get Tenant Admin", term.Blank)

	_, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		th := TenantHandler{}
		adminUsers := make([]InviteUsers, 0)
		admins := th.GetTenantAdmin(TenantID)
		for _, admin := range admins {
			singleAdmin := InviteUsers{}
			singleAdmin.UserID = admin["UserID"]
			singleAdmin.Name = admin["Name"]
			singleAdmin.Email = admin["EmailAddress"]
			singleAdmin.SecurityLevel = "admin"
			adminUsers = append(adminUsers, singleAdmin)
		}
		return adminUsers
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		emptyArray := make([]InviteUsers, 0)
		return emptyArray
	}
}

func (T TenantSvc) GetDefaultTenant(UserID string) Tenant {
	//Get Default tenant for a user
	term.Write("Executing Method : Get Defaut Tenant", term.Blank)

	_, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		th := TenantHandler{}
		return th.GetDefaultTenant(UserID)
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		return Tenant{}
	}
}

func (T TenantSvc) SetDefaultTenant(UserID string, TenantID string) bool {
	//Set a User's Default Tenant
	term.Write("Executing Method : Set Default Tenant", term.Blank)

	_, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		th := TenantHandler{}
		return th.SetDefaultTenant(UserID, TenantID)
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		return false
	}
}

func (T TenantSvc) CreateTenant(t Tenant) {
	//Create a new Tenant
	term.Write("Executing Method : Create Tenant", term.Blank)

	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		th := TenantHandler{}
		b, _ := json.Marshal(th.CreateTenant(t, user, false))
		T.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)

	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		return
	}
}

func (T TenantSvc) TenantUpgrade(Otherdata map[string]string) {
	//Upgrade Tenant
	term.Write("Executing Method : Tenant Upgrade", term.Blank)

	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		th := TenantHandler{}
		t, err := th.UpgradPackage(user, Otherdata)

		if err == "" {
			b, _ := json.Marshal(t)
			T.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
			return
		} else {
			T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(err))
			return
		}

	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		return
	}

}

func (T TenantSvc) TranferAdmin(email string) bool {
	//Transfer Admin of Tenant to Email Address in params.
	term.Write("Executing Method : Transfer Admin", term.Blank)

	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		auth := AuthHandler{}
		u, err := auth.GetUser(email)
		if err == "" {
			th := TenantHandler{}
			return th.TransferAdmin(user, u.UserID)
		} else {
			return false
		}
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		return false
	}
}

func (T TenantSvc) Autherized(TenantID string) (outCrt TenantAutherized) {
	//Check if User is Authorized for Tenant
	term.Write("Executing Method : Autherized (Check if user is authorized for tenant)", term.Blank)

	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		th := TenantHandler{}
		b, d := th.Autherized(TenantID, user)

		if b {
			outCrt = d
			return d
		} else {
			T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Tenant ID " + TenantID + " not Atherized")))
			return
		}
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		return
	}
}

func (T TenantSvc) GetTenant(TenantID string) Tenant {
	//Get Tenant Information
	term.Write("Executing Method : Get Tenant (Tenant Information)", term.Blank)

	_, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		th := TenantHandler{}
		return th.GetTenant(TenantID)
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		return Tenant{}
	}
}

func (T TenantSvc) GetUsers(TenantID string) []string {
	//Get Users inside a Tenant
	term.Write("Executing Method : Get Users (Inside a tenant)", term.Blank)

	u, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")

	if error == "" {
		th := TenantHandler{}
		return th.GetUsersForTenant(u, TenantID)
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		return []string{}
	}
}

func (T TenantSvc) InviteUser(users []InviteUsers) {
	//Invite User to Tenant
	term.Write("Executing Method : Invite User (To Tenant)", term.Blank)

	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		th := TenantHandler{}
		th.AddUserToTenant(user, users)
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		return
	}
}

func (T TenantSvc) AddUser(email, level string) bool {
	//Add User to Tenant
	term.Write("Executing Method : Add User (To Tenant)", term.Blank)

	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		auth := AuthHandler{}
		a, err := auth.GetUser(email)
		emptyUser := User{}
		if err == "" && a != emptyUser {
			th := TenantHandler{}
			t := th.GetTenant(user.Domain)
			var inputParams map[string]string
			inputParams = make(map[string]string)
			inputParams["@@EMAIL@@"] = email
			inputParams["@@INVEMAIL@@"] = user.Email
			inputParams["@@NAME@@"] = user.Name
			inputParams["@@DOMAIN@@"] = user.Domain

			fmt.Println("-----------------------------------------------")
			fmt.Println("Tenant Invitation Existing ..... ")
			fmt.Println(inputParams)
			fmt.Println("-----------------------------------------------")

			go notifier.Notify("ignore", "tenant_invitation_existing", email, inputParams, nil)

			if th.IncreaseTenantCountInRatingEngine(user.Domain, T.Context.Request().Header.Get("Securitytoken")) {
				th.AddUsersToTenant(user.Domain, t.Name, a.UserID, level)
				return true
			} else {
				return false
			}
		} else {
			tmp := tempRequestGenerator{}
			th := TenantHandler{}
			t := th.GetTenant(user.Domain)
			o := make(map[string]string)
			o["process"] = "tenant_invitation"
			o["email"] = email
			o["invitedUserID"] = user.UserID
			o["name"] = user.Name
			o["domain"] = user.Domain
			o["fromuseremail"] = user.Email
			o["tname"] = t.Name
			o["level"] = level
			//o["userid"] = a.UserID
			code := tmp.GenerateRequestCode(o)
			var inputParams map[string]string
			inputParams = make(map[string]string)
			inputParams["@@EMAIL@@"] = email
			inputParams["@@INVEMAIL@@"] = user.Email
			inputParams["@@NAME@@"] = user.Name
			inputParams["@@DOMAIN@@"] = user.Domain
			inputParams["@@CODE@@"] = code

			fmt.Println("-----------------------------------------------")
			fmt.Println("Tenant Invitation New User ..... ")
			fmt.Println(inputParams)
			fmt.Println("-----------------------------------------------")

			go notifier.Notify("ignore", "tenant_invitation", email, inputParams, nil)
			return true
		}
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		return false
	}
}

func (T TenantSvc) RemoveUser(email string) bool {
	//Remove User from Tenant
	term.Write("Executing Method : Remove User (From Tenant)", term.Blank)

	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		auth := AuthHandler{}
		u, err := auth.GetUser(email)
		if err == "" {
			if user.Email == email {
				T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Cant remove yourself from the.")))
				return false
			}
			th := TenantHandler{}
			//_, p := th.Autherized(user.Domain, user)
			//if p.SecurityLevel == "admin" {
			//t := th.GetTenant(user.Domain)
			//th.AddUsersToTenant(user.Domain, t.Name, a.UserID, "level")
			//th := TenantHandler{}
			return th.RemoveUserFromTenant(u.UserID, user.Domain)
			//return true
			//} else {
			//T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Need to have Admin access to tenant to add user"))
			//return false
			//}

		} else {
			return false
		}
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		return false
	}
}

func (T TenantSvc) AcceptRequest(email, RequestToken string) bool {
	//Accept Request To Join to Tenant
	term.Write("Executing Method : Accept Request (To join to tenant)", term.Blank)

	//fmt.Println(T.Context.Request().Header.Get("Securitytoken"))
	//user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	//if error == "" {

	// if T.Context.Request().Header.Get("Securitytoken") == "" {
	// 	term.Write("Error : No SecurityToken found in Header", term.Error)
	// 	return false
	// }

	tmp := tempRequestGenerator{}

	o, _ := tmp.GetRequestCode(RequestToken)
	th := TenantHandler{}
	term.Write(o, term.Blank)
	term.Write(o["process"], term.Debug)

	inputParams := make(map[string]string)
	//changed domain to Domain
	switch o["process"] {
	case "tenant_invitation":
		auth := AuthHandler{}
		a, err := auth.GetUser(o["email"])
		if err == "" {
			if th.IncreaseTenantCountInRatingEngine(o["Domain"], "ignore") {
				//if th.IncreaseTenantCountInRatingEngine(o["domain"], T.Context.Request().Header.Get("Securitytoken")) {
				th.AddUsersToTenant(o["Domain"], o["tname"], a.UserID, o["level"])
				inputParams["@@CNAME@@"] = a.Name
				inputParams["@@DOMAIN@@"] = o["Domain"]
				inputParams["@@TENANTID@@"] = o["TenantID"]
				go notifier.Notify("ignore", "tenant_accepted_success", email, inputParams, nil)
				go notifier.Notify("ignore", "tenant_invitation_added_success", email, inputParams, nil)
				return true
			} else {
				return false
			}
		} else {
			T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Email not registered.")))
			return false
		}
		break
	case "tenant_useradd":

		auth := AuthHandler{}
		a, err := auth.GetUser(o["email"])
		if err == "" {
			if th.IncreaseTenantCountInRatingEngine(o["Domain"], T.Context.Request().Header.Get("Securitytoken")) {
				th.AddUsersToTenant(o["TenantID"], o["tname"], a.UserID, o["level"])
				th.RemovePendingRequest(o["TenantID"], a.EmailAddress)
				inputParams["@@CNAME@@"] = a.Name
				inputParams["@@DOMAIN@@"] = o["tname"]
				inputParams["@@TENANTID@@"] = o["TenantID"]
				go notifier.Notify("ignore", "tenant_accepted_success", email, inputParams, nil)
				//go notifier.Notify("ignore", "tenant_invitation_added_success", email, inputParams, nil)
				return true
			} else {
				return false
			}
		} else {
			T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Email not registered.")))
			return false
		}
		break
	default:
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Unatherized token")))
		return false
		break
	}
	//return th.AcceptRequest(user, securityLevel, RequestToken, accept)
	//} else {

	//}
	T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Unatherized token")))
	return false
}

func (T TenantSvc) GetTenants(securityToken string) []TenantMinimum {
	//Get Tenants for a user.
	term.Write("Executing Method : Get Tenants (For a User)", term.Blank)

	tns := []TenantMinimum{}
	user, error := session.GetSession(securityToken, "Nil")

	if error == "" {
		th := TenantHandler{}
		return th.GetTenantsForUser(user.UserID)
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
	}
	return tns
}

func (T TenantSvc) SearchTenants(SearchString string, pageSize, startPoint int) []Tenant {
	//Search Tenants
	term.Write("Executing Method : Search Tenants", term.Blank)
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
	//Subscribe to a Tenant
	term.Write("Executing Method : Subscribe (To a Tenant) ", term.Blank)

	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		//check for available tenants.. If Tenant ID is there.. Reject else continue
		tenantsForUser := T.GetTenants(T.Context.Request().Header.Get("Securitytoken"))

		for _, tenant := range tenantsForUser {
			if tenant.TenantID == TenantID {
				term.Write(("User : " + user.Email + " is already Subscribed to Tenant : " + TenantID), term.Information)
				return false
			}
		}

		//Request to tenent
		th := TenantHandler{}
		return th.RequestToTenant(user, TenantID)
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("SecurityToken  not Autherized"))
		return false
	}
}

type PendingUserRequest struct {
	UserID   string
	Email    string
	TenantID string
	Name     string
	Code     string
}

func (T TenantSvc) GetPendingTenantRequest() (m []PendingUserRequest) {
	//Get pending tenant requests for a user
	term.Write("Executing Method : Get Pending Tenant Requests (For a User)", term.Blank)

	var tns []PendingUserRequest
	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")

	if error == "" {
		th := TenantHandler{}
		tns, _ = th.GetPendingRequests(user)

	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
	}
	m = tns
	return
}

func (T TenantSvc) GetMyPendingTenantRequest() (m []PendingUserRequest) {
	//Get My pending tenant requests
	term.Write("Executing Method : Get My Pending Tenant Requests", term.Blank)

	var tns []PendingUserRequest
	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")

	if error == "" {
		th := TenantHandler{}
		tns, _ = th.GetMyPendingRequests(user)

	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
	}
	m = tns
	return
}
