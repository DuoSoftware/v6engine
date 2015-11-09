package main

import (
	"duov6.com/apisvc"
	"duov6.com/applib"
	"duov6.com/authlib"
	"duov6.com/cebadapter"
	"duov6.com/config"
	"duov6.com/email"
	"duov6.com/gorest"
	"duov6.com/pog"
	"duov6.com/stat"
	"duov6.com/statservice"
	"duov6.com/term"
	"encoding/json"
	"net/http"
)

var Config ServiceConfig



func GetConfig() ServiceConfig {
	b, err := config.Get("Service")
	if err == nil {
		json.Unmarshal(b, &Config)
	} else {
		Config = ServiceConfig{}
		config.Add(Config, "Service")
	}
	return Config

}

func main() {

	cebadapter.Attach("DuoAuth", func(s bool) {
		cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
			term.Write("Store Configuration Successfully Loaded...", term.Information)

			agent := cebadapter.GetAgent()

			agent.Client.OnEvent("globalConfigChanged.StoreConfig", func(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
				cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
					term.Write("Store Configuration Successfully Updated...", term.Information)
				})
			})
		})
		term.Write("Successfully registered in CEB", term.Information)
	})

	authlib.SetupConfig()
	term.GetConfig()

	//go Bingo()
	stat.Start()
	go webServer()
	go runRestFul()

	term.SplashScreen("splash.art")
	term.Write("================================================================", term.Splash)
	term.Write("|     Admintration Console running on  :9000                   |", term.Splash)
	term.Write("|     https RestFul Service running on :3048                   |", term.Splash)
	term.Write("|     Duo v6 Auth Service 6.0                                  |", term.Splash)
	term.Write("|     New updat		                                   |", term.Splash)
	term.Write("================================================================", term.Splash)
	term.StartCommandLine()

}

func webServer() {
	http.Handle(
		"/",
		http.StripPrefix(
			"/",
			http.FileServer(http.Dir("html")),
		),
	)
	http.ListenAndServe(":9000", nil)
}

func runRestFul() {
	gorest.RegisterService(new(authlib.Auth))
	gorest.RegisterService(new(authlib.TenantSvc))
	gorest.RegisterService(new(authlib.UserSVC))
	gorest.RegisterService(new(pog.POGSvc))
	gorest.RegisterService(new(applib.AppSvc))
	gorest.RegisterService(new(config.ConfigSvc))
	gorest.RegisterService(new(statservice.StatSvc))
	gorest.RegisterService(new(apisvc.ApiSvc))
	
	c := authlib.GetConfig()
	email.EmailAddress = c.Smtpusername
	email.Password = c.Smtppassword
	email.SMTPServer = c.Smtpserver

	if c.Https_Enabled {
		err := http.ListenAndServeTLS(":3048", c.Cirtifcate, c.PrivateKey, gorest.Handle())
		if err != nil {
			term.Write(err.Error(), term.Error)
			return
		}
	} else {
		err := http.ListenAndServe(":3048", gorest.Handle())
		if err != nil {
			term.Write(err.Error(), term.Error)
			return
		}
	}

}
