package main

import (
	"duov6.com/authlib"
	"duov6.com/gorest"
	"duov6.com/term"
	"net/http"

	//"duov6.com/session"
	"fmt"
)

func main() {

	gorest.RegisterService(new(authlib.Auth))
	gorest.RegisterService(new(authlib.TenantSvc))
	err := http.ListenAndServe(":6001", gorest.Handle())
	if err != nil {
		term.Write(err.Error(), term.Error)
		return
	}
	fmt.Scanln()
}
