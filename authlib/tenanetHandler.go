package authlib

import (
	"duov6.com/common"
	//"duov6.com/email"
	email "duov6.com/duonotifier/client"
	"duov6.com/objectstore/client"
	"duov6.com/session"
	"duov6.com/term"
	"encoding/json"
)

type Tenant struct {
	TenantID  string
	Name      string
	Shell     string
	Statistic map[string]string
	Private   bool
	OtherData map[string]string
}

type InviteUsers struct {
	Email         string
	Name          string
	UserID        string
	SecurityLevel string
}

type InviteUserRequest struct {
	Email         string
	Name          string
	UserID        string
	FromName      string
	FromEmail     string
	FromUserID    string
	TenantID      string
	SecurityLevel string
	RequestToken  string
}

type TenantMinimum struct {
	TenantID string
	Name     string
}

//type  int8

type TenantAutherized struct {
	ID            string
	UserID        string
	TenantID      string
	SecurityLevel string
	Autherized    bool
}

type UserTenants struct {
	UserID    string
	TenantIDs []TenantMinimum
}

type TenantUsers struct {
	TenantID string
	Users    []string
}

type TenantHandler struct {
}

func (h *TenantHandler) CreateTenant(t Tenant, user session.AuthCertificate, update bool) Tenant {
	term.Write("CreateTenant saving user  "+t.Name, term.Debug)
	//client.c
	bytes, err := client.Go("ignore", "com.duosoftware.tenant", "tenants").GetOne().ByUniqueKey(t.TenantID).Ok()
	if err == "" {
		var uList Tenant
		err := json.Unmarshal(bytes, &uList)
		if err != nil || uList.TenantID == "" {
			if t.TenantID == "" {
				t.TenantID = common.GetGUID()
				term.Write("Auto Gen TID  "+t.TenantID+" New Tenant "+t.Name, term.Debug)
			}
			term.Write("Save Tenant saving Tenant  "+t.Name+" New Tenant "+t.Name, term.Debug)
			var inputParams map[string]string
			inputParams = make(map[string]string)
			inputParams["@@email@@"] = user.Email
			inputParams["@@name@@"] = user.Name
			inputParams["@@tenantID@@"] = t.TenantID
			inputParams["@@tenantName@@"] = t.Name
			h.AddUsersToTenant(t.TenantID, t.Name, user.UserID, "admin")
			email.Send("ignore", "Tenent Creation Notification!", "com.duosoftware.auth", "tenant", "tenant_creation", inputParams, nil, user.Email)
			//email.Send("ignore", "com.duosoftware.auth", "tenant", "tenant_creation", inputParams, user.Email)
			client.Go("ignore", "com.duosoftware.tenant", "tenants").StoreObject().WithKeyField("TenantID").AndStoreOne(t).Ok()
		} else {
			if update {
				term.Write("SaveUser saving Tenant  "+t.TenantID+" Update user "+user.UserID, term.Debug)
				client.Go("ignore", "com.duosoftware.tenant", "tenants").StoreObject().WithKeyField("TenantID").AndStoreOne(t).Ok()
			}
		}
	} else {
		term.Write("SaveUser saving Tenant fetech Error #"+err, term.Error)
	}
	return t
}

func (h *TenantHandler) AutherizedUser(TenantID, UserID string) (bool, TenantAutherized) {
	term.Write("Start Autherized Domain #"+TenantID, term.Debug)
	bytes, err := client.Go("ignore", "com.duosoftware.tenant", "authorized").GetOne().ByUniqueKey(common.GetHash(UserID + "-" + TenantID)).Ok()
	term.Write("SaveUser saving Tenant fetech Error #", term.Debug)
	if err == "" {
		var uList TenantAutherized
		err := json.Unmarshal(bytes, &uList)
		if err == nil {
			term.Write("Autherized #", term.Debug)
			return uList.Autherized, uList
		} else {
			term.Write("Fail to deasseble Not Autherized #", term.Debug)
			return false, TenantAutherized{}
		}
	} else {
		term.Write("Not Autherized #", term.Debug)
		return false, TenantAutherized{}
	}
	term.Write("Start Global Autherized Domain #"+TenantID, term.Debug)
	bytes1, err1 := client.Go("ignore", "com.duosoftware.tenant", "authorized").GetOne().ByUniqueKey(TenantID).Ok()
	if err1 == "" {
		var uList TenantAutherized
		err := json.Unmarshal(bytes1, &uList)
		if err == nil {
			term.Write("Autherized #", term.Debug)
			return uList.Autherized, uList
		} else {
			term.Write("Fail to deasseble Not Autherized #", term.Debug)
			return false, TenantAutherized{}
		}
	} else {
		term.Write("Not Autherized #", term.Debug)
		return false, TenantAutherized{}
	}
}

func (h *TenantHandler) Autherized(TenantID string, user session.AuthCertificate) (bool, TenantAutherized) {
	return h.AutherizedUser(TenantID, user.UserID)
}

