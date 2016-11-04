package authlib

import (
	//"duov6.com/applib"
	"encoding/json"

	"duov6.com/common"
	notifier "duov6.com/duonotifier/client"
	"duov6.com/gorest"
	"duov6.com/objectstore/client"
	"runtime"
	//"fmt"
	//"golang.org/x/oauth2"
	//"crypto/hmac"
	"duov6.com/session"
	"duov6.com/term"
	"strconv"
	"strings"
	///"strings"
)

type AuthCertificate struct {
	UserID, Username, Name, Email, SecurityToken, Domain, DataCaps, ClientIP string
	Otherdata                                                                map[string]string
}

type AuthorizeAppData struct {
	Object map[string]interface{}
}

type Auth struct {
	gorest.RestService
	verify           gorest.EndPoint `method:"GET" path:"/" output:"string"`
	login            gorest.EndPoint `method:"GET" path:"/Login/{username:string}/{password:string}/{domain:string}" output:"AuthCertificate"`
	noPasswordLogin  gorest.EndPoint `method:"GET" path:"/NoPasswordLogin/{OTP:string}" output:"AuthCertificate"`
	loginOTP         gorest.EndPoint `method:"GET" path:"/LoginOTP/{username:string}/{password:string}/{domain:string}" output:"string"`
	loginOTPNoPass   gorest.EndPoint `method:"GET" path:"/LoginOTPNoPass/{username:string}/{domain:string}" output:"string"`
	getLoginSessions gorest.EndPoint `method:"GET" path:"/GetLoginSessions/{UserID:string}" output:"[]AuthCertificate"`
	authorize        gorest.EndPoint `method:"GET" path:"/Authorize/{SecurityToken:string}/{ApplicationID:string}" output:"AuthCertificate"`
	getSession       gorest.EndPoint `method:"GET" path:"/GetSession/{SecurityToken:string}/{Domain:string}" output:"AuthCertificate"`
	getSessionStatic gorest.EndPoint `method:"GET" path:"/GetSessionStatic/{SecurityToken:string}" output:"AuthCertificate"`
	getSecret        gorest.EndPoint `method:"GET" path:"/GetSecret/{Key:string}" output:"string"`
	getAuthCode      gorest.EndPoint `method:"GET" path:"/GetAuthCode/{SecurityToken:string}/{ApplicationID:string}/{URI:string}" output:"string"`
	//Lasith's method - Don't Delete
	//autherizeApp       gorest.EndPoint `method:"GET" path:"/AutherizeApp/{SecurityToken:string}/{Code:string}/{ApplicationID:string}/{AppSecret:string}" output:"bool"`
	autherizeApp            gorest.EndPoint `method:"POST" path:"/AutherizeApp/{SecurityToken:string}/{Code:string}/{ApplicationID:string}/{AppSecret:string}" postdata:"AuthorizeAppData"`
	updateScope             gorest.EndPoint `method:"POST" path:"/UpdateScope/{SecurityToken:string}/{UserID:string}/{ApplicationID:string}" postdata:"AuthorizeAppData"`
	addUser                 gorest.EndPoint `method:"POST" path:"/UserRegistation/" postdata:"User"`
	invitedUserRegistration gorest.EndPoint `method:"POST" path:"/InvitedUserRegistration/" postdata:"User"`
	registerTenantUser      gorest.EndPoint `method:"POST" path:"/RegisterTenantUser/" postdata:"User"`
	userActivation          gorest.EndPoint `method:"GET" path:"/UserActivation/{token:string}" output:"bool"`
	logOut                  gorest.EndPoint `method:"GET" path:"/LogOut/{SecurityToken:string}" output:"bool"`
	checkPassword           gorest.EndPoint `method:"GET" path:"/Checkpassword/{SecurityToken:string}" output:"bool"`
	getUser                 gorest.EndPoint `method:"GET" path:"/GetUser/{Email:string}" output:"User"`
	blockUser               gorest.EndPoint `method:"GET" path:"/BlockUser/{Email:string}" output:"bool"`
	releaseUser             gorest.EndPoint `method:"GET" path:"/ReleaseUser/{Email:string}/{b4:string}" output:"bool"`
	getGUID                 gorest.EndPoint `method:"GET" path:"/GetGUID/" output:"string"`
	forgotPassword          gorest.EndPoint `method:"GET" path:"/ForgotPassword/{EmailAddress:string}/{RequestCode:string}" output:"bool"`
	changePassword          gorest.EndPoint `method:"GET" path:"/ChangePassword/{OldPassword:string}/{NewPassword:string}" output:"bool"`
	arbiterAuthorize        gorest.EndPoint `method:"POST" path:"/ArbiterAuthorize/" postdata:"map[string]string"`
	getUserByUserId         gorest.EndPoint `method:"POST" path:"/GetUserByUserID/" postdata:"[]string"`
}

