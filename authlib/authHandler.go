package authlib

import (
	"duov6.com/applib"
	"duov6.com/common"
	//"duov6.com/email"
	notifier "duov6.com/duonotifier/client"
	"duov6.com/session"
	//"duov6.com/config"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"duov6.com/objectstore/client"
	"duov6.com/term"
)

// A AuthHandler represents a Method collection for Auth
type AuthHandler struct {
	//Config AuthConfig
}

// newAuthHandler will create a new AuthHandler
func newAuthHandler() *AuthHandler {
	authhld := new(AuthHandler) // Create new Object
	//authhld.Config = GetConfig()
	return authhld // Return new Object
}

// A ActivationEmail represents Access tokens for Email activations
type ActivationEmail struct {
	GUUserID string // GUUserID
	Token    string // Token for the email actiavte form
}

//A LoginAttemts represent Tracking login atttemts
type LoginAttemts struct {
	Email          string
	Domain         string
	Count          int
	LastAttemttime string
	BlockUser      string
}

//A LoginSessions represents tracking login sessions
type LoginSessions struct {
	Email  string
	Domain string
	Count  int64
}

// AppAutherize Autherize the application for the user
func (h *AuthHandler) AppAutherize(ApplicationID, UserID, Domain string) bool {
	bytes, err := client.Go("ignore", Domain, "atherized10x564xv").GetOne().ByUniqueKey(ApplicationID + "-" + UserID).Ok() // fetech user autherized
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

// CanLogin checked if the user can login
func (h *AuthHandler) CanLogin(email, domain string) (bool, string) {
	o, m := h.CheckLoginConcurrency(email)
	if !o {
		return o, m
	}

	bytes, err := client.Go("ignore", "com.duosoftware.auth", "loginAttemts").GetOne().ByUniqueKey(email).Ok() // fetech user autherized
	term.Write("CanLogin For Login "+email+" Domain "+domain, term.Debug)
	if err == "" {
		if bytes != nil {
			var uList LoginAttemts
			err := json.Unmarshal(bytes, &uList)
			if err == nil {
				if uList.BlockUser == "block" {
					return false, "User is blocked by the adminstrator"
				}
				if uList.Count >= 5 {
					Ttime1, _ := time.Parse("2006-01-02 15:04:05", uList.LastAttemttime)
					Ttime2 := time.Now().UTC()
					difference := Ttime1.Sub(Ttime2)
					minutesTime := difference.Minutes()
					if minutesTime <= 0 {
						h.RemoveAttemts(uList)
						return true, ""
					} else {
						m := strconv.FormatFloat(difference.Minutes(), 'f', 6, 64)
						//s := strconv.FormatFloat(difference.Seconds(), 'f', 6, 64)
						return false, "User account is locked try again in " + m + " Minutes"
					}
				} else {
					return true, ""
				}
			}
		}
	} else {
		term.Write("CanLogin Error "+err, term.Error)
	}
	return true, ""
}

//CheckLoginConcurrency helps to check and block the concurrent user logins
func (h *AuthHandler) CheckLoginConcurrency(email string) (bool, string) {
	if Config.NumberOFUserLogins != 0 {
		bytes, err := client.Go("ignore", "com.duosoftware.auth", "loginsessions").GetOne().ByUniqueKey(email).Ok() // fetech user autherized
		term.Write("CanLogin For Login "+email+" Domain ", term.Debug)
		if err == "" {
			if bytes != nil {
				var uList LoginSessions
				err := json.Unmarshal(bytes, &uList)
				if err == nil {
					if uList.Count >= Config.NumberOFUserLogins {
						return false, "Login Exceeeded please logout your sessions."
					}
				}
			}
		}

	}

	return true, ""
}

// Release will release the blocked users
func (h *AuthHandler) Release(email string) {
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "loginAttemts").GetOne().ByUniqueKey(email).Ok() // fetech user autherized
	term.Write("CanLogin For Login "+email+" Domain ", term.Debug)
	if err == "" {
		if bytes != nil {
			var uList LoginAttemts
			err := json.Unmarshal(bytes, &uList)
			if err == nil {
				h.RemoveAttemts(uList)
			}
		}
	}

}

