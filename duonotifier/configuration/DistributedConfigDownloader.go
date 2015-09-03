package configuration

import "duov6.com/cebadapter"

import ()

type DistributedConfigDownloader struct {
}

func (c DistributedConfigDownloader) DownloadConfiguration() NotifierConfiguration {
	configAll := cebadapter.GetGlobalConfig("DuoNotifier")
	configMap := configAll[0].(map[string]interface{})
	retConfig := NotifierConfiguration{}

	retConfig.NotifyMethodsConfig = make(map[string]map[string]string)

	retConfig.NotifyId = configMap["NotifyId"].(string)

	for k, v := range configMap["NotifyMethodsConfig"].(map[string]interface{}) {
		inMap := make(map[string]string)

		for k1, v2 := range v.(map[string]interface{}) {
			inMap[k1] = v2.(string)
		}
		retConfig.NotifyMethodsConfig[k] = inMap
	}
	return retConfig

}