//GetClientIP Represent to get ClientIP
func GetClientIP() string {
	//hmac.New(hash.)
	return "hope"
}

//GetDataCaps Represent to getting datacaps
func GetDataCaps(Domain, UserID string) string {
	return "#" + Domain + "#" + UserID + "#1#2#4"
}

//UserActivation Represent activation of the user account
func (A Auth) UserActivation(token string) bool {
	h := newAuthHandler()
	status := h.UserActivation(token)
	if status == "alreadyActivated" {
		A.ResponseBuilder().SetResponseCode(300)
		return true
	} else if status == "true" {
		A.ResponseBuilder().SetResponseCode(200)
		return true
	} else if status == "false" {
		A.ResponseBuilder().SetResponseCode(500)
		return false
	}
	return false
}

func (A Auth) GetLoginSessions(UserID string) []session.AuthCertificate {
	return session.GetRunningSession(UserID)
}

/*func (A Auth) ForceLogout(UserID string) {

}*/

func (A Auth) GetUserByUserId(object []string) {
	h := AuthHandler{}
	userDetails := h.GetMultipleUserDetails(object)
	objectByteArray, _ := json.Marshal(userDetails)
	A.ResponseBuilder().SetResponseCode(200).WriteAndOveride(objectByteArray)
	return
}

func (A Auth) LogOut(SecurityToken string) bool {
	h := newAuthHandler()

	c, err := h.GetSession(SecurityToken, "Nil")
	if err == "" {
		go h.LogOut(c)
		return true
	}

	A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Session or Application not exist.")))

	return false
}

func (A Auth) ForgotPassword(EmailAddress, RequestCode string) bool {
	h := newAuthHandler()
	return h.ForgetPassword(EmailAddress)

}

func (A Auth) ChangePassword(OldPassword, NewPassword string) bool {
	h := newAuthHandler()
	user, error := h.GetSession(A.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		_, err := h.Login(user.Email, OldPassword)
		if err != "" {
			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Wrong Current Password.")))
			return false
		}
		return h.ChangePassword(user, NewPassword)
	} else {
		return false
	}
}

func (A Auth) Verify() (output string) {
	//output = "{\"name\": \"DuoAuth\",\"version\": \"6.0.24-a\",\"Change Log\":\"Added Check for tenant subscription invitation.\",\"author\": {\"name\": \"Duo Software\",\"url\": \"http://www.duosoftware.com/\"},\"repository\": {\"type\": \"git\",\"url\": \"https://github.com/DuoSoftware/v6engine/\"}}"
	cpuUsage := strconv.Itoa(int(common.GetProcessorUsage()))
	cpuCount := strconv.Itoa(runtime.NumCPU())

	versionData := make(map[string]interface{})
	versionData["API Name"] = "Duo Auth"
	versionData["API Version"] = "6.1.08"

	changeLogs := make(map[string]interface{})

	changeLogs["6.1.08"] = [...]string{
		"Added user deny check",
		"Added User Deactivate if user has no accesible tenants.",
	}

	changeLogs["6.1.07"] = [...]string{
		"Added Activation Skip Endpoint for Registration. <InvitedUserRegistration>",
	}

	changeLogs["6.1.06"] = [...]string{
		"Commented SecurityToken from AcceptRequest",
		"Added response codes for ActivateUser method",
	}

	changeLogs["6.1.05"] = [...]string{
		"Added New Login password,username message and Activate message",
		"Added GetTenantAdmin method for auth",
		"Removed rating engine check for tenant add.",
	}

	changeLogs["6.1.04"] = [...]string{
		"Added Activate User Email Check..",
		"Added Reset Password Check by checking user activated or not",
	}

	versionData["Change Logs"] = changeLogs

	gitMap := make(map[string]string)
	gitMap["Type"] = "git"
	gitMap["URL"] = "https://github.com/DuoSoftware/v6engine/"
	versionData["Repository"] = gitMap

	statMap := make(map[string]string)
	statMap["CPU"] = cpuUsage + " (percentage)"
	statMap["CPU Cores"] = cpuCount
	versionData["System Usage"] = statMap

	authorMap := make(map[string]string)
	authorMap["Name"] = "Duo Software Pvt Ltd"
	authorMap["URL"] = "http://www.duosoftware.com/"
	versionData["Project Author"] = authorMap

	byteArray, _ := json.Marshal(versionData)
	output = string(byteArray)
	return
}