func (a *AuthHandler) RemoveAttemts(Attemt LoginAttemts) {
	fmt.Println(Attemt)
	//if Attemt.BlockUser != "block" {
	client.Go("ignore", "com.duosoftware.auth", "loginAttemts").DeleteObject().WithKeyField("Email").AndDeleteObject(Attemt).Ok()
	//}

}

func (a *AuthHandler) LogFailedAttemts(email, domain, blockstatus string) {
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "loginAttemts").GetOne().ByUniqueKey(email).Ok() // fetech user autherized
	var uList LoginAttemts
	uList.Email = email
	uList.Domain = domain
	uList.Count = 1
	uList.BlockUser = blockstatus
	term.Write("LogFailedAttemts For Login "+email+" Domain "+domain, term.Debug)
	if err == "" {
		if bytes != nil {
			var x LoginAttemts
			fmt.Println("Attem")
			err := json.Unmarshal(bytes, &x)
			fmt.Println(err)
			fmt.Println(string(bytes))
			if err == nil {
				fmt.Println(x)
				x.Count = x.Count + 1
				//x.LastAttemttime = ""
				uList = x
			}
		}
	}

	nowTime := time.Now().UTC()
	nowTime = nowTime.Add(3 * time.Minute)
	uList.LastAttemttime = nowTime.Format("2006-01-02 15:04:05")
	fmt.Println(uList)
	client.Go("ignore", "com.duosoftware.auth", "loginAttemts").StoreObject().WithKeyField("Email").AndStoreOne(uList).Ok()

}

func (a *AuthHandler) LogLoginSessions(email, domain string, item int64) {
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "loginsessions").GetOne().ByUniqueKey(email).Ok() // fetech user autherized
	var uList LoginSessions
	uList.Email = email
	uList.Domain = domain
	uList.Count = item
	//uList.BlockUser = blockstatus
	term.Write("LogLoginSessions For Login "+email+" Domain "+domain, term.Debug)
	if err == "" {
		if bytes != nil {
			var x LoginSessions
			fmt.Println("Attem")
			err := json.Unmarshal(bytes, &x)
			fmt.Println(err)
			fmt.Println(string(bytes))
			if err == nil {
				fmt.Println(x)
				x.Count = x.Count + item
				//x.LastAttemttime = ""
				uList = x
			}
		}
	}

	//nowTime := time.Now().UTC()
	//nowTime = nowTime.Add(3 * time.Minute)
	//uList.LastAttemttime = nowTime.Format("2006-01-02 15:04:05")
	fmt.Println(uList)
	client.Go("ignore", "com.duosoftware.auth", "loginsessions").StoreObject().WithKeyField("Email").AndStoreOne(uList).Ok()

}

// GetAuthCode helps to get the Code to authendicate and add wait for the authendications
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

