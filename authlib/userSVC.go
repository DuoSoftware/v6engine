package authlib

import (
	//"duov6.com/common"
	"duov6.com/gorest"
	//"encoding/json"
	//"fmt"
)

type UserSVC struct {
	addUserRoles    gorest.EndPoint `method:"POST" path:"/user/AddUserRoles/" postdata:"UserRole"`
	getMyRoles      gorest.EndPoint `method:"GET" path:"/user/GetMyRoles" output:"[]RoleMinimum"`
	getRoles        gorest.EndPoint `method:"GET" path:"/user/GetRoles/{GUUserID:string}" output:"[]RoleMinimum"`
	removeUserRoles gorest.EndPoint `method:"GET" path:"/user/RemoveUserRoles/{RoleID:string}" output:"bool"`
	addOtherData    gorest.EndPoint `method:"GET" path:"/user/addOtherdata/{UserID:string}/{Filed:string}/{Value:string}" output:"bool"`
	gorest.RestService
}

func (A UserSVC) AddUserRoles(u UserRole) {
	h := newAuthHandler()
	_, err := h.GetSession(A.Context.Request().Header.Get("Securitytoken"), "Nil")

	if err == "" {
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Security Token Not Incorrect"))
	}

}

func (A UserSVC) AddOtherData(UserID, Filed, Value string) bool {
	return true
}

func (A UserSVC) RemoveUserRoles(RoleID string) bool {
	h := newAuthHandler()
	_, err := h.GetSession(A.Context.Request().Header.Get("Securitytoken"), "Nil")
	if err == "" {
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Security Token Not Incorrect"))
	}
	return true

}

func (A UserSVC) GetMyRoles() []RoleMinimum {
	h := newAuthHandler()
	_, err := h.GetSession(A.Context.Request().Header.Get("Securitytoken"), "Nil")
	r := []RoleMinimum{}
	if err == "" {
		return r
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Security Token Not Incorrect"))
		return r
	}
}

func (A UserSVC) GetRoles(UserID string) []RoleMinimum {
	h := newAuthHandler()
	_, err := h.GetSession(A.Context.Request().Header.Get("Securitytoken"), "Nil")
	r := []RoleMinimum{}
	if err == "" {
		return r
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Security Token Not Incorrect"))
		return r
	}
}