func (A Auth) ArbiterAuthorize(object map[string]string) {
	var outCrt AuthCertificate
	issue := object["authority"]
	th := TenantHandler{}
	//th.Autherized(domain, user)

	switch issue {
	case "auth0":
		ah := auth0{}
		c, err := ah.RegisterToken(object)
		if err != "" {
			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(err)))
			return
		} else {
			outCrt = c
		}
	case "FaceBook":
		ah := facebookAuth{}
		c, err := ah.RegisterToken(object)
		if err != "" {
			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(err)))
			return
		} else {
			outCrt = c
		}
		break
	case "twitter":
		ah := twitterAuth{}
		c, err := ah.RegisterToken(object)
		if err != "" {
			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(err)))
			return
		} else {
			outCrt = c
		}
		break
	case "googleplus":
		ah := googlePlusAuth{}
		c, err := ah.RegisterToken(object)
		if err != "" {
			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(err)))
			return
		} else {
			outCrt = c
		}
		break
	default:
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Unautherized Arbiter Form.")))
		return
		break
	}
	x, _ := th.AutherizedUser(outCrt.Domain, outCrt.UserID)
	if !x {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(outCrt.Domain + " Is not autherized for signin.")))
		//A.Context.Request().
		return
	}
	if A.Context.Request().Header.Get("PHP") != "101" {
		outCrt.ClientIP = A.Context.Request().RemoteAddr
	} else {
		outCrt.ClientIP = A.Context.Request().Header.Get("IP")
	}
	h := AuthHandler{}
	outCrt.Otherdata["UserAgent"] = A.Context.Request().UserAgent()
	bytes, _ := client.Go("ignore", outCrt.Domain, "scope").GetOne().ByUniqueKey(outCrt.Domain).Ok() // fetech user autherized
	//term.Write("AppAutherize For Application "+ApplicationID+" UserID "+UserID, term.Debug)
	outCrt.DataCaps = string(bytes[:])
	payload := common.JWTPayload(outCrt.Domain, outCrt.SecurityToken, outCrt.UserID, outCrt.Email, outCrt.Domain, bytes)
	outCrt.Otherdata["JWT"] = common.Jwt(h.GetSecretKey(outCrt.Domain), payload)
	outCrt.Otherdata["Scope"] = strings.Replace(string(bytes[:]), "\"", "`", -1)
	//outCrt.Otherdata["Tempkey"] = "No"
	//th := TenantHandler{}
	tlist := th.GetTenantsForUser(outCrt.UserID)
	b, _ := json.Marshal(tlist)
	outCrt.Otherdata["TenentsAccessible"] = strings.Replace(string(b[:]), "\"", "`", -1)

	h.AddSession(outCrt)
	var inputParams map[string]string
	inputParams = make(map[string]string)
	inputParams["@@email@@"] = outCrt.Email
	inputParams["@@name@@"] = outCrt.Name
	inputParams["@@UserAgent@@"] = A.Context.Request().UserAgent()
	inputParams["@@ClientIP@@"] = outCrt.ClientIP
	inputParams["@@Domain@@"] = outCrt.Domain
	inputParams["@@SecurityToken@@"] = outCrt.SecurityToken
	//Change activation status to true and save
	//term.Write("Activate User  "+u.Name+" Update User "+u.UserID, term.Debug)
	//go notifier.Send("ignore", "User Login Notification.", "com.duosoftware.auth", "email", "user_login", inputParams, nil, outCrt.Email)
	go notifier.Notify("ignore", "user_login", outCrt.Email, inputParams, nil)
	f, _ := json.Marshal(outCrt)
	A.ResponseBuilder().SetResponseCode(200).WriteAndOveride(f)
	return

}

