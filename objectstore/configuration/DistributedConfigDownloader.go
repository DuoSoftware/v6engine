package configuration

import "duov6.com/fws"

type DistributedConfigDownloader struct {
}

func (c DistributedConfigDownloader) DownloadConfiguration() StoreConfiguration {
	configAll := fws.GetGlobalConfig("StoreConfig")
	configMap := configAll[0].(map[string]interface{})
	retConfig := StoreConfiguration{}

	retConfig.ServerConfiguration = make(map[string]map[string]string)

	retConfig.StorageEngine = configMap["StorageEngine"].(string)
	retConfig.StoreId = configMap["StoreId"].(string)

	for k, v := range configMap["ServerConfiguration"].(map[string]interface{}) {
		inMap := make(map[string]string)

		for k1, v2 := range v.(map[string]interface{}) {
			inMap[k1] = v2.(string)
		}
		retConfig.ServerConfiguration[k] = inMap
	}

	retConfig.StoreConfiguration = make(map[string]map[string]string)

	for k, v := range configMap["StoreConfiguration"].(map[string]interface{}) {

		inMap := make(map[string]string)

		for k1, v2 := range v.(map[string]interface{}) {
			inMap[k1] = v2.(string)
		}

		retConfig.StoreConfiguration[k] = inMap
	}

	return retConfig
}
