package authlib

type userSVC struct {
	addUserRoles    gorest.EndPoint `method:"POST" path:"/user/AddUserRoles/" postdata:"UserRole"`
	getMyRoles      gorest.EndPoint `method:"GET" path:"/user/GetMyRoles" output:"[]RoleMinimum"`
	getRoles        gorest.EndPoint `method:"GET" path:"/user/GetRoles/{GUUserID:string}" output:"[]RoleMinimum"`
	removeUserRoles gorest.EndPoint `method:"GET" path:"/user/RemoveUserRoles/{RoleID:string}" output:"bool"`
}

func (A userSVC) AddUserRoles(u UserRole) {
	h := newAuthHandler()
	c, err := h.GetSession(A.Context.Request().Header.Get("Securitytoken"), "Nil")
	if err == "" {
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Security Token Not Incorrect"))
	}

}

func (A userSVC) RemoveUserRoles(string RoleID) bool {
	h := newAuthHandler()
	c, err := h.GetSession(A.Context.Request().Header.Get("Securitytoken"), "Nil")
	if err == "" {
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Security Token Not Incorrect"))
	}
	return true

}

func (A userSVC) GetMyRoles() []RoleMinimum {
	h := newAuthHandler()
	c, err := h.GetSession(A.Context.Request().Header.Get("Securitytoken"), "Nil")
	r := []RoleMinimum{}
	if err == "" {
		return r
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Security Token Not Incorrect"))
		return r
	}
}

func (A userSVC) GetRoles(UserID string) []RoleMinimum {
	h := newAuthHandler()
	c, err := h.GetSession(A.Context.Request().Header.Get("Securitytoken"), "Nil")
	r := []RoleMinimum{}
	if err == "" {
		return r
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Security Token Not Incorrect"))
		return r
	}
}