func (A Auth) Login(username, password, domain string) (outCrt AuthCertificate) {
	h := newAuthHandler()
	c, msg := h.CanLogin(username, domain)
	if !c {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(msg)))
		//A.Context.Request().
		return
	}
	u, err := h.Login(username, password)

	//if()
	if err == "" {
		//fmt.Println("login succeful")
		//securityToken := common.GetGUID()
		th := TenantHandler{}
		//th.Autherized(domain, user)
		x, _ := th.AutherizedUser(domain, u.UserID)
		if !x {
			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(domain + " Is not autherized for signin.")))
			//A.Context.Request().
			return
		}
		if A.Context.Request().Header.Get("PHP") != "101" {
			outCrt.ClientIP = A.Context.Request().RemoteAddr
		} else {
			outCrt.ClientIP = A.Context.Request().Header.Get("IP")
		}
		outCrt.DataCaps = GetDataCaps(domain, u.UserID)
		outCrt.Email = u.EmailAddress
		outCrt.UserID = u.UserID
		outCrt.Name = u.Name
		outCrt.SecurityToken = common.GetGUID()
		outCrt.Domain = domain
		outCrt.Username = u.EmailAddress
		outCrt.Otherdata = make(map[string]string)
		outCrt.Otherdata["UserAgent"] = A.Context.Request().UserAgent()
		bytes, _ := client.Go("ignore", domain, "scope").GetOne().ByUniqueKey(domain).Ok() // fetech user autherized
		//term.Write("AppAutherize For Application "+ApplicationID+" UserID "+UserID, term.Debug)
		outCrt.DataCaps = string(bytes[:])
		payload := common.JWTPayload(domain, outCrt.SecurityToken, outCrt.UserID, outCrt.Email, outCrt.Domain, bytes)
		outCrt.Otherdata["JWT"] = common.Jwt(h.GetSecretKey(domain), payload)
		outCrt.Otherdata["Scope"] = strings.Replace(string(bytes[:]), "\"", "`", -1)
		//outCrt.Otherdata["Tempkey"] = "No"
		//th := TenantHandler{}
		tlist := th.GetTenantsForUser(u.UserID)
		b, _ := json.Marshal(tlist)
		outCrt.Otherdata["TenentsAccessible"] = strings.Replace(string(b[:]), "\"", "`", -1)
		//outCrt = AuthCertificate{u.UserID, u.EmailAddress, u.Name, u.EmailAddress, securityToken, "http://192.168.0.58:9000/instaltionpath", "#0so0936#sdasd", "IPhere"}
		if Config.NumberOFUserLogins != 0 {
			h.LogLoginSessions(username, domain, 1)
		}
		h.AddSession(outCrt)
		var inputParams map[string]string
		inputParams = make(map[string]string)
		inputParams["@@email@@"] = u.EmailAddress
		inputParams["@@name@@"] = u.Name
		inputParams["@@UserAgent@@"] = A.Context.Request().UserAgent()
		inputParams["@@ClientIP@@"] = outCrt.ClientIP
		inputParams["@@Domain@@"] = domain
		inputParams["@@SecurityToken@@"] = outCrt.SecurityToken
		//Change activation status to true and save
		//term.Write("Activate User  "+u.Name+" Update User "+u.UserID, term.Debug)
		//go notifier.Send("ignore", "User Login Notification.", "com.duosoftware.auth", "email", "user_login", inputParams, nil, u.EmailAddress)
		go notifier.Notify("ignore", "user_login", u.EmailAddress, inputParams, nil)
		return
	}
	h.LogFailedAttemts(username, domain, "")
	A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(err)))
	//A.Context.Request().
	return
}

// NoPasswordLogin Represent nopassword
func (A Auth) NoPasswordLogin(OTP string) (outCrt AuthCertificate) {
	h := newAuthHandler()
	r := requestHandler{}
	o, err := r.GetRequestCode(OTP)
	if err == "" {
		outCrt.ClientIP = o["ClientIP"]
		outCrt.DataCaps = o["DataCaps"]
		outCrt.Email = o["Email"]
		outCrt.UserID = o["UserID"]
		outCrt.Name = o["Name"]
		outCrt.SecurityToken = o["SecurityToken"]
		outCrt.Domain = o["Domain"]
		outCrt.Username = o["Username"]
		outCrt.Otherdata = make(map[string]string)
		outCrt.Otherdata["UserAgent"] = o["UserAgent"]
		outCrt.Otherdata["JWT"] = o["JWT"]
		outCrt.Otherdata["Scope"] = o["Scope"]
		outCrt.Otherdata["TenentsAccessible"] = o["TenentsAccessible"]
		h.AddSession(outCrt)
		var inputParams map[string]string
		inputParams = make(map[string]string)
		inputParams["@@email@@"] = o["Email"]
		inputParams["@@name@@"] = o["Name"]
		inputParams["@@UserAgent@@"] = A.Context.Request().UserAgent()
		inputParams["@@ClientIP@@"] = outCrt.ClientIP
		inputParams["@@Domain@@"] = o["Domain"]
		inputParams["@@SecurityToken@@"] = outCrt.SecurityToken
		data := make(map[string]interface{})
		for key, value := range o {
			data[key] = value
		}
		r.Remove(data)
		//Change activation status to true and save
		//term.Write("Activate User  "+u.Name+" Update User "+u.UserID, term.Debug)
		//go notifier.Send("ignore", "User Login Notification.", "com.duosoftware.auth", "email", "user_login", inputParams, nil, o["Email"])
		go notifier.Notify("ignore", "user_login", o["Email"], inputParams, nil)
		return
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(err)))
		return
	}
}

