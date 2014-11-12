package authlib

import (
	"duov6.com/config"
)

func NewUser(userID, EmailAddress, Name, Password string) User {
	return User{userID, EmailAddress, Name, Password, Password, false}
}

func GetConfig() AuthConfig {
	return AuthConfig{Cirtifcate: "", PrivateKey: ""}
}

func SetConfig(c AuthConfig) {
	//c.PrivateKey
}

type User struct {
	UserID          string
	EmailAddress    string
	Name            string
	Password        string
	ConfirmPassword string
	Active          bool
}

type AuthConfig struct {
	Cirtifcate    string
	PrivateKey    string
	Https_Enabled bool
	StoreID       string
	smtpserver    string
	smtpusername  string
	smtppassword  string
}
