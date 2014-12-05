package main

import (
	"duov6.com/applib"
	"duov6.com/authlib"
	"duov6.com/config"
	"duov6.com/gorest"
	"duov6.com/stat"
	"duov6.com/statservice"
	"duov6.com/term"
	"encoding/json"
	"net/http"
)

var Config ServiceConfig

type ServiceConfig struct {
	AuthService    bool
	AppService     bool
	Master         bool
	MasterServerIP bool
	//ConfigService bool
}

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

func SetConfig(c ServiceConfig) {

	config.Add(c, "Service")
}

func main() {
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
	term.Write("================================================================", term.Splash)
	term.StartCommandLine()

}

func status() {
	term.Write("Status is running", term.Information)
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

func Bingo() {

	//bingo.RenderNoLayoutToHTML(template, data)
}

func runRestFul() {
	gorest.RegisterService(new(authlib.Auth))
	gorest.RegisterService(new(applib.AppSvc))
	gorest.RegisterService(new(config.ConfigSvc))
	gorest.RegisterService(new(statservice.StatSvc))
	c := authlib.GetConfig()
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

func Setup() {

}