func (A Auth) LoginOTP(username, password, domain string) string {
	h := newAuthHandler()
	c, msg := h.CanLogin(username, domain)
	if !c {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(msg)))
		//A.Context.Request().
		return common.ErrorJson(msg)
	}
	u, err := h.Login(username, password)

	//if()
	if err == "" {
		//fmt.Println("login succeful")
		//securityToken := common.GetGUID()
		r := requestHandler{}
		th := TenantHandler{}
		//th.Autherized(domain, user)
		x, _ := th.AutherizedUser(domain, u.UserID)
		if !x {
			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(domain + " Is not autherized for signin.")))
			//A.Context.Request().
			return common.ErrorJson(domain + " Is not autherized for signin.")
		}
		o := make(map[string]string)

		if A.Context.Request().Header.Get("PHP") != "101" {
			o["ClientIP"] = A.Context.Request().RemoteAddr
		} else {
			o["ClientIP"] = A.Context.Request().Header.Get("IP")
		}
		o["DataCaps"] = GetDataCaps(domain, u.UserID)
		o["Email"] = u.EmailAddress
		o["UserID"] = u.UserID
		o["Name"] = u.Name
		o["SecurityToken"] = common.GetGUID()
		o["Domain"] = domain
		o["Username"] = u.EmailAddress
		//outCrt.Otherdata = make(map[string]string)
		o["UserAgent"] = A.Context.Request().UserAgent()
		bytes, _ := client.Go("ignore", domain, "scope").GetOne().ByUniqueKey(domain).Ok() // fetech user autherized
		//term.Write("AppAutherize For Application "+ApplicationID+" UserID "+UserID, term.Debug)
		o["DataCaps"] = string(bytes[:])
		payload := common.JWTPayload(domain, o["SecurityToken"], o["UserID"], o["Email"], o["Domain"], bytes)
		o["JWT"] = common.Jwt(h.GetSecretKey(domain), payload)
		o["Scope"] = strings.Replace(string(bytes[:]), "\"", "`", -1)

		//outCrt.Otherdata["Tempkey"] = "No"
		//th := TenantHandler{}
		tlist := th.GetTenantsForUser(u.UserID)
		b, _ := json.Marshal(tlist)
		o["TenentsAccessible"] = strings.Replace(string(b[:]), "\"", "`", -1)
		//outCrt = AuthCertificate{u.UserID, u.EmailAddress, u.Name, u.EmailAddress, securityToken, "http://192.168.0.58:9000/instaltionpath", "#0so0936#sdasd", "IPhere"}
		if Config.NumberOFUserLogins != 0 {
			h.LogLoginSessions(username, domain, 1)
		}
		//h.AddSession(outCrt)
		code := r.GenerateRequestCode(o)
		var inputParams map[string]string
		inputParams = make(map[string]string)
		inputParams["@@email@@"] = u.EmailAddress
		inputParams["@@name@@"] = u.Name
		inputParams["@@UserAgent@@"] = o["UserAgent"]
		inputParams["@@ClientIP@@"] = o["ClientIP"]
		inputParams["@@Domain@@"] = domain
		//inputParams["@@SecurityToken@@"] = o["UserAgent"]
		inputParams["@@Code@@"] = code
		//Change activation status to true and save
		//term.Write("Activate User  "+u.Name+" Update User "+u.UserID, term.Debug)
		//go notifier.Send("ignore", "One time password for user login.", "com.duosoftware.auth", "email", "user_otp", inputParams, nil, u.EmailAddress)
		go notifier.Notify("ignore", "user_otp", u.EmailAddress, inputParams, nil)
		A.ResponseBuilder().SetResponseCode(200).WriteAndOveride([]byte(common.MsgJson("One time password sent.")))
		return common.MsgJson("One time password sent.")
	} else {
		h.LogFailedAttemts(username, domain, "")
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Invalid user name password.")))
		//A.Context.Request().
		return common.ErrorJson("Invalid user name password.")
	}
}

