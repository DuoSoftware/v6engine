package authlib

import (
	"duov6.com/common"
	"duov6.com/gorest"
	"duov6.com/objectstore/client"
	"encoding/json"
	"fmt"
	"strconv"
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
	verify      gorest.EndPoint `method:"GET" path:"/" output:"string"`
	login       gorest.EndPoint `method:"GET" path:"/Login/{username:string}/{password:string}/{domain:string}" output:"AuthCertificate"`
	authorize   gorest.EndPoint `method:"GET" path:"/Authorize/{SecurityToken:string}/{ApplicationID:string}" output:"AuthCertificate"`
	getSession  gorest.EndPoint `method:"GET" path:"/GetSession/{SecurityToken:string}/{Domain:string}" output:"AuthCertificate"`
	getAuthCode gorest.EndPoint `method:"GET" path:"/GetAuthCode/{SecurityToken:string}/{ApplicationID:string}/{URI:string}" output:"string"`
	//Lasith's method - Don't Delete
	//autherizeApp       gorest.EndPoint `method:"GET" path:"/AutherizeApp/{SecurityToken:string}/{Code:string}/{ApplicationID:string}/{AppSecret:string}" output:"bool"`
	autherizeApp       gorest.EndPoint `method:"POST" path:"/AutherizeApp/{SecurityToken:string}/{Code:string}/{ApplicationID:string}/{AppSecret:string}" postdata:"AuthorizeAppData"`
	updateScope        gorest.EndPoint `method:"POST" path:"/UpdateScope/{SecurityToken:string}/{UserID:string}/{ApplicationID:string}" postdata:"AuthorizeAppData"`
	addUser            gorest.EndPoint `method:"POST" path:"/UserRegistation/" postdata:"User"`
	registerTenantUser gorest.EndPoint `method:"POST" path:"/RegisterTenantUser/" postdata:"User"`
	userActivation     gorest.EndPoint `method:"GET" path:"/UserActivation/{token:string}" output:"bool"`
	logOut             gorest.EndPoint `method:"GET" path:"/LogOut/{SecurityToken:string}" output:"bool"`
	getUser            gorest.EndPoint `method:"GET" path:"/GetUser/{Email:string}" output:"User"`
	getGUID            gorest.EndPoint `method:"GET" path:"/GetGUID/" output:"string"`
	forgotPassword     gorest.EndPoint `method:"GET" path:"/ForgotPassword/{EmailAddress:string}/{RequestCode:string}" output:"bool"`
	changePassword     gorest.EndPoint `method:"GET" path:"/ChangePassword/{OldPassword:string}/{NewPassword:string}" output:"bool"`
}

func GetClientIP() string {
	return "hope"
}

func GetDataCaps(Domain, UserID string) string {
	return "#" + Domain + "#" + UserID + "#1#2#4"
}

func (A Auth) UserActivation(token string) bool {
	h := newAuthHandler()
	return h.UserActivation(token)
}

func (A Auth) LogOut(SecurityToken string) bool {
	h := newAuthHandler()

	c, err := h.GetSession(SecurityToken, "")
	if err == "" {
		h.LogOut(c)
		return true
	}
	A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Session or Application not exist"))

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
			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Wrong Current Password."))
			return false
		}
		return h.ChangePassword(user, NewPassword)
	} else {
		return false
	}
}

func (A Auth) Verify() (output string) {
	output = "{\"name\": \"DuoAuth\",\"version\": \"1.0.12-a\",\"Change Log\":\"Added Verify!\",\"author\": {\"name\": \"Duo Software\",\"url\": \"http://www.duosoftware.com/\"},\"repository\": {\"type\": \"git\",\"url\": \"https://github.com/DuoSoftware/v6engine/\"}}"
	return
}

func (A Auth) Login(username, password, domain string) (outCrt AuthCertificate) {
	h := newAuthHandler()
	u, err := h.Login(username, password)
	if err == "" {
		//fmt.Println("login succeful")
		//securityToken := common.GetGUID()
		outCrt.ClientIP = A.Context.Request().RemoteAddr
		outCrt.Otherdata["UserAgent"] = A.Context.Request().UserAgent()
		outCrt.DataCaps = GetDataCaps(domain, u.UserID)
		outCrt.Email = u.EmailAddress
		outCrt.UserID = u.UserID
		outCrt.Name = u.Name
		outCrt.SecurityToken = common.GetGUID()
		outCrt.Domain = domain
		outCrt.Username = u.EmailAddress
		outCrt.Otherdata = make(map[string]string)
		//outCrt.Otherdata["Tempkey"] = "No"
		th := TenantHandler{}
		tlist := th.GetTenantsForUser(u.UserID)
		b, _ := json.Marshal(tlist)
		outCrt.Otherdata["TenentsAccessible"] = string(b[:])
		//outCrt = AuthCertificate{u.UserID, u.EmailAddress, u.Name, u.EmailAddress, securityToken, "http://192.168.0.58:9000/instaltionpath", "#0so0936#sdasd", "IPhere"}
		h.AddSession(outCrt)
		return
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Invalid user name password"))
		return
	}
}