func (h *TenantHandler) AuthorizedGlobalTenants(TenantID string) (bool, TenantAutherized) {
	bytes, err := client.Go("ignore", "com.duosoftware.tenant", "authorized").GetOne().ByUniqueKey(common.GetHash(TenantID)).Ok()
	if err == "" {
		var uList TenantAutherized
		err := json.Unmarshal(bytes, &uList)
		if err == nil {
			return uList.Autherized, uList
		} else {
			return false, TenantAutherized{}
		}
	} else {
		return false, TenantAutherized{}
	}
}

func (h *TenantHandler) GetTenant(TenantID string) Tenant {
	bytes, err := client.Go("ignore", "com.duosoftware.tenant", "tenants").GetOne().ByUniqueKey(TenantID).Ok()
	var t Tenant
	if err == "" {
		err := json.Unmarshal(bytes, &t)
		if err == nil {
			return t
		} else {
			return t
		}
	} else {
		return t
	}
}

func (h *TenantHandler) AddTenantForUsers(Tenant TenantMinimum, UserID string) UserTenants {
	bytes, err := client.Go("ignore", "com.duosoftware.tenant", "userstenantmappings").GetOne().ByUniqueKey(UserID).Ok()
	var t UserTenants

	//t.UserID
	if err == "" {
		err := json.Unmarshal(bytes, &t)
		if err != nil || t.UserID == "" {
			term.Write("No Users yet assigied "+UserID, term.Debug)
			t = UserTenants{UserID, []TenantMinimum{}}
			t.UserID = UserID
		} else {
			for _, element := range t.TenantIDs {
				if element.TenantID == Tenant.TenantID {
					return t
				}
			}
		}
		t.TenantIDs = append(t.TenantIDs, Tenant)
		client.Go("ignore", "com.duosoftware.tenant", "userstenantmappings").StoreObject().WithKeyField("UserID").AndStoreOne(t).Ok()
		term.Write("Saved Tenant users"+UserID, term.Debug)
		return t
	} else {
		return t
	}
}

func (h *TenantHandler) GetTenantsForUser(UserID string) []TenantMinimum {
	bytes, err := client.Go("ignore", "com.duosoftware.tenant", "userstenantmappings").GetOne().ByUniqueKey(UserID).Ok()
	var t UserTenants
	if err == "" {
		err := json.Unmarshal(bytes, &t)
		if err == nil {
			return t.TenantIDs
		} else {
			return []TenantMinimum{}
		}
	} else {
		return []TenantMinimum{}
	}
}

func (h *TenantHandler) GetUsersForTenant(u session.AuthCertificate, TenantID string) []string {
	bytes, err := client.Go("ignore", "com.duosoftware.tenant", "users").GetOne().ByUniqueKey(TenantID).Ok()
	var t TenantUsers
	if err == "" {
		err := json.Unmarshal(bytes, &t)
		if err == nil {
			return t.Users
		} else {
			return []string{}
		}
	} else {
		return []string{}
	}
}

func (h *TenantHandler) AddUserToTenant(u session.AuthCertificate, users []InviteUsers) {
	for _, user := range users {
		var inputParams map[string]string
		inputParams = make(map[string]string)
		inputParams["@@email@@"] = user.Email
		inputParams["@@name@@"] = user.Name
		inputParams["@@userID@@"] = user.UserID
		inputParams["@@tenantID@@"] = u.Domain
		inputParams["@@FromName@@"] = u.Username
		inputParams["@@FromID@@"] = u.UserID
		inputParams["@@FromEmail@@"] = u.Email
		req := InviteUserRequest{}
		req.UserID = user.UserID
		req.TenantID = u.Domain
		req.RequestToken = common.RandText(10)
		req.Name = user.Name
		req.FromUserID = u.UserID
		req.FromName = u.Name
		req.FromEmail = u.Email
		req.Email = user.Email
		req.SecurityLevel = user.SecurityLevel

		//h.AddUsersToTenant(t.TenantID, user.UserID, "admin")
		client.Go("ignore", "com.duosoftware.tenant", "userrequest").StoreObject().WithKeyField("RequestToken").AndStoreOne(req).Ok()
		//email.Send("ignore", "com.duosoftware.auth", "tenant", "tenant_request", inputParams, user.Email)
		email.Send("ignore", "Tenent User Allocation Notification!", "com.duosoftware.auth", "tenant", "tenant_request", inputParams, nil, user.Email)

	}
}

func (h *TenantHandler) RequestToTenant(u session.AuthCertificate, TenantID string) bool {
	/*
		//t.
		//for _, user := range users {
		var inputParams map[string]string
		inputParams = make(map[string]string)
		inputParams["email"] = user.Email
		inputParams["name"] = u.Name
		inputParams["userID"] = u.UserID
		inputParams["tenantID"] = Tenant

		req := InviteUserRequest{}
		req.UserID = user.UserID
		req.TenantID = u.Domain
		req.RequestToken = common.RandText(10)
		req.Name = user.Name
		req.FromUserID = u.UserID
		req.FromName = u.Name
		req.FromEmail = u.Email
		req.Email = user.Email
		req.SecurityLevel = "request"
		inputParams["RequestToken"] = req.RequestToken
		//h.AddUsersToTenant(t.TenantID, user.UserID, "admin")
		client.Go("ignore", "com.duosoftware.tenant", "userrequest").StoreObject().WithKeyField("RequestToken").AndStoreOne(req).Ok()
		email.Send("ignore", "com.duosoftware.auth", "tenant", "tenant_request", inputParams, user.Email)
		//}
	*/
	return true
}

