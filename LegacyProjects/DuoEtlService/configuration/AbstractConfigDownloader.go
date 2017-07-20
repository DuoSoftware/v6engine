package configuration

type AbstractConfigDownloader interface {
	DownloadConfiguration() ETLConfiguration
}
