package authlib

import (
	"duov6.com/common"
	"duov6.com/gorest"
	"encoding/json"
)

type AuthCertificate struct {
	UserID, Username, Name, Email, SecurityToken, Domain, DataCaps, ClientIP string
	//Otherdata                                                                map[string]interface{}
}

type Auth struct {
	gorest.RestService
	login        gorest.EndPoint `method:"GET" path:"/Login/{username:string}/{password:string}/{domain:string}" output:"AuthCertificate"`
	autherize    gorest.EndPoint `method:"GET" path:"/Autherize/{SecurityToken:string}/{ApplicationID:string}" output:"AuthCertificate"`
	getAuthCode  gorest.EndPoint `method:"GET" path:"/GetAuthCode/{SecurityToken:string}/{ApplicationID:string}/{URI:string}" output:"string"`
	autherizeApp gorest.EndPoint `method:"GET" path:"/AutherizeApp/{SecurityToken:string}/{Code:string}/{ApplicationID:string}/{AppSecret:string}" output:"bool"`
	addUser      gorest.EndPoint `method:"POST" path:"/AddUser/" postdata:"User"`
	//addApplication gorest.EndPoint `method:"POST" path:"/AddApplication/" postdata:"applib.Application"`
	getUser gorest.EndPoint `method:"GET" path:"/GetUser/" output:"User"`
}

func GetClientIP() string {
	return "hope"
}

func GetDataCaps(Domain, UserID string) string {
	return "#" + Domain + "#" + UserID + "#1#2#4"
}

func (A Auth) LogOut(SecurityToken string) bool {
	h := newAuthHandler()

	c, err := h.GetSession(SecurityToken)
	if err == "" {
		h.LogOut(c)
		return true
	}
	A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Session or Application not exist"))

	return false
}

func (A Auth) Login(username, password, domain string) (outCrt AuthCertificate) {
	h := newAuthHandler()
	u, err := h.Login(username, password)
	if err == "" {
		//fmt.Println("login succeful")
		//securityToken := common.GetGUID()
		outCrt.ClientIP = A.Context.Request().RemoteAddr
		outCrt.DataCaps = GetDataCaps(domain, u.UserID)
		outCrt.Email = u.EmailAddress
		outCrt.UserID = u.UserID
		outCrt.Name = u.Name
		outCrt.SecurityToken = common.GetGUID()
		outCrt.Username = u.EmailAddress

		//outCrt = AuthCertificate{u.UserID, u.EmailAddress, u.Name, u.EmailAddress, securityToken, "http://192.168.0.58:9000/instaltionpath", "#0so0936#sdasd", "IPhere"}
		h.AddSession(outCrt)
		return
	} else {
		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Invalid user name password"))
		return
	}
}

func (A Auth) Autherize(SecurityToken string, ApplicationID string) (a AuthCertificate) {
	h := newAuthHandler()
	//var a AuthCertificate
	c, err := h.GetSession(SecurityToken)
	if err == "" {
		if h.AppAutherize(ApplicationID, c.UserID) == true {
			a = c
			a.ClientIP = A.Context.Request().RemoteAddr
			a.SecurityToken = common.GetGUID()
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
	c, err := h.GetSession(SecurityToken)
	if err == "" {
		authCode = h.GetAuthCode(ApplicationID, c.UserID, URI)
		return
	}
	A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Application Not exist."))
	return
}

func (A Auth) AutherizeApp(SecurityToken, Code, ApplicationID, AppSecret string) bool {
	h := newAuthHandler()
	c, err := h.GetSession(SecurityToken)
	if err == "" {
		out, err := h.AutherizeApp(Code, ApplicationID, AppSecret, c.UserID)
		if err != "" {
			A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(err))

		}
		return out
	}
	A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte("Application Not exist."))
	return false

}

func (A Auth) AddUser(u User) {
	h := newAuthHandler()
	u = h.SaveUser(u)
	b, _ := json.Marshal(u)

	A.ResponseBuilder().SetResponseCode(200).WriteAndOveride(b)

}

func (A Auth) GetUser() User {
	return User{"", "", "", "", "", false}
}