func (h *TenantHandler) AcceptRequest(u session.AuthCertificate, securityLevel, RequestToken string, accept bool) bool {
	bytes, err := client.Go("ignore", "com.duosoftware.tenant", "userrequest").GetOne().ByUniqueKey(RequestToken).Ok()
	var t InviteUserRequest
	if err == "" {
		err := json.Unmarshal(bytes, &t)
		if err != nil || t.SecurityLevel == "" {
			if securityLevel == "" {
				securityLevel = t.SecurityLevel
			}
			if accept {
				h.AddUsersToTenant(t.TenantID, t.Name, t.UserID, securityLevel)
				return true
			} else {
				return true
			}
		} else {
			return false
		}
	} else {
		return false
	}
}

func (h *TenantHandler) AddUsersToTenant(TenantID, Name string, users, SecurityLevel string) TenantUsers {
	bytes, err := client.Go("ignore", "com.duosoftware.tenant", "users").GetOne().ByUniqueKey(TenantID).Ok()
	var t TenantUsers
	if err == "" {
		err := json.Unmarshal(bytes, &t)
		if err != nil || t.TenantID == "" {
			term.Write("No Users yet assigied "+t.TenantID, term.Debug)
			//t=TenantUsers{}
			t = TenantUsers{TenantID, []string{}}
			t.TenantID = TenantID
		} else {
			for _, element := range t.Users {
				if element == users {
					return t
				}
			}
		}
		h.AddTenantForUsers(TenantMinimum{TenantID, Name}, users)
		t.Users = append(t.Users, users)
		var Activ TenantAutherized
		Activ = TenantAutherized{}
		id := common.GetHash(users + "-" + TenantID)
		Activ.Autherized = true
		Activ.ID = id
		Activ.TenantID = TenantID
		Activ.SecurityLevel = SecurityLevel
		Activ.UserID = users
		client.Go("ignore", "com.duosoftware.tenant", "authorized").StoreObject().WithKeyField("ID").AndStoreOne(Activ).Ok()
		client.Go("ignore", "com.duosoftware.tenant", "users").StoreObject().WithKeyField("TenantID").AndStoreOne(t).Ok()
		term.Write("Saved Tenant users"+t.TenantID, term.Debug)
		return t
	} else {
		return t
	}
}

func (h *TenantHandler) RemoveUserFromTenant(UserID, TenantID string) bool {
	id := common.GetHash(UserID + "-" + TenantID)
	var Activ TenantAutherized
	Activ = TenantAutherized{}
	Activ.ID = id
	Activ.TenantID = TenantID

	bytes, err := client.Go("ignore", "com.duosoftware.tenant", "users").GetOne().ByUniqueKey(TenantID).Ok()
	var t TenantUsers
	if err == "" {
		err := json.Unmarshal(bytes, &t)
		if err == nil {
			s := []string{}
			for _, element := range t.Users {
				if element != UserID {
					s = append(s, element)
				}

			}
			t.Users = s
			client.Go("ignore", "com.duosoftware.tenant", "users").StoreObject().WithKeyField("TenantID").AndStoreOne(t).Ok()
			client.Go("ignore", "com.duosoftware.tenant", "authorized").DeleteObject().AndDeleteObject(Activ).ByUniqueKey("ID").Ok()
			//client.Go(securityToken, namespace, class)
			//return t.Users
		}
	} else {
		return false
	}

	var ut UserTenants
	client.Go("ignore", "com.duosoftware.tenant", "authorized").DeleteObject().AndDeleteObject(Activ).ByUniqueKey("ID").Ok()
	bytes1, err1 := client.Go("ignore", "com.duosoftware.tenant", "userstenantmappings").GetOne().ByUniqueKey(UserID).Ok()
	//var t TenantUsers
	if err1 == "" {
		err := json.Unmarshal(bytes1, &ut)
		if err == nil {
			s := []TenantMinimum{}
			//ut.UserID
			for _, element := range ut.TenantIDs {
				if element.TenantID != TenantID {
					s = append(s, element)
				}
			}
			ut.TenantIDs = s
			client.Go("ignore", "com.duosoftware.tenant", "userstenantmappings").StoreObject().WithKeyField("UserID").AndStoreOne(ut).Ok()
		} else {
			return false
		}
	}

	return true

}

func (h *TenantHandler) SearchTenants(Search string, since, pagesize int) []Tenant {
	bytes, err := client.Go("ignore", "com.duosoftware.tenant", "tenants").GetMany().BySearching(Search).Ok()
	var t []Tenant
	if err == "" {
		err := json.Unmarshal(bytes, &t)
		if err != nil {
			return t
		}
	}

	return t
}