func (A Auth) LoginOTPNoPass(username, domain string) string {
	h := newAuthHandler()
	c, msg := h.CanLogin(username, domain)
	if !c {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(msg)))
		//A.Context.Request().
		return common.ErrorJson(msg)
	}
	u, err := h.GetUser(username)
	//h.GetUser
	//if()
	if err == "" {
		//fmt.Println("login succeful")
		//securityToken := common.GetGUID()
		r := requestHandler{}
		th := TenantHandler{}
		//th.Autherized(domain, user)
		x, _ := th.AutherizedUser(domain, u.UserID)
		if !x {
			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(domain + " Is not autherized for signin.")))
			//A.Context.Request().
			return common.ErrorJson(domain + " Is not autherized for signin.")
		}
		o := make(map[string]string)

		if A.Context.Request().Header.Get("PHP") != "101" {
			o["ClientIP"] = A.Context.Request().RemoteAddr
		} else {
			o["ClientIP"] = A.Context.Request().Header.Get("IP")
		}
		o["DataCaps"] = GetDataCaps(domain, u.UserID)
		o["Email"] = u.EmailAddress
		o["UserID"] = u.UserID
		o["Name"] = u.Name
		o["SecurityToken"] = common.GetGUID()
		o["Domain"] = domain
		o["Username"] = u.EmailAddress
		//outCrt.Otherdata = make(map[string]string)
		o["UserAgent"] = A.Context.Request().UserAgent()
		bytes, _ := client.Go("ignore", domain, "scope").GetOne().ByUniqueKey(domain).Ok() // fetech user autherized
		//term.Write("AppAutherize For Application "+ApplicationID+" UserID "+UserID, term.Debug)
		o["DataCaps"] = string(bytes[:])
		payload := common.JWTPayload(domain, o["SecurityToken"], o["UserID"], o["Email"], o["Domain"], bytes)
		o["JWT"] = common.Jwt(h.GetSecretKey(domain), payload)
		o["Scope"] = strings.Replace(string(bytes[:]), "\"", "`", -1)

		//outCrt.Otherdata["Tempkey"] = "No"
		//th := TenantHandler{}
		tlist := th.GetTenantsForUser(u.UserID)
		b, _ := json.Marshal(tlist)
		o["TenentsAccessible"] = strings.Replace(string(b[:]), "\"", "`", -1)
		//outCrt = AuthCertificate{u.UserID, u.EmailAddress, u.Name, u.EmailAddress, securityToken, "http://192.168.0.58:9000/instaltionpath", "#0so0936#sdasd", "IPhere"}
		if Config.NumberOFUserLogins != 0 {
			h.LogLoginSessions(username, domain, 1)
		}
		//h.AddSession(outCrt)
		code := r.GenerateRequestCode(o)
		var inputParams map[string]string
		inputParams = make(map[string]string)
		inputParams["@@email@@"] = u.EmailAddress
		inputParams["@@name@@"] = u.Name
		inputParams["@@UserAgent@@"] = o["UserAgent"]
		inputParams["@@ClientIP@@"] = o["ClientIP"]
		inputParams["@@Domain@@"] = domain
		//inputParams["@@SecurityToken@@"] = o["UserAgent"]
		inputParams["@@Code@@"] = code
		//Change activation status to true and save
		//term.Write("Activate User  "+u.Name+" Update User "+u.UserID, term.Debug)
		//go notifier.Send("ignore", "One time password for user login.", "com.duosoftware.auth", "email", "user_otp", inputParams, nil, u.EmailAddress)
		go notifier.Notify("ignore", "user_otp", u.EmailAddress, inputParams, nil)
		A.ResponseBuilder().SetResponseCode(200).WriteAndOveride([]byte(common.MsgJson("One time password sent.")))
		return common.MsgJson("One time password sent.")
	} else {
		h.LogFailedAttemts(username, domain, "")
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Invalid user name password.")))
		//A.Context.Request().
		return common.ErrorJson("Invalid user name password.")
	}
}

func (A Auth) BlockUser(email string) bool {
	_, error := session.GetSession(A.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		h := newAuthHandler()
		h.LogFailedAttemts(email, "domain", "block")
		return true
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		return false
	}
}

func (A Auth) ReleaseUser(email, b4 string) bool {
	_, error := session.GetSession(A.Context.Request().Header.Get("Securitytoken"), "Nil")
	if error == "" {
		h := newAuthHandler()
		if b4 == "login" {
			h.Release(email)
			return true
		}
		if b4 == "block" {
			//h.LogFailedAttemts(email, "domain", "release")
			h.Release(email)
			return true
		}
		return false
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("SecurityToken  not Autherized")))
		return false
	}
}

func (A Auth) GetUser(Email string) (outCrt User) {
	h := newAuthHandler()
	outCrt, err := h.GetUser(Email)
	if err == "" {
		outCrt.Password = "****"
		outCrt.ConfirmPassword = "******"
		return
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("User Dose not exist.")))
		return
	}
}

func (A Auth) GetSecret(Key string) string {
	h := newAuthHandler()
	return h.GetSecretKey(Key)
}

