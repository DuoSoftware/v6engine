package configuration

import "duov6.com/cebadapter"

import (
	"reflect"
)

type DistributedConfigDownloader struct {
}

func (c DistributedConfigDownloader) DownloadConfiguration() StoreServiceConfiguration {
	configAll := cebadapter.GetGlobalConfig("ServiceConfig")
	configMap := configAll[0].(map[string]interface{})
	retConfig := StoreServiceConfiguration{}

	retConfig.ServerConfiguration = make(map[string]map[string]string)

	//retConfig.StorageEngine = configMap["StorageEngine"].(string)
	//retConfig.StoreId = configMap["StoreId"].(string)

	for k, v := range configMap["ServerConfiguration"].(map[string]interface{}) {
		inMap := make(map[string]string)

		for k1, v2 := range v.(map[string]interface{}) {
			inMap[k1] = v2.(string)
		}
		retConfig.ServerConfiguration[k] = inMap
	}

	retConfig.PublisherConfiguration = make(map[string]map[string]RoutingKeys)

	for k, v := range configMap["PublisherConfiguration"].(map[string]interface{}) {

		inMap := make(map[string]RoutingKeys)

		for k1, v2 := range v.(map[string]interface{}) {

			temp := ""

			if reflect.TypeOf(v2) == reflect.TypeOf(temp) {
				inMap[k1] = RoutingKeys{}
			} else {

				var routeKeyMap map[string]string
				routeKeyMap = make(map[string]string)

				for kk, vv := range v2.(map[string]interface{}) {
					routeKeyMap[kk] = vv.(string)
				}

				inMap[k1] = RoutingKeys{routeKeyMap}

			}

		}

		retConfig.PublisherConfiguration[k] = inMap
	}

	return retConfig

}
