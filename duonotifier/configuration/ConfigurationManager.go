package configuration

type ConfigurationManager struct {
}

func (c ConfigurationManager) Get() (configuration NotifierConfiguration) {
	//var downloader AbstractConfigDownloader = MockNotifierConfigurationDownloader{}
	var downloader AbstractConfigDownloader = DistributedConfigDownloader{}
	return downloader.DownloadConfiguration()
}
