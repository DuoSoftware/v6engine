package authlib

import (
	"duov6.com/applib"
	"duov6.com/common"
	"duov6.com/email"
	"duov6.com/session"
	//"duov6.com/config"
	"duov6.com/objectstore/client"
	"duov6.com/term"
	"encoding/json"
	"strings"
	"fmt"
)

type AuthHandler struct {
	//Config AuthConfig
}

func newAuthHandler() *AuthHandler {
	authhld := new(AuthHandler)
	//authhld.Config = GetConfig()
	return authhld
}

type ActivationEmail struct {
	GUUserID string
	Token    string
}

func (h *AuthHandler) AppAutherize(ApplicationID, UserID string) bool {
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "atherized").GetOne().ByUniqueKey(ApplicationID + "-" + UserID).Ok()
	term.Write("AppAutherize For Application "+ApplicationID+" UserID "+UserID, term.Debug)
	if err == "" {
		if bytes != nil {
			var uList AppAutherize
			err := json.Unmarshal(bytes, &uList)
			if err == nil {
				return true
			}
		}
	} else {
		term.Write("AppAutherize Error "+err, term.Error)
	}
	return false
}

func (h *AuthHandler) GetAuthCode(ApplicationID, UserID, URI string) string {
	var a AuthCode
	a.ApplicationID = ApplicationID
	a.UserID = UserID
	a.URI = URI
	a.Code = common.RandText(10)
	client.Go("ignore", "com.duosoftware.auth", "authcode").StoreObject().WithKeyField("Code").AndStoreOne(a).Ok()
	term.Write("GetAuthCode for "+ApplicationID+" with SecurityToken :"+UserID, term.Debug)
	return a.Code
}

func (h *AuthHandler) AutherizeApp(Code, ApplicationID, AppSecret, UserID string) (bool, string) {
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "authcode").GetOne().ByUniqueKey(Code).Ok()
	term.Write("AutherizeApp For ApplicationID "+ApplicationID+" Code "+Code+" Secret "+AppSecret+" Err "+err, term.Debug)
	var uList AuthCode
	err1 := json.Unmarshal(bytes, &uList)
	term.Write(string(bytes[:]), term.Debug)
	if err1 != nil {

		var appH applib.Apphanler
		application, err := appH.Get(ApplicationID, "ignorelib")
		if err == "" {
			if application.SecretKey == AppSecret && uList.UserID == UserID && Code == uList.Code {
				var appAth AppAutherize
				appAth.AppliccatioID = ApplicationID
				appAth.AutherizeKey = ApplicationID + "-" + UserID
				appAth.Name = application.Name

				client.Go("ignore", "com.duosoftware.auth", "atherized").StoreObject().WithKeyField("AutherizeKey").AndStoreOne(appAth).Ok()

				return true, ""
			}
		} else {
			return false, err
		}
	} else {
		return false, "Code invalid"
	}
	return false, "process error"

}

func (h *AuthHandler) AddSession(a AuthCertificate) {
	var c session.AuthCertificate
	c.ClientIP = a.ClientIP
	c.DataCaps = a.DataCaps
	c.Domain = a.Domain
	c.Email = a.Email
	c.Name = a.Name
	c.SecurityToken = a.SecurityToken
	c.UserID = a.UserID
	c.Username = a.Username
	session.AddSession(c)
}

func (h *AuthHandler) LogOut(a AuthCertificate) {
	client.Go("ignore", "s.duosoftware.auth", "sessions").DeleteObject().ByUniqueKey(a.SecurityToken)
	//client.Go("ignore", "s.duosoftware.auth", "sessions").StoreObject().WithKeyField("SecurityToken").AndStoreOne(a).Ok()
	term.Write("LogOut for "+a.Name+" with SecurityToken :"+a.SecurityToken, term.Debug)
	//return true
}

func (h *AuthHandler) GetSession(key, Domain string) (AuthCertificate, string) {
	//bytes, err := client.Go(key, "s.duosoftware.auth", "sessions").GetOne().ByUniqueKey(key).Ok()
	//term.Write("GetSession For SecurityToken "+key, term.Debug)
	a, err := session.GetSession(key, Domain)
	var c AuthCertificate
	if err == "" {
		c.ClientIP = a.ClientIP
		c.DataCaps = a.DataCaps
		c.Domain = a.Domain
		c.Email = a.Email
		c.Name = a.Name
		c.SecurityToken = a.SecurityToken
		c.UserID = a.UserID
		c.Username = a.Username
		return c, ""
	} else {
		term.Write("GetSession Error "+err, term.Error)
	}
	term.Write("GetSession No Session for SecurityToken "+key, term.Debug)

	return c, "Error Session Not Found"
}

func (h *AuthHandler) ForgetPassword(emailaddress string) bool {
	u, error := h.GetUser(emailaddress)
	if error == "" {
		passowrd := common.RandText(6)
		u.ConfirmPassword = passowrd
		u.Password = passowrd
		term.Write("Password : "+passowrd, term.Debug)
		h.SaveUser(u, true)
		var inputParams map[string]string
		inputParams = make(map[string]string)
		inputParams["email"] = u.EmailAddress
		inputParams["name"] = u.Name
		inputParams["password"] = passowrd
		email.Send("ignore", "com.duosoftware.auth", "email", "user_resetpassword", inputParams, u.EmailAddress)
		term.Write("E Mail Sent", term.Debug)
		return true
	}
	return false
}

