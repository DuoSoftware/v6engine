package applib

import (
	//"code.google.com/p/gorest"
	"duov6.com/common"
	"duov6.com/objectstore/client"
	"duov6.com/pog"
	"duov6.com/session"
	"duov6.com/term"
	"encoding/json"
)

type Apphanler struct {
}

func (app *Apphanler) Get(ApplicationID string, securityToken string) (App Application, errMessage string) {
	term.Write("Get  App  by ID"+ApplicationID, term.Debug)
	_, status := session.GetSession(securityToken, "Nil")
	bytes, err := client.Go(securityToken, s.Domain, "apps").GetOne().ByUniqueKey(ApplicationID).Ok()
	//bytes, err := client.Go(securityToken, "com.duosoftware.application", "apps").GetOne().ByUniqueKey(ApplicationID).Ok()
	var a Application
	if err == "" {
		if bytes != nil || status == "error" {
			var uList []Application
			err := json.Unmarshal(bytes, &uList)

			if err == nil && len(uList) != 0 {
				App = uList[0]
				errMessage = ""
				return
			} else {
				if err != nil {
					term.Write("Login  user Error "+err.Error(), term.Error)
					App = a
					errMessage = "Get Application  Error " + err.Error()
					return
				}
			}
		} else {
			App = a
			errMessage = "Application Not Found " + ApplicationID
			return

		}

	} else {

		term.Write("Login  user  Error "+err, term.Error)
		App = a
		errMessage = "Get Application  Error " + err
		return
	}

	App = a
	errMessage = "Unable to process"
	return
}

func (app *Apphanler) Add(a Application, securityToken string) (ourApp Application, errMessage string) {
	term.Write("Add saving Application  "+a.Name, term.Debug)

	s, err := session.GetSession(securityToken, "Nil")
	if err != "" {
		ourApp = a
		errMessage = err
		return
	}
	bytes, err := client.Go(securityToken, s.Domain, "apps").GetOne().ByUniqueKey(a.ApplicationID).Ok()
	//bytes, err := client.Go(securityToken, "com.duosoftware.application", "apps").GetOne().ByUniqueKey(a.ApplicationID).Ok()
	if err == "" {

		var uList Application

		json.Unmarshal(bytes, &uList)

		if bytes == nil {
			a.ApplicationID = common.GetGUID()
			a.SecretKey = common.RandText(10)
			//security token token from header
			term.Write("Add saving Add aplication  "+a.Name+" New App "+a.ApplicationID, term.Debug)

			client.Go(securityToken, s.Domain, "apps").StoreObject().WithKeyField("ApplicationID").AndStoreOne(a).Ok()
			//client.Go(securityToken, "com.duosoftware.Application", "apps").StoreObject().WithKeyField("ApplicationID").AndStoreOne(a).Ok()
			OtherData := make(map[string]string)
			OtherData["Description"] = a.Description

			pog.Add(s.UserID, a.ApplicationID, a.Name, "admin", OtherData, pog.SecInfo{s.Domain, s.SecurityToken})
			ourApp = a
			errMessage = ""
			return
		} else {
			a.ApplicationID = uList.ApplicationID
			term.Write("SaveUser saving user  "+a.Name+" Update User "+a.ApplicationID, term.Debug)
			client.Go(securityToken, s.Domain, "apps").StoreObject().WithKeyField("ApplicationID").AndStoreOne(a).Ok()
			//client.Go(securityToken, "com.duosoftware.Application", "apps").StoreObject().WithKeyField("ApplicationID").AndStoreOne(a).Ok()
			ourApp = a
			errMessage = ""
			return
		}

	} else {
		term.Write("SaveUser saving user fetech Error 1#"+err, term.Error)
		errMessage = "SaveUser saving user fetech Error #" + err
		return
	}
	errMessage = "Unable to process"
	ourApp = a
	return
	//return a
}
