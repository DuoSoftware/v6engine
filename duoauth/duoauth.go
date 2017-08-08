package main

import (
	"duov6.com/cebadapter"
	"duov6.com/common"
	"duov6.com/duoauth/api"
	"duov6.com/term"
	"encoding/json"
	"fmt"
	"github.com/SiyaDlamini/gorest"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"
)

func main() {
	common.VerifyConfigFiles()
	initializeSettingsFile()
	api.StartTime = time.Now()

	runtime.GOMAXPROCS(runtime.NumCPU())
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

	api.SetupConfig()
	term.GetConfig()

	go runRestFul()

	term.Write("================================================================", term.Splash)
	term.Write("|     https RestFul Service running on :3048                   |", term.Splash)
	term.Write("|     Duo v6 Auth Service ( Azure AD backend for SmoothFlow)   |", term.Splash)
	term.Write("================================================================", term.Splash)

	forever := make(chan bool)
	<-forever

}

func runRestFul() {
	if !common.VerifyGlobalConfig() {
		//GetConfigs from REST...
		if status := cebadapter.GetGlobalConfigFromREST("StoreConfig"); !status {
			fmt.Println("Error retrieving configurations from CEB... Exiting...")
			os.Exit(1)
		}
	}
	gorest.RegisterService(new(api.Auth))
	gorest.RegisterService(new(api.TenantSvc))

	c := api.GetConfig()

	if c.Https_Enabled {
		err := http.ListenAndServeTLS(":3048", c.Certificate, c.PrivateKey, gorest.Handle())
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

func initializeSettingsFile() {
	From := os.Getenv("SMTP_ADDRESS")
	content, err := ioutil.ReadFile("settings.config")
	if err != nil {
		data := make(map[string]interface{})
		if From == "" {
			data["From"] = "DuoWorld.com <mail-noreply@duoworld.com>"
		} else {
			data["From"] = From
		}
		dataBytes, _ := json.Marshal(data)
		_ = ioutil.WriteFile("settings.config", dataBytes, 0666)
	} else {
		vv := make(map[string]interface{})
		_ = json.Unmarshal(content, &vv)
		if From != "" {
			vv["From"] = From
		}
		dataBytes, _ := json.Marshal(vv)
		_ = ioutil.WriteFile("settings.config", dataBytes, 0666)
		fmt.Println(vv)
	}
}
