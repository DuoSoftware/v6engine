package pog

import (
	"duov6.com/common"
	"duov6.com/objectstore/client"
	//"duov6.com/session"
	"duov6.com/term"
	"encoding/json"
)

type PermistionRecords struct {
	RecordID    string
	Name        string
	AccessLevel string
	OtherData   map[string]string
}

type Cirtificat struct {
	POGid       string
	AccessLevel string
	RecordID    string
	OtherData   map[string]string
}

type RecordUsers struct {
	RecordID string
	///PageNo   string
	UserIDs []string
}

type UserRecords struct {
	UserID  string
	Records []string
}
type SecInfo struct {
	POGDomain     string
	SecurityToken string
}

func Add(GUUserID, RecordID, Name, AccessLevel string, OtherData map[string]string, s SecInfo) {
	var p PermistionRecords
	p.AccessLevel = AccessLevel
	p.RecordID = RecordID
	p.Name = Name
	p.OtherData = OtherData
	addCirt(GUUserID, p, s)
}

func Access(UserID, recordID string, s SecInfo) Cirtificat {
	return getCirt(UserID, recordID, s)
}

func GetUsers(RecordID string, s SecInfo) []string {
	return getUsers(RecordID, s)
}

func getCirt(UserID, recordID string, s SecInfo) Cirtificat {
	term.Write("Methed Invoke getCirt", term.Debug)
	var c Cirtificat
	RecID := getPODID(UserID, recordID)
	bytes, err := client.Go(s.SecurityToken, s.POGDomain, "cirts").GetOne().ByUniqueKey(RecID).Ok()
	//var t Tenant
	if err == "" {
		err := json.Unmarshal(bytes, &c)
		if err == nil {
			return c
		} else {
			term.Write("Methed Invoke getCirt :"+err.Error(), term.Error)
			return c
		}
	} else {
		term.Write("Methed Invoke getCirt :"+err, term.Error)
		return c
	}
}

func addCirt(GUUserID string, p PermistionRecords, s SecInfo) {
	var c Cirtificat
	var r RecordUsers
	term.Write("Methed Invoke addCirt", term.Debug)

	c.POGid = getPODID(GUUserID, p.RecordID)
	c.AccessLevel = p.AccessLevel
	c.OtherData = p.OtherData
	r.RecordID = p.RecordID
	r.UserIDs = getUsers(p.RecordID, s)
	client.Go(s.SecurityToken, s.POGDomain, "cirts").StoreObject().WithKeyField("POGid").AndStoreOne(c).Ok()
	term.Write("Methed Invoke addCirt Insearted to com.duosoftware.pog.cirts."+c.POGid, term.Debug)
	for _, element := range r.UserIDs {
		if element == GUUserID {
			return
		}
	}
	r.UserIDs = append(r.UserIDs, GUUserID)
	client.Go(s.SecurityToken, s.POGDomain, "records").StoreObject().WithKeyField("RecordID").AndStoreOne(r).Ok()
	term.Write("Methed Invoke addCirt Inserted to com.duosoftware.pog.records."+r.RecordID, term.Debug)
}

func getUsers(RecordID string, s SecInfo) []string {
	term.Write("Methed Invoke getUsers", term.Debug)
	var c RecordUsers
	//RecID = getPODID(UserID, recordID)
	bytes, err := client.Go(s.SecurityToken, s.POGDomain, "records").GetOne().ByUniqueKey(RecordID).Ok()
	//var t Tenant
	if err == "" {
		err := json.Unmarshal(bytes, &c)
		if err == nil {
			return c.UserIDs
		} else {
			return c.UserIDs
		}
	} else {
		return c.UserIDs
	}
}

func getPODID(UserID, recordID string) string {
	term.Write("Methed Invoke getPODID", term.Debug)
	return common.GetHash(UserID + "-" + recordID)
}
