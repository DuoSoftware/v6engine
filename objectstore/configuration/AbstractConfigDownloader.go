package configuration

type AbstractConfigDownloader interface {
	DownloadConfiguration() StoreConfiguration
}