func (A Auth) GetSession(SecurityToken, Domain string) (a AuthCertificate) {
	h := newAuthHandler()
	/*
		t := new(TenantHandler)
		//var a AuthCertificate
		//h.GetSession(key, Domain)
		if Domain != "nil" {
			user, _ := h.GetSession(SecurityToken, "nil")
			x, _ := t.Autherized(Domain, user)
			if !x {
				A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(Domain + "Not Authorized"))
				return
			}
		}*/
	c, err := h.GetSession(SecurityToken, Domain)
	//fmt.Println(c)
	if err == "" {
		a = c
		return a
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Not Autherized Err:" + err)))
		return
	}

}

func (A Auth) GetSessionStatic(SecurityToken string) (a AuthCertificate) {
	h := newAuthHandler()
	c, err := h.GetSession(SecurityToken, "Nil")
	scope := A.Context.Request().Header.Get("scope")
	if c.Otherdata["OneTimeToken"] == "yes" {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Unautherized Security Token")))
		return
	}
	//fmt.Println(c)
	if err == "" {
		c.SecurityToken = common.GetGUID()
		c.Otherdata["expairyTime"] = ""
		c.Otherdata["OneTimeToken"] = "yes"
		payload := common.JWTPayload(c.Domain, c.SecurityToken, c.UserID, c.Email, c.Domain, []byte(scope))
		c.Otherdata["JWT"] = common.Jwt(h.GetSecretKey(c.Domain), payload)
		c.Otherdata["Scope"] = strings.Replace(scope, "\"", "`", -1)
		c.Otherdata["UserAgent"] = A.Context.Request().UserAgent()
		//c.ClientIP=
		if A.Context.Request().Header.Get("PHP") != "101" {
			c.ClientIP = A.Context.Request().RemoteAddr
		} else {
			c.ClientIP = A.Context.Request().Header.Get("IP")
		}
		h.AddSession(c)
		a = c
		return a
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Not Autherized Err:" + err)))
		return
	}

}

/*
func (A Auth) GetSessionTemporary(SecurityToken string, NumberOftries int) (a AuthCertificate) {

}*/

func (A Auth) Authorize(SecurityToken string, ApplicationID string) (a AuthCertificate) {
	h := newAuthHandler()
	//var a AuthCertificate
	c, err := h.GetSession(SecurityToken, "Nil")
	if err == "" {
		if c.Otherdata["ApplicationID"] == ApplicationID {
			return c
		}
		if h.AppAutherize(ApplicationID, c.UserID, c.Domain) == true {
			//var appH applib.Apphanler
			//application, err := appH.Get(ApplicationID, SecurityToken)
			//if err != "" {
			a = c
			a.ClientIP = A.Context.Request().RemoteAddr

			a.SecurityToken = common.GetGUID()
			//data := make(map[string]interface{})
			id := common.GetHash(ApplicationID + c.UserID)
			bytes, _ := client.Go("ignore", a.Domain, "scope").GetOne().ByUniqueKey(id).Ok() // fetech user autherized
			//term.Write("AppAutherize For Application "+ApplicationID+" UserID "+UserID, term.Debug)
			a.DataCaps = string(bytes[:])
			a.Otherdata["Scope"] = string(bytes[:])
			a.Otherdata["ApplicationID"] = ApplicationID
			a.Otherdata["UserAgent"] = A.Context.Request().UserAgent()
			payload := common.JWTPayload(ApplicationID, a.SecurityToken, a.UserID, a.Email, a.Domain, bytes)
			a.Otherdata["JWT"] = common.Jwt(h.GetSecretKey(ApplicationID), payload)
			h.AddSession(a)
			return a
			//} else {
			//return
			//A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Application ID " + ApplicationID + " not Atherized"))

			//}
		} else {
			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Application ID " + ApplicationID + " not Atherized")))
			return
		}

	}
	A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Session or Application not exist.")))

	return
}

func (A Auth) GetAuthCode(SecurityToken, ApplicationID, URI string) (authCode string) {
	h := newAuthHandler()
	c, err := h.GetSession(SecurityToken, "Nil")
	if err == "" {
		authCode = h.GetAuthCode(ApplicationID, c.UserID, URI)
		return
	}
	A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Application Not exist.")))
	return
}

// ----  FUNCTION BY LASITHA --- DONT DELETE ------------

// func (A Auth) AutherizeApp(SecurityToken, Code, ApplicationID, AppSecret string) bool {
// 	h := newAuthHandler()
// 	c, err := h.GetSession(SecurityToken, "Nil")
// 	if err == "" {
// 		out, err := h.AutherizeApp(Code, ApplicationID, AppSecret, c.UserID)
// 		if err != "" {
// 			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(err))

// 		}
// 		return out
// 	}
// 	A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Application Not exist."))
// 	return false
// }

