package configuration

type AbstractConfigDownloader interface {
	DownloadConfiguration() NotifierConfiguration
}
