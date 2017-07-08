package authlib

import (
	"duov6.com/config"
	"duov6.com/term"
	"encoding/json"
	"strconv"
	"time"
)

var Config AuthConfig

//var configRead

func NewUser(userID, EmailAddress, Name, Password string) User {
	return User{userID, EmailAddress, Name, Password, Password, false, true}
}

func GetConfig() AuthConfig {
	b, err := config.Get("Auth")
	if err == nil {
		json.Unmarshal(b, &Config)
	} else {
		Config = AuthConfig{}
		config.Add(Config, "Auth")
	}
	return Config
}

func SetConfig(c AuthConfig) {
	config.Add(c, "Auth")
}

func SetupConfig() {

	Config = GetConfig()
	time.Sleep(1 * time.Second)

	if Config.UserLoginTries != 0 {
		return
	}

	/*if Config.UserName != "" {
		return
	}
	term.SplashScreen("setup.art")
	if term.Read("Https Enabled (y/n)") == "y" {
		Config.Https_Enabled = true
		Config.Certificate = term.Read("Certificate filename")
		Config.PrivateKey = term.Read("PrivateKey filename")
	} else {
		Config.Https_Enabled = false
	}

	Config.UserName = term.Read("Username")
	Config.Password = term.Read("Password")
	Config.Smtpserver = term.Read("SMTP Server")
	Config.Smtpusername = term.Read("SMTP Username")
	Config.Smtppassword = term.Read("SMTP Password")
	*/

	s, _ := strconv.ParseInt(term.Read("Number of user login Attempts"), 10, 32)
	x, _ := strconv.ParseInt(term.Read("Number of user login Sessions (0 to Any number of user logins)"), 10, 32)
	v, _ := strconv.ParseInt(term.Read("Session Timeout Period (hours) : "), 10, 32)
	Config.UserLoginTries = s
	Config.NumberOFUserLogins = x
	Config.SessionTimeout = v
	SetConfig(Config)
}

type AppScope struct {
	ScopID        string
	ApplicationID string
	UserID        string
	Scopes        map[string]string
}

type AppAutherize struct {
	Name          string
	AppliccatioID string
	AutherizeKey  string
	OtherData     map[string]interface{}
}

type AuthResponse struct {
	Status    bool
	Message   string
	OtherData map[string]interface{}
}

type AppCertificate struct {
	AuthKey       string
	UserID        string
	ApplicationID string
	AppSecretKey  string
	Otherdata     map[string]interface{}
}

type User struct {
	UserID          string
	EmailAddress    string
	Name            string
	Password        string
	ConfirmPassword string
	Active          bool
	Status          bool
	//UserName        string
	//MobileNo        string
	//OtherData       map[string]string
}

// A AuthConfig represents a authconfig for Configuration For Auth Service.
type AuthConfig struct { // Auth Config
	Certificate        string // ssl cirtificate
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

// A AuthCode represents a authcode cirtificate to Application auth.
type AuthCode struct { // Clas starts here
	ApplicationID string // Application ID
	Code          string // Code for authendication
	UserID        string // User ID of the person who is getting activated
	URI           string // Auth URI
}

type ResetPasswordRequests struct {
	Email             string
	Timestamp         string
	ResetRequestCount int
}

type ResetPasswordToken struct {
	Email string
	Token string
}
