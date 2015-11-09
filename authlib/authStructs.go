package authlib

import (
	"duov6.com/config"
	"duov6.com/term"
	"encoding/json"
)

var Config AuthConfig

//var configRead

func NewUser(userID, EmailAddress, Name, Password string) User {
	return User{userID, EmailAddress, Name, Password, Password, false}
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
	if Config.UserName != "" {
		return
	}
	term.SplashScreen("setup.art")
	if term.Read("Https Enabled (y/n)") == "y" {
		Config.Https_Enabled = true
		Config.Cirtifcate = term.Read("Cirtifcate filename")
		Config.PrivateKey = term.Read("PrivateKey filename")
	} else {
		Config.Https_Enabled = false
	}

	Config.UserName = term.Read("Username")
	Config.Password = term.Read("Password")
	Config.Smtpserver = term.Read("SMTP Server")
	Config.Smtpusername = term.Read("SMTP Username")
	Config.Smtppassword = term.Read("SMTP Password")
	//Config. = term.Read("SMTP Username")

	//Config.
	SetConfig(Config)

}


type AppAutherize struct {
	Name          string
	AppliccatioID string
	AutherizeKey  string
	OtherData     map[string]interface{}
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
	//OtherData       map[string]string
}

type AuthConfig struct {
	Cirtifcate    string
	PrivateKey    string
	Https_Enabled bool
	StoreID       string
	Smtpserver    string
	Smtpusername  string
	Smtppassword  string
	UserName      string
	Password      string
}

type AuthCode struct {
	ApplicationID string
	Code          string
	UserID        string
	URI           string
}
