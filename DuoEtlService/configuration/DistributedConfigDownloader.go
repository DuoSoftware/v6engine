package configuration

import "duov6.com/cebadapter"

import ()

type DistributedConfigDownloader struct {
}

func (c DistributedConfigDownloader) DownloadConfiguration() ETLConfiguration {
	configAll := cebadapter.GetGlobalConfig("DuoEtl")
	configMap := configAll[0].(map[string]interface{})
	retConfig := ETLConfiguration{}

	retConfig.EtlConfig = make(map[string]map[string]string)

	retConfig.DataPath = configMap["DataPath"].(string)

	for k, v := range configMap["EtlConfig"].(map[string]interface{}) {
		inMap := make(map[string]string)

		for k1, v2 := range v.(map[string]interface{}) {
			inMap[k1] = v2.(string)
		}
		retConfig.EtlConfig[k] = inMap
	}
	return retConfig

}
