package pog

import (
	"duov6.com/gorest"
	//"duov6.com/pog"
	"duov6.com/session"
	//"encoding/json"
	//"fmt"
)

type POGc struct {
	GUUserID    string
	RecordID    string
	Name        string
	AccessLevel string
	OtherData   map[string]string
}

type POGSvc struct {
	gorest.RestService
	add       gorest.EndPoint `method:"POST" path:"/POG/Add/" postdata:"POGc"`
	getAccess gorest.EndPoint `method:"GET" path:"/POG/GetAccess/{RecordID:string}" output:"Cirtificat"`
	getUsers  gorest.EndPoint `method:"GET" path:"/POG/GetUsers/{RecordID:string}" output:"[]string"`
}

func (T POGSvc) Add(t POGc) {
	u, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		s := SecInfo{u.Domain, u.SecurityToken}
		Add(t.GUUserID, t.RecordID, t.Name, t.AccessLevel, t.OtherData, s)
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Security Token Not Incorrect"))
	}
}

func (T POGSvc) GetAccess(RecordID string) (c Cirtificat) {
	u, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		s := SecInfo{u.Domain, u.SecurityToken}
		c = Access(u.UserID, RecordID, s)
		return
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Security Token Not Incorrect"))
		return
	}
}

func (T POGSvc) GetUsers(RecordID string) (c []string) {
	u, error := session.GetSession(T.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		s := SecInfo{u.Domain, u.SecurityToken}
		c = GetUsers(RecordID, s)
		return
	} else {
		T.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Security Token Not Incorrect"))
		return
	}
}
