package authlib

import (
	"duov6.com/config"
)

type AuthHandler struct {
	Config config.File
}

func newAuthHandler() *AuthHandler {
	authhld := new(AuthHandler)
	authhld.Config = config.File{Filename: "auth.cofig"}
	return authhld
}

func (h *AuthHandler) ChangePassword() {

}

func (h *AuthHandler) SaveUser(u User) User {
	return u
}

func SendNotification(u User, Message string) {

}