func (A Auth) GetUser(Email string) (outCrt User) {
	h := newAuthHandler()
	outCrt, err := h.GetUser(Email)
	if err == "" {
		return
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("User Dose not exist"))
		return
	}
}

func (A Auth) GetSession(SecurityToken, Domain string) (a AuthCertificate) {
	h := newAuthHandler()
	//var a AuthCertificate
	//h.GetSession(key, Domain)
	c, err := h.GetSession(SecurityToken, Domain)
	fmt.Println(c)
	if err == "" {
		a = c
		return a
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Security Token Not Incorrect"))
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
		if h.AppAutherize(ApplicationID, c.UserID) == true {
			a = c
			a.ClientIP = A.Context.Request().RemoteAddr

			a.SecurityToken = common.GetGUID()
			//data := make(map[string]interface{})
			id := common.GetHash(ApplicationID + c.UserID)
			bytes, _ := client.Go("ignore", a.Domain, "scope").GetOne().ByUniqueKey(id).Ok() // fetech user autherized
			//term.Write("AppAutherize For Application "+ApplicationID+" UserID "+UserID, term.Debug)
			a.DataCaps = string(bytes[:])
			a.Otherdata["scope"] = string(bytes[:])
			a.Otherdata["ApplicationID"] = ApplicationID
			a.Otherdata["UserAgent"] = A.Context.Request().UserAgent()
			h.AddSession(a)
			return a
		} else {
			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Application ID " + ApplicationID + " not Atherized"))
			return
		}

	}
	A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Session or Application not exist"))

	return
}

func (A Auth) GetAuthCode(SecurityToken, ApplicationID, URI string) (authCode string) {
	h := newAuthHandler()
	c, err := h.GetSession(SecurityToken, "Nil")
	if err == "" {
		authCode = h.GetAuthCode(ApplicationID, c.UserID, URI)
		return
	}
	A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Application Not exist."))
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

func (A AuthHandler) UpdateScope(object AuthorizeAppData, SecurityToken, UserID, ApplicationID string) {
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

		//insert to Objectstore ends here
	}
}
func (A Auth) AutherizeApp(object AuthorizeAppData, SecurityToken, Code, ApplicationID, AppSecret string) {
	h := newAuthHandler()
	c, err := h.GetSession(SecurityToken, "Nil")
	if err == "" {

		//Insert Object To Objectore
		id := common.GetHash(ApplicationID + c.UserID)
		data := make(map[string]interface{})
		data["id"] = id
		data["userid"] = c.UserID
		data["ApplicationID"] = ApplicationID
		//data["email"] = c.UserID
		for key, value := range object.Object {
			data[key] = value
		}
		client.Go("ignore", c.Domain, "scope").StoreObject().WithKeyField("id").AndStoreOne(data).Ok()
		//insert to Objectstore ends here

		out, err := h.AutherizeApp(Code, ApplicationID, AppSecret, c.UserID)
		if err != "" {
			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(err))
			return
		}
		A.ResponseBuilder().SetResponseCode(200).WriteAndOveride([]byte(strconv.FormatBool(out)))
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Application Not exist."))
	}
}

func (A Auth) GetGUID() string {
	return common.GetGUID()
}

func (A Auth) AddUser(u User) {
	h := newAuthHandler()
	u = h.SaveUser(u, false)
	b, _ := json.Marshal(u)

	A.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)

}

func (A Auth) RegisterTenantUser(u User) {
	h := newAuthHandler()
	c, err := h.GetSession(A.Context.Request().Header.Get("Securitytoken"), "Nil")
	//c, err := h.GetSession(SecurityToken, "Nil")
	if err == "" {
		t := TenantHandler{}
		//u.EmailAddress=strings.ToLower(u.EmailAddress

		u = h.SaveUser(u, false)
		b, _ := json.Marshal(u)
		x := t.GetTenant(c.Domain)
		t.AddUsersToTenant(x.TenantID, x.Name, u.UserID, "User")
		A.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Security Token Not Incorrect"))
		//return
	}

}
