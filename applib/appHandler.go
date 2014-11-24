package applib

import (
	//"code.google.com/p/gorest"
	"duov6.com/common"
	"duov6.com/objectstore/client"
	"duov6.com/term"
	"encoding/json"
)

type Apphanler struct {
}

func (app *Apphanler) Get(ApplicationID string) (App Application, errMessage string) {
	term.Write("Get  App  by ID"+ApplicationID, term.Debug)
	bytes, err := client.Go("ignore", "com.duosoftware.application", "apps").GetOne().BySearching(ApplicationID).Ok()
	var a Application
	if err == "" {
		if bytes != nil {
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

func (app *Apphanler) Add(a Application) (ourApp Application, errMessage string) {
	term.Write("Add saving Application  "+a.Name, term.Debug)

	bytes, err := client.Go("ignore", "com.duosoftware.application", "apps").GetOne().BySearching(a.ApplicationID).Ok()
	if err == "" {
		var uList []Application
		//if bytes==nil{
		err := json.Unmarshal(bytes, &uList)

		if err == nil || bytes == nil {
			if len(uList) == 0 {

				a.ApplicationID = common.GetGUID()
				a.SecretKey = common.RandText(10)
				term.Write("Add saving Addaplication  "+a.Name+" New App "+a.ApplicationID, term.Debug)

				client.Go("ignore", "com.duosoftware.Application", "apps").StoreObject().WithKeyField("ApplicationID").AndStoreOne(a).Ok()
				ourApp = a
				errMessage = ""
				return
			} else {
				a.ApplicationID = uList[0].ApplicationID
				term.Write("SaveUser saving user  "+a.Name+" Update User "+a.ApplicationID, term.Debug)
				client.Go("ignore", "com.duosoftware.Application", "apps").StoreObject().WithKeyField("ApplicationID").AndStoreOne(a).Ok()
				ourApp = a
				errMessage = ""
				return
			}

		} else {
			term.Write("SaveUser saving user store Error 2#"+err.Error(), term.Error)
			errMessage = " saving Application fetech Error #" + err.Error()
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
