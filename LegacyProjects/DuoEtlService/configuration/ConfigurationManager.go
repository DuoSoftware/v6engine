package configuration

type ConfigurationManager struct {
}

func (c ConfigurationManager) Get() (configuration ETLConfiguration) {
	var downloader AbstractConfigDownloader = DistributedConfigDownloader{}
	return downloader.DownloadConfiguration()
}
