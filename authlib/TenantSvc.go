package authlib

import (
	"duov6.com/common"
	notifier "duov6.com/duonotifier/client"
	"duov6.com/gorest"
	"duov6.com/objectstore/client"
	"duov6.com/session"
	"duov6.com/term"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type TenantSvc struct {
	gorest.RestService
	autherized                  gorest.EndPoint `method:"GET" path:"/tenant/Autherized/{TenantID:string}" output:"TenantAutherized"`
	getTenant                   gorest.EndPoint `method:"GET" path:"/tenant/GetTenant/{TenantID:string}" output:"Tenant"`
	acceptRequest               gorest.EndPoint `method:"GET" path:"/tenant/AcceptRequest/{email:string}/{RequestToken:string}" output:"bool"`
	getTenants                  gorest.EndPoint `method:"GET" path:"/tenant/GetTenants/{securityToken:string}" output:"[]TenantMinimum"`
	getSampleTenantForm         gorest.EndPoint `method:"GET" path:"/tenant/GetSampleTenantForm/" output:"Tenant"`
	inviteUser                  gorest.EndPoint `method:"POST" path:"/tenant/InviteUser/" postdata:"[]InviteUsers"`
	createTenant                gorest.EndPoint `method:"POST" path:"/tenant/CreateTenant/" postdata:"Tenant"`
	tenantUpgrade               gorest.EndPoint `method:"POST" path:"/tenant/TenantUpgrad/" postdata:"map[string]string"`
	searchTenants               gorest.EndPoint `method:"GET" path:"/tenant/SearchTenants/{SearchString:string}/{pagesize:int}/{startPoint:int}" output:"[]Tenant"`
	subciribe                   gorest.EndPoint `method:"GET" path:"/tenant/Subciribe/{TenantID:string}" output:"bool"`
	getUsers                    gorest.EndPoint `method:"GET" path:"/tenant/GetUsers/{TenantID:string}" output:"[]string"`
	getUserDetails              gorest.EndPoint `method:"GET" path:"/tenant/GetUserDetails/{TenantID:string}" output:"[]User"`
	addUser                     gorest.EndPoint `method:"GET" path:"/tenant/AddUser/{email:string}/{level:string}" output:"bool"`
	removeUser                  gorest.EndPoint `method:"GET" path:"/tenant/RemoveUser/{email:string}" output:"bool"`
	tranferAdmin                gorest.EndPoint `method:"GET" path:"/tenant/TranferAdmin/{email:string}" output:"bool"`
	getPendingTenantRequest     gorest.EndPoint `method:"GET" path:"/tenant/GetPendingTenantRequest/" output:"[]PendingUserRequest"`
	getMyPendingTenantRequest   gorest.EndPoint `method:"GET" path:"/tenant/GetMyPendingTenantRequest/" output:"[]PendingUserRequest"`
	getDefaultTenant            gorest.EndPoint `method:"GET" path:"/tenant/GetDefaultTenant/{UserID:string}" output:"Tenant"`
	setDefaultTenant            gorest.EndPoint `method:"GET" path:"/tenant/SetDefaultTenant/{UserID:string}/{TenantID:string}" output:"bool"`
	getTenantAdmin              gorest.EndPoint `method:"GET" path:"/tenant/GetTenantAdmin/{TenantID:string}" output:"[]InviteUsers"`
	getAllPendingTenantRequests gorest.EndPoint `method:"GET" path:"/tenant/GetAllPendingTenantRequests/" output:"PendingRequests"`
	cancelAddTenantUser         gorest.EndPoint `method:"GET" path:"/tenant/CancelAddUser/{email:string}/" output:"bool"`
	validateCode                gorest.EndPoint `method:"GET" path:"/tenant/verifytoken/{token:string}" output:"bool"`
	initTenantDelete            gorest.EndPoint `method:"GET" path:"/tenant/deleteinit/{tid:string}" output:"string"`
	consentedTenantDelete       gorest.EndPoint `method:"GET" path:"/tenant/delete/{token:string}" output:"string"`

	//support Methods..
	bulkTenantDelete gorest.EndPoint `method:"POST" path:"/tenant/bulkdelete/" postdata:"[]string"`
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

type PendingRequests struct {
	SubscribeRequests []PendingUserRequest
	AddUserRequests   []PendingUserRequest
}

func (T TenantSvc) GetAllPendingTenantRequests() (m PendingRequests) {
	//Get pending tenant requests for a user
	term.Write("Executing Method : Get Pending Tenant Requests (For a User)", term.Blank)

	var tns PendingRequests
	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")

	if error == "" {
		th := TenantHandler{}
		tns.SubscribeRequests, _ = th.GetPendingRequests(user)
		tns.AddUserRequests, _ = th.GetAddUserRequests(user)
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
	}
	m = tns
	return
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
		loweredTenant := t.TenantID
		match, _ := regexp.MatchString("(^app.12thdoor.com|^billing.12thdoor.com|^staging.12thdoor.com|^developer.12thdoor.com|^qa.12thdoor.com)", loweredTenant)
		if match {
			T.ResponseBuilder().SetResponseCode(500).WriteAndOveride([]byte(common.ErrorJson("Tenant ID not allowed.")))
			return
		} else {
			th := TenantHandler{}
			if strings.Contains(strings.ToLower(t.TenantID), "12thdoor.com") {
				//check for tenant count for user
				if len(th.GetTenantsForUser(user.UserID)) >= 5 {
					T.ResponseBuilder().SetResponseCode(500).WriteAndOveride([]byte(common.ErrorJson("Maximum allowed tenant count exceeded.")))
					return
				}
			}

			b, _ := json.Marshal(th.CreateTenant(t, user, false))
			T.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
			return
		}
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

	//_, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	//if error == "" {
	th := TenantHandler{}
	return th.GetTenant(TenantID)
	//} else {
	//	T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
	//	return Tenant{}
	//}
}

func (T TenantSvc) InitTenantDelete(tid string) string {
	//Delete tenant and all associated data
	term.Write("Executing Method : InitTenantDelete)", term.Blank)

	//Check if requester is tenant admin.

	var err error
	var code string

	user, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		auth := AuthHandler{}
		_, uerr := auth.GetUser(user.Email)
		if uerr == "" {
			//check if requester is tenant admin
			isAdmin := false
			admins := T.GetTenantAdmin(user.Domain)
			for _, individualadmin := range admins {
				if individualadmin.Email == user.Email {
					isAdmin = true
					break
				}
			}

			if !isAdmin {
				err = errors.New("Access Denied. Requester must be an admin to perform tenant delete initiation.")
			} else {
				//Check if tenant is available to delete.
				if T.GetTenant(tid).TenantID == "" {
					err = errors.New("Tenant : " + tid + " not found for delete initiation.")
				} else {
					//If all okay send the email.
					tmp := tempRequestGenerator{}
					o := make(map[string]string)
					o["process"] = "InitiateTenantDelete"
					o["email"] = user.Email
					o["invitedUserID"] = "none"
					o["name"] = user.Username
					o["domain"] = tid
					o["fromuseremail"] = "none"
					o["tname"] = tid
					o["level"] = "none"
					o["TenantID"] = tid
					o["inviteeName"] = user.Name
					code = tmp.GenerateRequestCode(o)
					fmt.Println(code)

					var inputParams map[string]string
					inputParams = make(map[string]string)
					inputParams["@@BIZOWNER_NAME@@"] = user.Name
					inputParams["@@BIZOWNER_USERNAME@@"] = user.Username
					inputParams["@@BIZOWNER_EMAIL@@"] = user.Email
					inputParams["@@BIZ_DOMAIN@@"] = tid
					inputParams["@@CODE@@"] = code
					fmt.Println("-----------------------------------------------")
					fmt.Println("Sending initiation email to delete tenant ..... ")
					fmt.Println(inputParams)
					fmt.Println("-----------------------------------------------")

					go notifier.Notify("ignore", "tenant_delete_init", user.Email, inputParams, nil)
				}
			}

		} else {
			err = errors.New(uerr)
		}
	} else {
		err = errors.New(error)
	}

	response := make(map[string]interface{})

	if err != nil {
		response["Status"] = false
		response["Message"] = err.Error()
	} else {
		response["Status"] = true
		response["Message"] = "Concent email sent successfully. Code : " + code
	}

	b, _ := json.Marshal(response)

	return string(b)
}

func (T TenantSvc) ConsentedTenantDelete(token string) string {
	//Delete tenant and all associated data
	term.Write("Executing Method :  Consented Tenant Delete)", term.Blank)

	tmp := tempRequestGenerator{}
	o, _ := tmp.GetRequestCode(token)

	tid := o["TenantID"]
	adminEmail := o["email"]
	adminName := o["inviteeName"]
	adminUserName := o["name"]

	var err error
	isAllDeleted := true

	if len(o["TenantID"]) != 0 {
		//Get All users for tenant
		th := TenantHandler{}
		users := th.GetUsersForTenantInDetail(session.AuthCertificate{}, tid)

		//Remove all users from the tenant.
		for _, user := range users {
			status := th.RemoveUserFromTenant(user.UserID, tid)
			if status {
				//switch the person if default tenant is this tenant
				defT := th.GetDefaultTenant(user.UserID)
				if defT.TenantID == tid {
					//get all tenants for user
					allTenants := th.GetTenantsForUser(user.UserID)
					if len(allTenants) == 0 {
						//when user have no other tenants. delete default tenant so
						//boarding process will begin in next login
						client.Go("ignore", "com.duosoftware.tenant", "defaulttenant").DeleteObject().WithKeyField("UserID").AndDeleteObject(user).Ok()
					} else {
						for _, tenant := range allTenants {
							if tenant.TenantID != tid {
								//Change the default tenant
								th.SetDefaultTenant(user.UserID, tenant.TenantID)
								break
							}
						}
					}
				}
			} else {
				isAllDeleted = false
			}
		}

		//delete tenant
		tObj := make(map[string]interface{})
		tObj["TenantID"] = tid
		err = client.Go("ignore", "com.duosoftware.tenant", "tenants").DeleteObject().WithKeyField("TenantID").AndDeleteOne(tObj).Ok()

		//Delete database
		err = client.Go("ignore", tid, "ignore").DeleteNamespace().Ok()

		//delete token
		obj := make(map[string]interface{})
		obj["id"] = o["id"]
		tmp.Remove(obj)

	} else {
		err = errors.New("Expired token.")
	}

	response := make(map[string]interface{})

	if err != nil {
		response["Status"] = false
		response["Message"] = err.Error()
		response["TenantID"] = "Nil"
	} else {
		response["Status"] = true
		response["TenantID"] = tid

		if isAllDeleted {
			response["Message"] = "All tenant related data successfully removed."
		} else {
			response["Message"] = "All tenant related data successfully removed but failed to remove some users."
		}

		var inputParams map[string]string
		inputParams = make(map[string]string)
		inputParams["@@BIZOWNER_NAME@@"] = adminName
		inputParams["@@BIZOWNER_USERNAME@@"] = adminUserName
		inputParams["@@BIZOWNER_EMAIL@@"] = adminEmail
		inputParams["@@BIZ_DOMAIN@@"] = tid

		fmt.Println("-----------------------------------------------")
		fmt.Println("Sending delete tenant successful email ..... ")
		fmt.Println(inputParams)
		fmt.Println("-----------------------------------------------")

		go notifier.Notify("ignore", "tenant_delete_success", adminEmail, inputParams, nil)

	}

	b, _ := json.Marshal(response)

	return string(b)

}

func (T TenantSvc) BulkTenantDelete(tenants []string) {
	//Delete tenants and all associated data
	term.Write("Executing Method :  Bulk Tenant Delete)", term.Blank)

	var err error
	isAllDeleted := true

	for _, tid := range tenants {
		//Get All users for tenant
		th := TenantHandler{}
		users := th.GetUsersForTenantInDetail(session.AuthCertificate{}, tid)

		//Remove all users from the tenant.
		for _, user := range users {
			status := th.RemoveUserFromTenant(user.UserID, tid)
			if status {
				//switch the person if default tenant is this tenant
				defT := th.GetDefaultTenant(user.UserID)
				if defT.TenantID == tid {
					//get all tenants for user
					allTenants := th.GetTenantsForUser(user.UserID)
					if len(allTenants) == 0 {
						//when user have no other tenants. delete default tenant so
						//boarding process will begin in next login
						client.Go("ignore", "com.duosoftware.tenant", "defaulttenant").DeleteObject().WithKeyField("UserID").AndDeleteObject(user).Ok()
					} else {
						for _, tenant := range allTenants {
							if tenant.TenantID != tid {
								//Change the default tenant
								th.SetDefaultTenant(user.UserID, tenant.TenantID)
								break
							}
						}
					}
				}
			} else {
				isAllDeleted = false
			}
		}

		//delete tenant
		tObj := make(map[string]interface{})
		tObj["TenantID"] = tid
		err = client.Go("ignore", "com.duosoftware.tenant", "tenants").DeleteObject().WithKeyField("TenantID").AndDeleteOne(tObj).Ok()
		if err != nil {
			isAllDeleted = false
		}

		//Delete database
		_ = client.Go("ignore", tid, "ignore").DeleteNamespace().Ok()
	}

	response := make(map[string]interface{})

	if err != nil {
		response["Status"] = false
		response["Message"] = err.Error()
	} else {
		response["Status"] = true
		if isAllDeleted {
			response["Message"] = "All tenants and related data successfully removed."
		} else {
			response["Message"] = "Some tenants and related data removal failed."
		}
	}

	b, _ := json.Marshal(response)
	T.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
	return
}

func (T TenantSvc) ValidateCode(token string) bool {
	//Get Users inside a Tenant
	term.Write("Executing Method : Validate Code)", term.Blank)

	tmp := tempRequestGenerator{}
	o, _ := tmp.GetRequestCode(token)

	if o["id"] == "" {
		return false
	} else {
		return true
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

func (T TenantSvc) GetUserDetails(TenantID string) []User {
	//Get Users inside a Tenant
	term.Write("Executing Method : Get Users Details (Inside a tenant)", term.Blank)

	u, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")

	if error == "" {
		th := TenantHandler{}
		return th.GetUsersForTenantInDetail(u, TenantID)
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		return []User{}
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

func (T TenantSvc) CancelAddTenantUser(email string) bool {
	//Add User to Tenant
	term.Write("Executing Method : Cancel Add User", term.Blank)
	th := TenantHandler{}

	u, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {

		//check if requester is Admin of his tenant
		isAdmin := false
		admins := th.GetTenantAdmin(u.Domain)
		for _, admin := range admins {
			fmt.Println("Admin : " + admin["EmailAddress"])
			if admin["UserID"] == u.UserID {
				isAdmin = true
				break
			}
		}

		if isAdmin {

			//Get pending add user requst
			addRequest := th.GetPendingAddUserRequest(email, u.Domain)
			if addRequest == (PendingUserRequest{}) {
				T.ResponseBuilder().SetResponseCode(500).WriteAndOveride([]byte(common.ErrorJson("Error : Tenant request in Domain : " + u.Domain + " for User : " + email + " not cound.")))
				return false
			}

			code := addRequest.Code
			tmp := tempRequestGenerator{}

			tmpObj := make(map[string]interface{})
			tmpObj["id"] = code
			tmp.Remove(tmpObj)
			th.RemoveAddUserRequest(email, addRequest.TenantID)
			//Remove any tokens in tmprequest from a email
			tmp.RemoveByEmail(email, addRequest.TenantID)
			T.ResponseBuilder().SetResponseCode(200).WriteAndOveride([]byte(common.MsgJson("Successfully removed tenant invitation.")))
			return true
		} else {
			T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Access Denied. Not an tenant administrator.")))
			return false
		}
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		return false
	}
}

func (T TenantSvc) AddUser(email, level string) bool {
	//Add User to Tenant
	term.Write("Executing Method : Add User (To Tenant)", term.Blank)

	auth := AuthHandler{}
	th := TenantHandler{}

	addUserType := T.Context.Request().Header.Get("AddUserType")

	fmt.Println("----------------------------------")
	fmt.Println("Add User Type : " + addUserType)
	fmt.Println("----------------------------------")

	inviter, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		//no error
		invitee, err := auth.GetUser(email)

		if err == "" && invitee != (User{}) {
			//User already exists in system
			t := th.GetTenant(inviter.Domain)

			//check if user already is in that tenant.
			isAlreadyInTenant := false
			tenants := th.GetTenantsForUser(invitee.UserID)

			for _, singleTenant := range tenants {
				if singleTenant.TenantID == inviter.Domain {
					isAlreadyInTenant = true
					break
				}
			}

			if isAlreadyInTenant {
				errStr := "User : " + email + " already a member of Tenant : " + inviter.Domain
				fmt.Println(errStr)
				T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(errStr)))
				return false
			} else {
				if strings.EqualFold(addUserType, "invite") {
					//send email to confirm. add to tenant from AcceptRequest
					tmp := tempRequestGenerator{}
					o := make(map[string]string)
					o["process"] = "tenant_invitation_existing_request_consent"
					o["email"] = email
					o["invitedUserID"] = inviter.UserID
					o["name"] = inviter.Name
					o["domain"] = inviter.Domain
					o["fromuseremail"] = inviter.Email
					o["tname"] = t.Name
					o["level"] = level
					o["TenantID"] = t.TenantID
					o["inviteeName"] = invitee.Name
					//o["userid"] = invitee.UserID
					code := tmp.GenerateRequestCode(o)
					var inputParams map[string]string
					inputParams = make(map[string]string)
					inputParams["@@EMAIL@@"] = email
					inputParams["@@INVEMAIL@@"] = inviter.Email
					inputParams["@@NAME@@"] = inviter.Name
					inputParams["@@DOMAIN@@"] = inviter.Domain
					inputParams["@@CODE@@"] = code

					s := PendingUserRequest{}
					s.UserID = invitee.UserID
					s.Email = invitee.EmailAddress
					s.TenantID = t.TenantID
					s.Name = invitee.Name
					//s.Code = "Not Available Reason : Tenant_Invitation_Existing"
					s.Code = code
					th.SavePendingAddUserRequest(s)

					fmt.Println("-----------------------------------------------")
					fmt.Println("Tenant Consent Email to Existing User ..... ")
					fmt.Println(inputParams)
					fmt.Println("-----------------------------------------------")

					go notifier.Notify("ignore", "tenant_invitation_existing_request_consent", email, inputParams, nil)
					return true
				} else {
					//send email and add to tenant without consent
					var inputParams map[string]string
					inputParams = make(map[string]string)
					inputParams["@@EMAIL@@"] = email
					inputParams["@@INVEMAIL@@"] = inviter.Email
					inputParams["@@NAME@@"] = inviter.Name
					inputParams["@@DOMAIN@@"] = inviter.Domain

					fmt.Println("-----------------------------------------------")
					fmt.Println("Tenant Invitation Existing ..... ")
					fmt.Println(inputParams)
					fmt.Println("-----------------------------------------------")

					go notifier.Notify("ignore", "tenant_invitation_existing", email, inputParams, nil)
					//add user to tenant
					th.AddUsersToTenant(inviter.Domain, t.Name, invitee.UserID, level)
					return true
				}
			}
		} else {
			//brand new user
			tmp := tempRequestGenerator{}
			t := th.GetTenant(inviter.Domain)
			o := make(map[string]string)
			o["process"] = "tenant_invitation"
			o["email"] = email
			o["invitedUserID"] = inviter.UserID
			o["name"] = inviter.Name
			o["domain"] = inviter.Domain
			o["fromuseremail"] = inviter.Email
			o["tname"] = t.Name
			o["level"] = level
			o["TenantID"] = t.TenantID
			//o["userid"] = invitee.UserID
			code := tmp.GenerateRequestCode(o)
			var inputParams map[string]string
			inputParams = make(map[string]string)
			inputParams["@@EMAIL@@"] = email
			inputParams["@@INVEMAIL@@"] = inviter.Email
			inputParams["@@NAME@@"] = inviter.Name
			inputParams["@@DOMAIN@@"] = inviter.Domain
			inputParams["@@CODE@@"] = code

			s := PendingUserRequest{}
			s.UserID = invitee.UserID
			s.Email = email
			s.TenantID = t.TenantID
			s.Name = invitee.Name
			s.Code = code
			th.SavePendingAddUserRequest(s)

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
			//check if requester is tenant admin
			isAdmin := false
			admins := T.GetTenantAdmin(user.Domain)
			for _, individualadmin := range admins {
				if individualadmin.Email == user.Email {
					isAdmin = true
					break
				}
			}

			if !isAdmin {
				T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("You should be a Tenant Administrator to perform this action.")))
				return false
			}

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
			status := th.RemoveUserFromTenant(u.UserID, user.Domain)
			if status {
				//switch the person if default tenant is this tenant
				defT := th.GetDefaultTenant(u.UserID)
				if defT.TenantID == user.Domain {
					//Change the default tenant
					//get all tenants for user
					allTenants := th.GetTenantsForUser(u.UserID)
					for _, tenant := range allTenants {
						if tenant.TenantID != user.Domain {
							th.SetDefaultTenant(u.UserID, tenant.TenantID)
							break
						}
					}
				}
			} else {
				//do nothing for now. add email later?
			}
			return status
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

	// tenant_invitation_existing_request_consent

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
			if th.IncreaseTenantCountInRatingEngine(o["domain"], "ignore") {
				//if th.IncreaseTenantCountInRatingEngine(o["domain"], T.Context.Request().Header.Get("Securitytoken")) {
				fmt.Println(o)
				fmt.Println("Adding User To Tenant Now")
				th.AddUsersToTenant(o["domain"], o["tname"], a.UserID, o["level"])
				inputParams["@@CNAME@@"] = a.Name
				inputParams["@@DOMAIN@@"] = o["domain"]
				inputParams["@@INVITEE@@"] = o["email"]
				inputParams["@@TENANTID@@"] = o["TenantID"]
				//go notifier.Notify("ignore", "tenant_accepted_success", email, inputParams, nil)
				fmt.Println("-----------------------------------")
				fmt.Println(email)
				fmt.Println(o["email"])
				fmt.Println(o["fromuseremail"])
				fmt.Println("-----------------------------------")
				go notifier.Notify("ignore", "tenant_invitation_added_success", o["fromuseremail"], inputParams, nil)
				th.RemoveAddUserRequest(o["email"], o["TenantID"])
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
				//th.RemoveAddUserRequest(o["email"], o["TenantID"])
				return true
			} else {
				return false
			}
		} else {
			T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Email not registered.")))
			return false
		}
		break
	case "tenant_invitation_existing_request_consent":
		auth := AuthHandler{}
		a, err := auth.GetUser(o["email"])
		if err == "" && a != (User{}) {
			fmt.Println(o)
			fmt.Println("Adding User To Tenant Now")
			th.AddUsersToTenant(o["domain"], o["tname"], a.UserID, o["level"])
			inputParams["@@CNAME@@"] = o["name"]
			//inputParams["@@INVITEE@@"] = o["inviteeName"]
			inputParams["@@INVITEE@@"] = o["email"]
			inputParams["@@DOMAIN@@"] = o["domain"]
			inputParams["@@TENANTID@@"] = o["TenantID"]

			th.RemoveAddUserRequest(o["email"], o["TenantID"])

			fmt.Println("-----------------------------------------------")
			fmt.Println("Tenant Invitation accepted mail to ADMIN..... ")
			fmt.Println(inputParams)
			fmt.Println("-----------------------------------------------")

			//send email to admin that user has agreed to accept the request
			go notifier.Notify("ignore", "tenant_invitation_added_success", o["fromuseremail"], inputParams, nil)
			T.ResponseBuilder().SetResponseCode(200).WriteAndOveride([]byte(common.MsgJson("You have been successfully completed tenant invitation process. Login and use tenant switcher to switch between avaiable tenants.")))
			return true
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

		//get the default tenant for user

		defaultTenant := th.GetDefaultTenant(user.UserID)

		allTenants := th.GetTenantsForUser(user.UserID)

		if len(allTenants) > 1 {
			for index, singleTenant := range allTenants {
				if singleTenant.TenantID == defaultTenant.TenantID && index != 0 {
					tempTenant := allTenants[0]
					allTenants[0] = singleTenant
					allTenants[index] = tempTenant
					break
				}
			}
			return allTenants
		} else {
			return allTenants
		}
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
				T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("User : " + user.Email + " is already Subscribed to Tenant : " + TenantID))
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