// AutherizeApp autherize apps using the secret key that the application provided
func (h *AuthHandler) AutherizeApp(Code, ApplicationID, AppSecret, UserID, SecurityToken, Domain string) (bool, string) {
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "authcode").GetOne().ByUniqueKey(Code).Ok()
	term.Write("AutherizeApp For ApplicationID "+ApplicationID+" Code "+Code+" Secret "+AppSecret+" Err "+err, term.Debug)
	var uList AuthCode
	err1 := json.Unmarshal(bytes, &uList)
	term.Write(string(bytes[:]), term.Debug)
	if err1 == nil {

		var appH applib.Apphanler
		application, err := appH.Get(ApplicationID, SecurityToken)
		if err == "" {
			if application.SecretKey == AppSecret && uList.UserID == UserID && Code == uList.Code {
				var appAth AppAutherize
				appAth.AppliccatioID = ApplicationID
				appAth.AutherizeKey = ApplicationID + "-" + UserID
				appAth.Name = application.Name
				client.Go("ignore", Domain, "atherized10x564xv").StoreObject().WithKeyField("AutherizeKey").AndStoreOne(appAth).Ok()
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

// AddSession helps to keep the session
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
	c.Otherdata = make(map[string]string)
	c.Otherdata = a.Otherdata
	if Config.ExpairyTime > 0 {
		nowTime := time.Now().UTC()
		nowTime = nowTime.Add(time.Duration(Config.ExpairyTime) * time.Minute)
		c.Otherdata["expairyTime"] = nowTime.Format("2006-01-02 15:04:05")
	}
	session.AddSession(c)
}

// LogOut make you logout,
func (h *AuthHandler) LogOut(a AuthCertificate) {
	//client.Go("ignore", "s.duosoftware.auth", "sessions").DeleteObject().ByUniqueKey(a.SecurityToken)
	client.Go("ignore", "s.duosoftware.auth", "sessions").DeleteObject().WithKeyField("SecurityToken").AndDeleteObject(a).Ok()
	//client.Go("ignore", "s.duosoftware.auth", "sessions").StoreObject().WithKeyField("SecurityToken").AndStoreOne(a).Ok()
	h.LogoutClildSessions(a.SecurityToken)
	if Config.NumberOFUserLogins != 0 {
		h.LogLoginSessions(a.Email, a.Domain, -1)
	}
	term.Write("LogOut for "+a.Name+" with SecurityToken :"+a.SecurityToken, term.Debug)
	//h.Release(a.Email)
	//return true
}

func (h *AuthHandler) LogoutClildSessions(SecurityToken string) {
	s := session.GetChildSession(SecurityToken)
	for _, a := range s {
		client.Go("ignore", "s.duosoftware.auth", "sessions").DeleteObject().WithKeyField("SecurityToken").AndDeleteObject(a).Ok()
		term.Write("LogOut for "+a.Name+" with SecurityToken :"+a.SecurityToken, term.Debug)
		h.LogoutClildSessions(a.SecurityToken)
	}
}

/*
func SetIlligalAttemts(clientIP, UserAgent, key string) {
	keyfile := make(map[string]interface{})
	keyfile["key"] = key
	keyfile["clientip"] = clientIP
	keyfile["secret"] = UserAgent
	keyfile["attemts"] = 0
	bytes, _ := client.Go("ignore", "com.duosoftware.auth", "attemts").GetOne().ByUniqueKey(key).Ok()
	if bytes != nil {
		err := json.Unmarshal(bytes, &keyfile)
	}
	t1, e := time.Parse(
		time.RFC3339,
		keyfile["lastattemt"])

	t := time.Now()

	keyfile["lastattemt"] = t.Format(time.RFC3339)
	keyfile["attemts"] = keyfile["attemts"] + 1
	client.Go("ignore", "com.duosoftware.auth", "keysecrets").StoreObject().WithKeyField("key").AndStoreOne(keyfile).Ok()
	//return keyfile["secret"]
}

func IsIlligale(clientIP, key string) bool {
	keyfile := make(map[string]string)
	keyfile["key"] = key
	//keyfile["clientip"] = clientIP
	//keyfile["secret"] = UserAgent
	keyfile["attemts"] = 0
	keyfile["lastattemt"] = t.Format(time.RFC3339)
	bytes, _ := client.Go("ignore", "com.duosoftware.auth", "attemts").GetOne().ByUniqueKey(key).Ok()
	if bytes != nil {
		err := json.Unmarshal(bytes, &keyfile)
	}
	t1, e := time.Parse(
		time.RFC3339,
		keyfile["lastattemt"])

}*/

// GetSession helps to get the session
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
		bytes, _ := client.Go("ignore", a.Domain, "scope").GetOne().ByUniqueKey(a.Domain).Ok() // fetech user autherized
		//term.Write("AppAutherize For Application "+ApplicationID+" UserID "+UserID, term.Debug)
		c.DataCaps = strings.Replace(string(bytes[:]), "\"", "`", -1)
		payload := common.JWTPayload(a.Domain, c.SecurityToken, c.UserID, c.Email, c.Domain, bytes)

		if a.Otherdata["expairyTime"] != "" {
			Ttime1, _ := time.Parse("2006-01-02 15:04:05", a.Otherdata["expairyTime"])
			Ttime2 := time.Now().UTC()
			difference := Ttime1.Sub(Ttime2)
			minutesTime := difference.Minutes()
			if minutesTime <= 0 {
				h.LogOut(c)
				return AuthCertificate{}, "Session Expaired."
			}
		}

		if a.Otherdata["JWT"] == "" {
			c.Otherdata = make(map[string]string)
			c.Otherdata["JWT"] = common.Jwt(h.GetSecretKey(a.Domain), payload)
			c.Otherdata["Scope"] = strings.Replace(string(bytes[:]), "\"", "`", -1)
			a.Otherdata["JWT"] = c.Otherdata["JWT"]
			a.Otherdata["Scope"] = c.Otherdata["Scope"]
			if Config.ExpairyTime > 0 {
				nowTime := time.Now().UTC()
				nowTime = nowTime.Add(time.Duration(Config.ExpairyTime) * time.Minute)
				a.Otherdata["expairyTime"] = nowTime.Format("2006-01-02 15:04:05")
			}
			session.AddSession(a)
		} else {
			c.Otherdata = make(map[string]string)
			c.Otherdata = a.Otherdata
		}
		//string(bytes[:])
		return c, ""
	} else {
		term.Write("GetSession Error "+err, term.Error)
	}
	term.Write("GetSession No Session for SecurityToken "+key, term.Debug)
	return c, "Error Session Not Found"
}

