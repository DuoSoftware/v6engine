package configuration

type ConfigurationManager struct {
}

func (c ConfigurationManager) Get(securityToken string, namespace string, class string) (configuration StoreConfiguration) {
	var downloader AbstractConfigDownloader = DistributedConfigDownloader{}
	//var downloader AbstractConfigDownloader = MockConfigurationDownloader{}
	return downloader.DownloadConfiguration(securityToken, namespace, class)
}
