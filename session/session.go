package session

import (
	//"duov6.com/applib"
	"duov6.com/common"
	//"github.com/fatih/color"
	//"duov6.com/email"
	"duov6.com/config"
	"duov6.com/objectstore/client"
	"duov6.com/term"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

type AuthCertificate struct {
	UserID, Username, Name, Email, SecurityToken, Domain, DataCaps, ClientIP, MainST string
	Otherdata                                                                        map[string]string
}

type TenantAutherized struct {
	ID            string
	UserID        string
	TenantID      string
	SecurityLevel string
	Autherized    bool
}

func AddSession(a AuthCertificate) {
	fmt.Println("ADDING AN SESSION............................")
	//color.Green("Add Session")
	nowTime := time.Now()
	o := make(map[string]interface{})
	o["ClientIP"] = a.ClientIP
	o["TenantID"] = a.Domain
	o["email"] = a.Username
	o["Name"] = a.Name
	o["LastLoginDate"] = nowTime.UTC().Format("2006-01-02 15:04:05")
	o["CreateDate"] = nowTime.UTC().Format("2006-01-02 15:04:05")

	client.Go(a.SecurityToken, "reports.duosoftware.auth", "lastlogin").StoreObject().WithKeyField("TenantID").AndStoreOne(o).Ok()
	client.Go(a.SecurityToken, "reports.duosoftware.auth", "sessions").StoreObject().WithKeyField("SecurityToken").AndStoreOne(a).Ok()
	fmt.Println(a.SecurityToken)
	client.Go(a.SecurityToken, "s.duosoftware.auth", "sessions").StoreObject().WithKeyField("SecurityToken").AndStoreOne(a).Ok()
	term.Write("AddSession for "+a.Name+" with SecurityToken :"+a.SecurityToken, term.Debug)

	if SessionStateMap == nil {
		SessionStateMap = make(map[string]time.Time)
	}

	SetSessionState(a.SecurityToken, nowTime)

}

func RemoveSession(SecurityToken string) {
	//color.Green("Remove Session")
	RemoveSessionState(SecurityToken)
	//client.Go("ignore", "s.duosoftware.auth", "sessions").DeleteObject().ByUniqueKey(SecurityToken)
	Activ, err := GetSession(SecurityToken, "Nil")
	if err == "" {
		client.Go("ignore", "com.duosoftware.tenant", "authorized").DeleteObject().WithKeyField("SecurityToken").AndDeleteObject(Activ).Ok()
		//client,.Go("ignore", "s.duosoftware.auth", "sessions").StoreObject().WithKeyField("SecurityToken").AndStoreOne(a).Ok()
		term.Write("LogOut for SecurityToken :"+SecurityToken, term.Debug)
	}
	//return true
}

func AutherizedUser(TenantID, UserID string) (bool, TenantAutherized) {
	term.Write("Start Autherized Domain #"+TenantID, term.Debug)
	bytes, err := client.Go("ignore", "com.duosoftware.tenant", "authorized").GetOne().ByUniqueKey(common.GetHash(UserID + "-" + TenantID)).Ok()
	term.Write("SaveUser saving Tenant fetech Error #", term.Debug)
	if err == "" {
		var uList TenantAutherized
		err := json.Unmarshal(bytes, &uList)
		if err == nil {
			term.Write("Autherized #", term.Debug)
			return uList.Autherized, uList
		} else {
			term.Write("Normal Fail to deasseble Not Autherized #"+err.Error(), term.Error)
			//return false, TenantAutherized{}
		}
	} else {
		term.Write("Not Autherized #", term.Debug)
		//return false, TenantAutherized{}
	}
	term.Write("Start Global Autherized Domain #"+TenantID, term.Debug)
	bytes1, err1 := client.Go("ignore", "com.duosoftware.tenant", "authorized").GetOne().ByUniqueKey(TenantID).Ok()
	if err1 == "" {
		var uList TenantAutherized
		err := json.Unmarshal(bytes1, &uList)
		if err == nil {
			term.Write("Autherized #", term.Debug)
			return uList.Autherized, uList
		} else {
			term.Write("Global Fail to deasseble Not Autherized #"+err.Error(), term.Error)
			return false, TenantAutherized{}
		}
	} else {
		term.Write("Not Autherized #", term.Debug)
		return false, TenantAutherized{}
	}
}

func GetRunningSession(UserID string) []AuthCertificate {
	var c []AuthCertificate
	bytes, err := client.Go("ignore", "s.duosoftware.auth", "sessions").GetMany().BySearching("UserID:" + UserID).Ok()
	if err == "" {
		if bytes != nil {
			err := json.Unmarshal(bytes, &c)
			if err != nil {
				term.Write("GetSession Error "+err.Error(), term.Error)
			}
		}
	}
	return c
}

func GetRunningSessionByEmail(Email string) []AuthCertificate {
	var c []AuthCertificate
	bytes, err := client.Go("ignore", "s.duosoftware.auth", "sessions").GetMany().BySearching("Email:" + Email).Ok()
	if err == "" {
		if bytes != nil {
			err := json.Unmarshal(bytes, &c)
			if err != nil {
				term.Write("GetSession Error "+err.Error(), term.Error)
			}
		}
	}
	return c
}

func GetChildSession(Key string) []AuthCertificate {
	var c []AuthCertificate
	bytes, err := client.Go("ignore", "s.duosoftware.auth", "sessions").GetMany().BySearching("MainST:" + Key).Ok()
	if err == "" {
		if bytes != nil {
			err := json.Unmarshal(bytes, &c)
			if err != nil {
				term.Write("GetSession Error "+err.Error(), term.Error)
			}
		}
	}
	return c
}

func GetSession(key, Domain string) (AuthCertificate, string) {
	//color.Green("Get Session")
	bytes, objerr := client.Go(key, "s.duosoftware.auth", "sessions").GetOne().ByUniqueKey(key).Ok()
	term.Write("GetSession For SecurityToken "+key, term.Debug)

	uList := AuthCertificate{}
	errString := ""

	if objerr != "" || bytes == nil {
		term.Write("GetSession Error "+objerr, term.Error)
		errString = "Error Session Not Found"
	} else {
		err := json.Unmarshal(bytes, &uList)
		if err != nil {
			term.Write("GetSession Error "+err.Error(), term.Error)
			errString = "GetSession Error " + err.Error()
		} else {
			if Domain != "Nil" {
				if strings.ToLower(uList.Domain) != strings.ToLower(Domain) {
					x, _ := AutherizedUser(Domain, uList.UserID)
					if x {
						uList.Domain = strings.ToLower(Domain)
						uList.MainST = key
						fmt.Println("FEKKED")
						fmt.Println("|" + strings.ToLower(uList.Domain) + "|")
						fmt.Println("|" + strings.ToLower(Domain) + "|")
						uList.SecurityToken = common.GetGUID()
						uList.Otherdata = make(map[string]string)
						uList.Otherdata["unused"] = "abc"
						term.Write("GetSession For SecurityToken "+key+" new key "+uList.SecurityToken, term.Debug)
					} else {
						uList = AuthCertificate{}
						errString = " Session Cound not be Created "
					}
				}
			}
		}
	}

	if uList.Email != "" && Config.SessionTimeout > 0 {
		//check for validity
		if !ValidateSession(key) {
			LogOut(uList)
			uList = AuthCertificate{}
			errString = "Session Timeout. Please Login again."
		} else {
			SetSessionState(uList.SecurityToken, time.Now())
		}
	}

	return uList, errString
}

func LogOut(a AuthCertificate) {
	//color.Green("Log Out")
	RemoveSessionState(a.SecurityToken)
	client.Go("ignore", "s.duosoftware.auth", "sessions").DeleteObject().WithKeyField("SecurityToken").AndDeleteObject(a).Ok()
	LogoutClildSessions(a.SecurityToken)

	if Config.NumberOFUserLogins != 0 {
		LogLoginSessions(a.Email, a.Domain, -1)
	}
}

func LogoutClildSessions(SecurityToken string) {
	//color.Green("Log out Child Sessions")
	s := GetChildSession(SecurityToken)
	for _, a := range s {
		RemoveSessionState(a.SecurityToken)
		client.Go("ignore", "s.duosoftware.auth", "sessions").DeleteObject().WithKeyField("SecurityToken").AndDeleteObject(a).Ok()
		term.Write("LogOut for "+a.Name+" with SecurityToken :"+a.SecurityToken, term.Debug)
		LogoutClildSessions(a.SecurityToken)
	}
}

type LoginSessions struct {
	Email  string
	Domain string
	Count  int64
}

func LogLoginSessions(email, domain string, item int64) {
	//color.Green("Update Login Sessions")
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
				if x.Count < 0 {
					x.Count = 0
				}
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

//----------------------- SESSION STATES --------------------------

var SessionStateMap map[string]time.Time
var SessionStateMapLock = sync.RWMutex{}

func GetSessionState(index string) (state time.Time) {
	if Config.SessionTimeout == 0 {
		return
	}
	//color.Green("Get Session State")
	SessionStateMapLock.RLock()
	defer SessionStateMapLock.RUnlock()
	state = SessionStateMap[index]
	return
}

func SetSessionState(index string, state time.Time) {
	if Config.SessionTimeout == 0 {
		return
	}
	//color.Green("Set Session State")
	SessionStateMapLock.Lock()
	defer SessionStateMapLock.Unlock()
	SessionStateMap[index] = state
}

func RemoveSessionState(index string) {
	if Config.SessionTimeout == 0 {
		return
	}
	//color.Green("Removing State")
	SessionStateMapLock.RLock()
	defer SessionStateMapLock.RUnlock()
	delete(SessionStateMap, index)
}

func ValidateSession(securityToken string) (status bool) {
	if Config.SessionTimeout == 0 {
		return true
	}
	//color.Green("Valdate Session")
	if SessionStateMap == nil {
		SessionStateMap = make(map[string]time.Time)
	}

	//fmt.Println("------------------------------")
	//fmt.Println(securityToken)
	//fmt.Println(SessionStateMap)
	//fmt.Println(GetSessionState(securityToken))
	//fmt.Println("--------------------------------")

	status = true

	if GetSessionState(securityToken) != (time.Time{}) {
		//securityToken available
		if time.Now().Sub(GetSessionState(securityToken)).Hours() >= float64(Config.SessionTimeout) {
			//time out.. clear the map and delete from session db
			fmt.Println("Validate Session : Time Out")
			status = false
		} else {
			fmt.Println("Validate Session : Valid")
			//All okay
		}
	} else {
		fmt.Println("Validate Session : Not Found.")
		//Auth has been restarted or memory loophole
		status = false
	}

	fmt.Print("Validated :")
	fmt.Println(status)
	return
}

//----------------------- CONFIGS --------------------------

type AuthConfig struct { // Auth Config
	Cirtifcate         string // ssl cirtificate
	PrivateKey         string // Private Key
	Https_Enabled      bool   // Https enabled or not
	StoreID            string // Store ID
	Smtpserver         string // Smptp Server Address
	Smtpusername       string // SMTP Username
	Smtppassword       string // SMTP password
	UserName           string // UserName login to advanced service potal
	Password           string // Password
	NumberOFUserLogins int64
	UserLoginTries     int64
	SessionTimeout     int64
	ExpairyTime        int64
}

var Config AuthConfig

func GetConfig() AuthConfig {
	if Config != (AuthConfig{}) {
		return Config
	}

	b, err := config.Get("Auth")
	if err == nil {
		json.Unmarshal(b, &Config)
	} else {
		Config = AuthConfig{}
	}
	return Config
}

/*
func GetSession(key, Domain string) (AuthCertificate, string) {
	bytes, err := client.Go(key, "s.duosoftware.auth", "sessions").GetOne().ByUniqueKey(key).Ok()
	//bytes, err := client.Go(key, "s.duosoftware.auth", "sessions").GetOne().ByUniqueKeyCache(key, 3600).Ok()
	term.Write("GetSession For SecurityToken "+key, term.Debug)
	//term.Write("GetSession For SecurityToken "+string(bytes), term.Debug)

	var c AuthCertificate
	//AuthCertificate.UserID
	if err == "" {
		if bytes != nil {
			var uList AuthCertificate
			err := json.Unmarshal(bytes, &uList)
			if err == nil {
				if Domain == "Nil" {
					return uList, ""
				} else {

					if strings.ToLower(uList.Domain) != strings.ToLower(Domain) {
						x, _ := AutherizedUser(Domain, uList.UserID)
						if x {
							uList.Domain = strings.ToLower(Domain)
							uList.MainST = key
							uList.SecurityToken = common.GetGUID()
							uList.Otherdata = make(map[string]string)
							uList.Otherdata["unused"] = "abc"
							term.Write("GetSession For SecurityToken "+key+" new key "+uList.SecurityToken, term.Debug)
							//AddSession(uList)
							return uList, ""
						} else {
							return c, Domain + " Session Cound not be Created "
						}
					} else {
						return uList, ""
					}
				}

			} else {
				term.Write("GetSession Error "+err.Error(), term.Error)
			}
		}
	} else {
		term.Write("GetSession Error "+err, term.Error)
	}
	term.Write("GetSession No Session for SecurityToken "+key, term.Debug)

	return c, "Error Session Not Found"
}
*/
