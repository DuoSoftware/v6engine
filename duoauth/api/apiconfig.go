package api

import (
	"duov6.com/config"
	"duov6.com/term"
	"encoding/json"
	"time"
)

var Config AuthConfig

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
	return //disable for now since Smoothflow uses azure unlimited sessions.

	Config = GetConfig()
	time.Sleep(1 * time.Second)

	if term.Read("Https Enabled (y/n)") == "y" {
		Config.Https_Enabled = true
		Config.Certificate = term.Read("Certificate filename")
		Config.PrivateKey = term.Read("PrivateKey filename")
	} else {
		Config.Https_Enabled = false
	}

	SetConfig(Config)
}

type AuthConfig struct { // Auth Config
	Certificate   string // ssl cirtificate
	PrivateKey    string // Private Key
	Https_Enabled bool   // Https enabled or not
}