func (h *AuthHandler) GetSecretKey(key string) string {
	keyfile := make(map[string]interface{})
	bytes, _ := client.Go("ignore", "com.duosoftware.auth", "keysecrets").GetOne().ByUniqueKey(key).Ok()
	if bytes != nil {
		err := json.Unmarshal(bytes, &keyfile)
		if err == nil {
			return keyfile["secret"].(string)
		}
	}

	keyfile["key1"] = key
	keyfile["secret"] = common.GetGUID()
	client.Go("ignore", "com.duosoftware.auth", "keysecrets").StoreObject().WithKeyField("key1").AndStoreOne(keyfile).Ok()
	return keyfile["secret"].(string)
}

// ForgetPassword to help the user to reset password
func (h *AuthHandler) ForgetPassword(emailaddress string) bool {
	u, error := h.GetUser(emailaddress)
	if error == "" {
		if u.Active {
			passowrd := common.RandText(6)
			u.ConfirmPassword = passowrd
			u.Password = passowrd
			term.Write("Password : "+passowrd, term.Debug)
			h.SaveUser(u, true, "forgotpassword")
			var inputParams map[string]string
			inputParams = make(map[string]string)
			// inputParams["@@email@@"] = u.EmailAddress
			// inputParams["@@name@@"] = u.Name
			// inputParams["@@password@@"] = passowrd
			// go email.Send("ignore", "Password Recovery.", "com.duosoftware.auth", "email", "user_resetpassword", inputParams, nil, u.EmailAddress)
			inputParams["@@CEMAIL@@"] = u.EmailAddress
			inputParams["@@CNAME@@"] = u.Name
			inputParams["@@@PASSWORD@@@"] = passowrd
			//go notifier.Send("ignore", "Password Recovery.", "com.duosoftware.auth", "email", "T_Email_FORGETPW", inputParams, nil, u.EmailAddress)
			go notifier.Notify("ignore", "FORGETPW", u.EmailAddress, inputParams, nil)
			term.Write("E Mail Sent", term.Debug)
			return true
		} else {
			term.Write("This User is not yet activated.. Cannot reset password!", term.Debug)
			return false
		}
	}
	return false
}

