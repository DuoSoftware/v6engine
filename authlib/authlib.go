package authlib

import (
	"code.google.com/p/gorest"
	"duov6.com/common"
)

type AuthCertificate struct {
	UserID, Username, Name, Email, SecurityToken, Domain, DataCaps, ClientIP string
}

type Auth struct {
	gorest.RestService
	login        gorest.EndPoint `method:"GET" path:"/Login/{username:string}/{password:string}/{domain:string}" output:"AuthCertificate"`
	autherize    gorest.EndPoint `method:"GET" path:"/Autherize/{SecurityToken:string}/{ApplicationID:string}" output:"AuthCertificate"`
	getAuthCode  gorest.EndPoint `method:"GET" path:"/GetAuthCode/{SecurityToken:string}/{ApplicationID:string}" output:"string"`
	autherizeApp gorest.EndPoint `method:"GET" path:"/AutherizeApp/{SecurityToken:string}/{Code:string}/{ApplicationID:string}/{AppSecret:string}" output:"bool"`
	addUser      gorest.EndPoint `method:"POST" path:"/AddUser/" postdata:"User"`
}

func (A Auth) Login(username, password, domain string) AuthCertificate {
	if username == "admin" {
		//fmt.Println("login succeful")
		securityToken := common.GetGUID()
		return AuthCertificate{"0", "Admin", "Administrator", "lasitha.senanayake@gmail.com", securityToken, "http://192.168.0.58:9000/instaltionpath", "#0so0936#sdasd", "IPhere"}

	} else {
		return AuthCertificate{}
	}
}

func (A Auth) Autherize(SecurityToken string, ApplicationID string) AuthCertificate {
	return AuthCertificate{"0", "Admin", "Administrator", "lasitha.senanayake@gmail.com", "SecurityToken", "http://192.168.0.58:9000/instaltionpath", "0so0936", "IPHere"}
}

func (A Auth) GetAuthCode(SecurityToken, ApplicationID string) string {
	return "12233"
}

func (A Auth) AutherizeApp(SecurityToken, Code, ApplicationID, AppSecret string) bool {
	return true
}

func (A Auth) AddUser(u User) {
	h := newAuthHandler()
	h.SaveUser(u)
}
