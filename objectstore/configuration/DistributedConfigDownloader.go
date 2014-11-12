package configuration

type DistributedConfigDownloader struct {
}

func (c DistributedConfigDownloader) DownloadConfiguration() StoreConfiguration {
	return StoreConfiguration{}
}
