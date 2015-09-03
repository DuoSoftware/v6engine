package configuration

type ConfigurationManager struct {
}

func (c ConfigurationManager) Get() (configuration StoreServiceConfiguration) {
	//var downloader AbstractConfigDownloader = MockServiceConfigurationDownloader{}
	var downloader AbstractConfigDownloader = DistributedConfigDownloader{}
	return downloader.DownloadConfiguration()
}