// ChangePassword Changes the password
func (h *AuthHandler) ChangePassword(a AuthCertificate, newPassword string) bool {
	u, error := h.GetUser(a.Email)
	if error == "" {
		//passowrd := common.RandText(6)
		u.ConfirmPassword = newPassword
		u.Password = newPassword
		h.SaveUser(u, true, "changepassword")
		return true
	}
	return false
}

// SaveUser helps to save the users
func (h *AuthHandler) SaveUser(u User, update bool, regtype string) (User, string) {
	term.Write("SaveUser saving user  "+u.Name, term.Debug)
	u.EmailAddress = strings.ToLower(u.EmailAddress)
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "users").GetOne().ByUniqueKey(u.EmailAddress).Ok()
	if err == "" {
		var uList User
		err := json.Unmarshal(bytes, &uList)
		//if err == nil || bytes == nil {
		term.Write("SaveUser saving user retrived", term.Debug)
		//fmt.Println(uList)
		term.Write("SaveUser saving user retrived", term.Debug)
		if err != nil || uList.UserID == "" {
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
			// inputParams["@@email@@"] = u.EmailAddress
			// inputParams["@@name@@"] = u.Name
			// inputParams["@@token@@"] = Activ.Token
			// inputParams["@@password@@"] = password
			inputParams["@@CEMAIL@@"] = u.EmailAddress
			inputParams["@@CNAME@@"] = u.Name

			//go notifier.Send("ignore", "Thank you for registering!", "com.duosoftware.auth", "email", "T_Email_Verification", inputParams, nil, u.EmailAddress)

			switch regtype {
			case "tenant":
				inputParams["@@PASSWORD@@"] = password
				u.Active = true
				term.Write("SaveUser saving user for tenat "+u.Name+" Update User "+u.UserID, term.Debug)
				go notifier.Notify("ignore", "TenantUser_Verification", u.EmailAddress, inputParams, nil)
				break
			default:
				inputParams["@@CODE@@"] = Activ.Token

				go notifier.Notify("ignore", "Verification", u.EmailAddress, inputParams, nil)
				break
			}
			term.Write("E Mail Sent", term.Debug)
			client.Go("ignore", "com.duosoftware.auth", "activation").StoreObject().WithKeyField("Token").AndStoreOne(Activ).Ok()
			term.Write("Activation stored", term.Debug)
			client.Go("ignore", "com.duosoftware.auth", "users").StoreObject().WithKeyField("EmailAddress").AndStoreOne(u).Ok()
			u.Password = "*****"
			u.ConfirmPassword = "*****"
			return u, ""
		} else {
			if update {
				u.UserID = uList.UserID
				u.Password = common.GetHash(u.Password)
				u.ConfirmPassword = common.GetHash(u.Password)
				term.Write("SaveUser saving user  "+u.Name+" Update User "+u.UserID, term.Debug)
				client.Go("ignore", "com.duosoftware.auth", "users").StoreObject().WithKeyField("EmailAddress").AndStoreOne(u).Ok()
				u.Password = "*****"
				u.ConfirmPassword = "*****"
				return u, ""
			} else {
				return u, "Already Registered."
			}
		}
	} else {

		term.Write("SaveUser saving user fetech Error #"+err, term.Error)
		return u, err
	}
	u.Password = "*****"
	u.ConfirmPassword = "*****"
	return u, "Error User Registered."
}

