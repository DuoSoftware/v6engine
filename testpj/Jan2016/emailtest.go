package main

import (
	email "duov6.com/duonotifier/client"
	//"fmt"
)

func main() {

	obj := make(map[string]string)
	obj["@@CNAME@@"] = "Kalana"
	obj["@@TITLE@@"] = "Account Creation Confirmation"
	obj["@@MESSAGE@@"] = "The account you created has been verified."
	obj["@@CNAME@@"] = "Kalana"
	obj["@@APPLICATION@@"] = "E-banks.lk"
	obj["@@FOOTER@@"] = "Copyright 2015"
	obj["@@LOGO@@"] = ""
	email.Send("ignore", "kalana", "com.SLT.space.cargills.com", "email", "T_Email_GENERAL", obj, nil, "prasad@duosoftware.com")
}
