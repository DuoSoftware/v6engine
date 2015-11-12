package configuration

import "duov6.com/cebadapter"
import (
	"fmt"
	"reflect"
	"strings"
)

type DistributedConfigDownloader struct {
}

func (c DistributedConfigDownloader) DownloadConfiguration(securityToken string, namespace string, class string) StoreConfiguration {

	//Getting Auto Increment Information
	//	incrementConfigs := cebadapter.GetGlobalConfig("AutoIncrementMetaStore")
	//retIncrementConfig := getDefaultConfigurations(getConfigurationIndex("Default", incrementConfigs), incrementConfigs)

	//Getting Objectstore Settings
	configAll := cebadapter.GetGlobalConfig("StoreConfig")
	//Get Default Configurations
	retConfig := getDefaultConfigurations(getConfigurationIndex("Default", configAll), configAll)
	//Check for overriding Configurations
	isOverride, overrideIndex := CheckIfOverridable(configAll, namespace, class)

	//if Overridable ->
	if isOverride {
		for x := 0; x < len(overrideIndex); x++ {
			retConfig = overrideConfigurations(overrideIndex[x], configAll, retConfig)
		}
	}
	//Copy Auto Incremental Settings to Final return Configurations
	//retConfig.AutoIncrementMetaData = retIncrementConfig.AutoIncrementMetaData
	return retConfig

}

func CheckIfOverridable(configAll []interface{}, namespace string, class string) (isOverride bool, overideIndex map[int]int) {

	isOverride = false
	overideIndex = make(map[int]int)

	index := 0

	for x := 0; x < len(configAll); x++ {
		configMap := configAll[x].(map[string]interface{})

		if configMap["StoreId"].(string) == namespace+"."+class {
			fmt.Println("NAMESPACE, CLASS OVERRIDE")
			isOverride = true
			overideIndex[index] = x
			index++
		} else if configMap["StoreId"].(string) == "*."+class {
			fmt.Println("CLASS OVERRIDE")
			isOverride = true
			overideIndex[index] = x
			index++
		} else if configMap["StoreId"].(string) == namespace+".*" {
			fmt.Println("CLASS RANGE OVERRIDE")
			isOverride = true
			overideIndex[index] = x
			index++
		} else if strings.Contains(namespace, strings.Replace(configMap["StoreId"].(string), "*", "", 1)) {
			fmt.Println("NAMESPACE RANGE OVERRIDE")
			isOverride = true
			overideIndex[index] = x
			index++
		}

	}
	return
}

func getConfigurationIndex(keyword string, configAll []interface{}) (index int) {
	index = 0

	for x := 0; x < len(configAll); x++ {
		configMap := configAll[x].(map[string]interface{})
		if configMap["StoreId"].(string) == keyword {
			index = x
			break
		}
	}
	return
}

func getDefaultConfigurations(index int, configAll []interface{}) StoreConfiguration {

	configMap := configAll[index].(map[string]interface{})
	retConfig := StoreConfiguration{}

	if configMap["StorageEngine"] != nil {
		retConfig.StorageEngine = configMap["StorageEngine"].(string)
	}
	if configMap["StoreId"] != nil {
		retConfig.StoreId = configMap["StoreId"].(string)
	}

	if configMap["StoreConfiguration"] != nil {
		retConfig.ServerConfiguration = make(map[string]map[string]string)
		for k, v := range configMap["ServerConfiguration"].(map[string]interface{}) {
			inMap := make(map[string]string)

			for k1, v2 := range v.(map[string]interface{}) {
				inMap[k1] = v2.(string)
			}
			retConfig.ServerConfiguration[k] = inMap
		}
	}

	if configMap["StoreConfiguration"] != nil {
		retConfig.StoreConfiguration = make(map[string]map[string]string)

		for k, v := range configMap["StoreConfiguration"].(map[string]interface{}) {

			inMap := make(map[string]string)

			for k1, v2 := range v.(map[string]interface{}) {
				inMap[k1] = v2.(string)
			}

			retConfig.StoreConfiguration[k] = inMap
		}
	}

	if configMap["AutoIncrementMetaData"] != nil {
		retConfig.AutoIncrementMetaData = make(map[string]map[string]string)

		fmt.Println(reflect.TypeOf(configMap["AutoIncrementMetaData"]))

		for k, v := range configMap["AutoIncrementMetaData"].(map[string]interface{}) {

			inMap := make(map[string]string)

			for k1, v2 := range v.(map[string]interface{}) {
				inMap[k1] = v2.(string)
			}

			retConfig.AutoIncrementMetaData[k] = inMap
		}
	}

	return retConfig
}

func overrideConfigurations(index int, configAll []interface{}, defaultConfig StoreConfiguration) StoreConfiguration {

	configMap := configAll[index].(map[string]interface{})
	retConfig := StoreConfiguration{}

	if configMap["StorageEngine"] != nil {
		retConfig.StorageEngine = configMap["StorageEngine"].(string)
		defaultConfig.StorageEngine = retConfig.StorageEngine
	}
	if configMap["StoreId"] != nil {
		retConfig.StoreId = configMap["StoreId"].(string)
		defaultConfig.StoreId = retConfig.StoreId
	}

	if configMap["ServerConfiguration"] != nil {
		retConfig.ServerConfiguration = make(map[string]map[string]string)

		for k, v := range configMap["ServerConfiguration"].(map[string]interface{}) {
			inMap := make(map[string]string)

			for k1, v2 := range v.(map[string]interface{}) {
				inMap[k1] = v2.(string)
			}
			retConfig.ServerConfiguration[k] = inMap
		}

		for key, value := range retConfig.ServerConfiguration {
			defaultConfig.ServerConfiguration[key] = value
		}
	}

	if configMap["StoreConfiguration"] != nil {
		retConfig.StoreConfiguration = make(map[string]map[string]string)

		for k, v := range configMap["StoreConfiguration"].(map[string]interface{}) {

			inMap := make(map[string]string)

			for k1, v2 := range v.(map[string]interface{}) {
				inMap[k1] = v2.(string)
			}

			retConfig.StoreConfiguration[k] = inMap
		}

		for key, value := range retConfig.StoreConfiguration {
			defaultConfig.StoreConfiguration[key] = value
		}
	}

	if configMap["AutoIncrementMetaData"] != nil {
		retConfig.AutoIncrementMetaData = make(map[string]map[string]string)

		for k, v := range configMap["AutoIncrementMetaData"].(map[string]interface{}) {

			inMap := make(map[string]string)

			for k1, v2 := range v.(map[string]interface{}) {
				inMap[k1] = v2.(string)
			}

			retConfig.AutoIncrementMetaData[k] = inMap
		}

		for key, value := range retConfig.AutoIncrementMetaData {
			defaultConfig.AutoIncrementMetaData[key] = value
		}
	}

	return defaultConfig
}
