package configuration

type AbstractConfigDownloader interface {
	DownloadConfiguration() StoreServiceConfiguration
}
