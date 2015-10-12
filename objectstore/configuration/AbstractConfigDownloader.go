package configuration

type AbstractConfigDownloader interface {
	DownloadConfiguration(securityToken string, namespace string, class string) StoreConfiguration
}