func (h *AuthHandler) ChangePassword(a AuthCertificate, newPassword string) bool {
	u, error := h.GetUser(a.Email)
	if error == "" {
		//passowrd := common.RandText(6)
		u.ConfirmPassword = newPassword
		u.Password = newPassword
		h.SaveUser(u, true)
		return true
	}
	return false
}

func (h *AuthHandler) SaveUser(u User, update bool) User {
	term.Write("SaveUser saving user  "+u.Name, term.Debug)

	bytes, err := client.Go("ignore", "com.duosoftware.auth", "users").GetOne().ByUniqueKey(u.EmailAddress).Ok()
	if err == "" {
		var uList User
		err := json.Unmarshal(bytes, &uList)
		//if err == nil || bytes == nil {
		term.Write("SaveUser saving user retrived", term.Debug)
		fmt.Println(uList)
		term.Write("SaveUser saving user retrived", term.Debug)
		if err!=nil || uList.UserID==""  {
			u.Active = false
			u.UserID = common.GetGUID()
			term.Write("SaveUser saving user  "+u.Name+" New User "+u.UserID, term.Debug)
			password := u.Password
			u.Password = common.GetHash(u.Password)
			u.ConfirmPassword = common.GetHash(u.ConfirmPassword)
			var Activ ActivationEmail
			Activ.GUUserID = u.EmailAddress
			Activ.Token = common.RandText(10)
			var inputParams map[string]string
			inputParams = make(map[string]string)
			inputParams["email"] = u.EmailAddress
			inputParams["name"] = u.Name
			inputParams["token"] = Activ.Token
			inputParams["password"] = password
			email.Send("ignore", "com.duosoftware.auth", "email", "user_activate", inputParams, u.EmailAddress)
			term.Write("E Mail Sent", term.Debug)
			client.Go("ignore", "com.duosoftware.auth", "activation").StoreObject().WithKeyField("Token").AndStoreOne(Activ).Ok()
			term.Write("Activation stored", term.Debug)
			client.Go("ignore", "com.duosoftware.auth", "users").StoreObject().WithKeyField("EmailAddress").AndStoreOne(u).Ok()
		} else {
			if update {
				u.UserID = uList.UserID
				u.Password = common.GetHash(u.Password)
				u.ConfirmPassword = common.GetHash(u.Password)
				term.Write("SaveUser saving user  "+u.Name+" Update User "+u.UserID, term.Debug)
				client.Go("ignore", "com.duosoftware.auth", "users").StoreObject().WithKeyField("EmailAddress").AndStoreOne(u).Ok()
			}
		}
		//} else {
		//term.Write("SaveUser saving user store Error #"+err.Error(), term.Error)
		//}
	} else {
		term.Write("SaveUser saving user fetech Error #"+err, term.Error)
	}
	u.Password = "*****"
	u.ConfirmPassword = "*****"
	return u
}

func (h *AuthHandler) UserActivation(token string) bool {
	//respond := ""
	//check user from db
	bytes, err := client.Go("ignore", "com.duosoftware.com", "activation").GetOne().ByUniqueKey(token).Ok()
	if err == "" {
		var uList ActivationEmail
		err := json.Unmarshal(bytes, &uList)
		if err == nil || bytes == nil {
			//new user
			if err != nil {

				term.Write("Token Not Found", term.Debug)
				return false

			} else {
				//uList[0].GUUserID
				var u User
				var inputParams map[string]string
				inputParams = make(map[string]string)
				inputParams["email"] = u.EmailAddress
				inputParams["name"] = u.Name
				//Change activation status to true and save
				term.Write("Activate User  "+u.Name+" Update User "+u.UserID, term.Debug)
				email.Send("ignore", "com.duosoftware.auth", "email", "user_activated", inputParams, u.EmailAddress)
				return true
			}
		}

	} else {
		term.Write("Activation Fail ", term.Debug)
		return false

	}
	return false
}

func (h *AuthHandler) Login(email, password string) (User, string) {
	term.Write("Login  user  email"+email, term.Debug)
	term.Write(Config.UserName, term.Debug)

	bytes, err := client.Go("ignore", "com.duosoftware.auth", "users").GetOne().ByUniqueKey(email).Ok()
	var user User
	if err == "" {
		if bytes != nil {
			var uList User
			err := json.Unmarshal(bytes, &uList)
			if err == nil {
				//fmt.Println();
				if uList.Password == common.GetHash(password) && strings.ToLower(uList.EmailAddress) == strings.ToLower(email) {
					return uList, ""
				} else {
					term.Write("password incorrect", term.Error)
				}
			} else {
				if err != nil {
					term.Write("Login  user Error "+err.Error(), term.Error)
				}
			}
		}
	} else {
		term.Write("Login  user  Error "+err, term.Error)
	}

	return user, "Error Validating user"
}

func (h *AuthHandler) GetUser(email string) (User, string) {
	term.Write("Login  user  email"+email, term.Debug)
	term.Write(Config.UserName, term.Debug)

	bytes, err := client.Go("ignore", "com.duosoftware.auth", "users").GetOne().ByUniqueKey(email).Ok()
	var user User
	if err == "" {
		if bytes != nil {
			var uList User
			err := json.Unmarshal(bytes, &uList)
			if err == nil {
				uList.Password = "-------------"
				uList.ConfirmPassword = "-------------"
				return uList, ""
			} else {
				if err != nil {
					term.Write("Login  user Error "+err.Error(), term.Error)
				}
			}
		}
	} else {
		term.Write("Login  user  Error "+err, term.Error)
	}

	return user, "Error Validating user"
}

func SendNotification(u User, Message string) {

}