// UserActivation Helps to activate the users
func (h *AuthHandler) UserActivation(token string) bool {
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "activation").GetOne().ByUniqueKey(token).Ok()
	if err == "" {
		var uList ActivationEmail
		err := json.Unmarshal(bytes, &uList)
		if err == nil {
			//new user

			//uList[0].GUUserID

			//var u User
			u, _ := h.GetUser(uList.GUUserID)
			var inputParams map[string]string
			inputParams = make(map[string]string)
			inputParams["@@email@@"] = u.EmailAddress
			inputParams["@@name@@"] = u.Name
			//Change activation status to true and save

			term.Write(u, term.Debug)

			if u.Active {
				term.Write("This User : "+u.EmailAddress+" is already activated!", term.Debug)
				return true
			} else {
				u.Active = true
				client.Go("ignore", "com.duosoftware.auth", "users").StoreObject().WithKeyField("EmailAddress").AndStoreOne(u).Ok()
				//h.SaveUser(u, true)
				term.Write("Activate User  "+u.Name+" Update User "+u.UserID, term.Debug)
				//go notifier.Send("ignore", "User Activation.", "com.duosoftware.auth", "email", "user_activated", inputParams, nil, u.EmailAddress)
				go notifier.Notify("ignore", "user_activated", u.EmailAddress, inputParams, nil)
				return true
			}
		} else {
			term.Write(err, term.Debug)
			term.Write(string(bytes), term.Debug)
		}

	} else {
		term.Write("Activation Fail ", term.Debug)
		term.Write(err, term.Debug)
		return false

	}
	return false
}

// Login helps to authedicate the users
func (h *AuthHandler) Login(email, password string) (User, string) {
	term.Write("Login  user  email"+email, term.Debug)
	term.Write(Config.UserName, term.Debug)
	email = strings.ToLower(email)
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "users").GetOne().ByUniqueKey(email).Ok()
	//fmt.Println(string(bytes))
	var user User
	if err == "" {
		if bytes != nil {
			var uList User
			err := json.Unmarshal(bytes, &uList)
			if err == nil {
				//fmt.Println(uList)
				if uList.Password == common.GetHash(password) && strings.ToLower(uList.EmailAddress) == strings.ToLower(email) {
					if uList.Active {
						return uList, ""
					} else {
						return user, "Email Address is not verified."
						//return user, "Email Address is not varified."
					}
				} else {
					term.Write("Username password incorrect", term.Error)
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
	return user, "The username or password is incorrect. Please try again with the correct credentials. 3 failed attempts will temporarily block the account."
	//return user, "Username password incorrect"
}

func (h *AuthHandler) GetUserByID(UserID string) (User, string) {
	term.Write("Login  user  UID"+UserID, term.Debug)
	//term.Write(Config.UserName, term.Debug)
	//email = strings.ToLower(email)
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "users").GetOne().BySearching("UserID:" + UserID).Ok()
	var user User
	if err == "" {
		if bytes != nil {
			var uList User
			err := json.Unmarshal(bytes, &uList)
			if err == nil {
				//uList.Password = "-------------"
				//uList.ConfirmPassword = "-------------"
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

// GetUser elps to retrive the User
func (h *AuthHandler) GetUser(email string) (User, string) {
	term.Write("Login  user  email"+email, term.Debug)
	term.Write(Config.UserName, term.Debug)
	email = strings.ToLower(email)
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "users").GetOne().ByUniqueKey(email).Ok()
	var user User
	if err == "" {
		if bytes != nil {
			var uList User
			err := json.Unmarshal(bytes, &uList)
			if err == nil {
				//uList.Password = "-------------"
				//uList.ConfirmPassword = "-------------"
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

func (h *AuthHandler) GetMultipleUserDetails(UserIDs []string) (users []map[string]interface{}) {
	users = make([]map[string]interface{}, 0)

	for x := 0; x < len(UserIDs); x++ {
		bytes, err := client.Go("ignore", "com.duosoftware.auth", "users").GetMany().BySearching("UserID:" + UserIDs[x]).Ok()
		if err == "" {
			if bytes != nil {
				var uList []User
				err := json.Unmarshal(bytes, &uList)
				if err == nil {
					singleUser := make(map[string]interface{})
					singleUser["UserID"] = uList[0].UserID
					singleUser["Name"] = uList[0].Name
					singleUser["EmailAddress"] = uList[0].EmailAddress
					users = append(users, singleUser)
				}
			}
		}
	}

	return users
}

func SendNotification(u User, Message string) {

}
