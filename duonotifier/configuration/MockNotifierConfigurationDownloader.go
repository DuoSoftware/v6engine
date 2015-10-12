package configuration

type MockNotifierConfigurationDownloader struct {
}

func (c MockNotifierConfigurationDownloader) DownloadConfiguration() NotifierConfiguration {
	config := NotifierConfiguration{}

	config.NotifyId = "com.duosoftware.auth.email"
	config.NotifyMethodsConfig = make(map[string]map[string]string)

	var EMAIL map[string]string
	EMAIL = make(map[string]string)
	EMAIL["Email"] = "prasad@duosoftware.com"
	EMAIL["Password"] = "1Qaz2Wsx3Edc4Rfv"
	EMAIL["Server"] = "smtp.gmail.com:465"
	config.NotifyMethodsConfig["EMAIL"] = EMAIL

	var SMS map[string]string
	SMS = make(map[string]string)
	SMS["Gateway"] = "mobitel"
	SMS["Password"] = "testPass"
	config.NotifyMethodsConfig["SMS"] = SMS

	return config
}