func (A Auth) GetScope(SecurityToken, Key, Value string) map[string]interface{} {
	h := newAuthHandler()
	_, err := h.GetSession(SecurityToken, "Nil")
	data := make(map[string]interface{})
	if err == "" {
		return data
	}
	A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Session or Application not exist.")))
	return data
}

func (A Auth) UpdateScope(object AuthorizeAppData, SecurityToken, UserID, ApplicationID string) {
	//(, AppSecret string) {
	h := newAuthHandler()
	c, err := h.GetSession(SecurityToken, "Nil")
	if err == "" {

		//Insert Object To Objectore
		id := common.GetHash(ApplicationID + UserID)
		data := make(map[string]interface{})
		data["id"] = id
		data["userid"] = UserID
		data["ApplicationID"] = ApplicationID
		//data["email"] = c.UserID
		for key, value := range object.Object {
			data[key] = value
		}
		client.Go("ignore", c.Domain, "scope").StoreObject().WithKeyField("id").AndStoreOne(data).Ok()
		b, _ := json.Marshal(data)
		A.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
		//insert to Objectstore ends here
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Security Token  Incorrect."))
		//return
	}
}

func (A Auth) AutherizeApp(object AuthorizeAppData, SecurityToken, Code, ApplicationID, AppSecret string) {
	h := newAuthHandler()
	c, err := h.GetSession(SecurityToken, "Nil")
	if err == "" {
		term.Write("AutherizeApp ---------------------------", term.Debug)
		term.Write(object, term.Debug)
		//Insert Object To Objectore
		id := common.GetHash(ApplicationID + c.UserID)
		data := make(map[string]interface{})
		data["id"] = id
		data["userid"] = c.UserID
		data["ApplicationID"] = ApplicationID
		//data["email"] = c.UserID
		for key, value := range object.Object {
			term.Write(value, term.Debug)
			data[key] = value
		}
		term.Write(data, term.Debug)
		client.Go("ignore", c.Domain, "scope").StoreObject().WithKeyField("id").AndStoreOne(data).Ok()
		//insert to Objectstore ends here
		term.Write("AutherizeApp ---------------------------", term.Debug)
		out, err := h.AutherizeApp(Code, ApplicationID, AppSecret, c.UserID, SecurityToken, c.Domain)
		if err != "" {
			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(err)))
			return
		}
		A.ResponseBuilder().SetResponseCode(200).WriteAndOveride([]byte(strconv.FormatBool(out)))
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Application Not exist.")))
	}
}

func (A Auth) GetGUID() string {
	return common.GetGUID()
}

func (A Auth) AddUser(u User) {
	h := newAuthHandler()
	u, err := h.SaveUser(u, false, "xxx")
	if err == "" {
		b, _ := json.Marshal(u)
		A.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson(err)))
	}

}

func (A Auth) InvitedUserRegistration(u User) {
	h := newAuthHandler()
	//t := TenantHandler{}

	u, err := h.SaveUser(u, false, "invitedUserRegistration")

	if err == "" {
		b, _ := json.Marshal(u)
		//x := t.GetTenant(c.Domain)
		//t.AddUsersToTenant(x.TenantID, x.Name, u.UserID, "User")
		A.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(err))
	}

}

func (A Auth) RegisterTenantUser(u User) {
	h := newAuthHandler()
	c, err := h.GetSession(A.Context.Request().Header.Get("Securitytoken"), "Nil")
	//c, err := h.GetSession(SecurityToken, "Nil")
	if err == "" {
		t := TenantHandler{}
		//u.EmailAddress=strings.ToLower(u.EmailAddress
		password := common.RandText(5)
		u.Password = password
		u.ConfirmPassword = password
		u, err := h.SaveUser(u, false, "tenant")

		if err == "" {
			b, _ := json.Marshal(u)
			x := t.GetTenant(c.Domain)
			t.AddUsersToTenant(x.TenantID, x.Name, u.UserID, "User")
			A.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
		} else {
			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(err))
		}
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Security Token Incorrect.")))
		//return
	}

}

func (A Auth) CheckPassword(password string) bool {
	h := newAuthHandler()
	c, err := h.GetSession(A.Context.Request().Header.Get("Securitytoken"), "Nil")
	//c, err := h.GetSession(SecurityToken, "Nil")
	if err == "" {
		//t := TenantHandler{}
		//u.EmailAddress=strings.ToLower(u.EmailAddress

		_, err = h.Login(c.Email, password)
		if err == "" {
			return true
		} else {
			return false
		}
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(common.ErrorJson("Security Token Incorrect.")))
		return false
	}

}
