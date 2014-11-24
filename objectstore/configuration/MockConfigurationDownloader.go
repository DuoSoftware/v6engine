package configuration

type MockConfigurationDownloader struct {
}

func (c MockConfigurationDownloader) DownloadConfiguration() StoreConfiguration {
	config := StoreConfiguration{}

	config.StoreId = "Default"
	config.StorageEngine = "REPLICATED"
	config.ServerConfiguration = make(map[string]map[string]string)
	config.StoreConfiguration = make(map[string]map[string]string)

	var couchmap map[string]string
	couchmap = make(map[string]string)
	couchmap["Url"] = "http://192.168.1.20:8091/"
	couchmap["Bucket"] = "WaterBucket"
	couchmap["UserName"] = ""
	couchmap["Password"] = ""
	config.ServerConfiguration["COUCH"] = couchmap

	var elasticmap map[string]string
	elasticmap = make(map[string]string)
	elasticmap["Host"] = "192.168.2.42"
	elasticmap["Port"] = "9200"
	config.ServerConfiguration["ELASTIC"] = elasticmap

	var getAllMap map[string]string
	getAllMap = make(map[string]string)
	getAllMap["1"] = "ELASTIC"
	//etAllMap["2"] = "COUCH"
	config.StoreConfiguration["GET-ALL"] = getAllMap

	var getSearchMap map[string]string
	getSearchMap = make(map[string]string)
	getSearchMap["1"] = "ELASTIC"
	config.StoreConfiguration["GET-SEARCH"] = getSearchMap

	var getByKey map[string]string
	getByKey = make(map[string]string)
	//getByKey["1"] = "COUCH"
	getByKey["2"] = "ELASTIC"
	config.StoreConfiguration["GET-KEY"] = getByKey

	var getByQuery map[string]string
	getByQuery = make(map[string]string)
	getByQuery["1"] = "ELASTIC"
	config.StoreConfiguration["GET-QUERY"] = getByQuery

	var insertMultipleMap map[string]string
	insertMultipleMap = make(map[string]string)
	//insertMultipleMap["1"] = "COUCH"
	insertMultipleMap["2"] = "ELASTIC"
	config.StoreConfiguration["INSERT-MULTIPLE"] = insertMultipleMap

	var insertSingleMap map[string]string
	insertSingleMap = make(map[string]string)
	//insertSingleMap["1"] = "COUCH"
	insertSingleMap["2"] = "ELASTIC"
	config.StoreConfiguration["INSERT-SINGLE"] = insertSingleMap

	var updateMultipleMap map[string]string
	updateMultipleMap = make(map[string]string)
	//updateMultipleMap["1"] = "COUCH"
	updateMultipleMap["2"] = "ELASTIC"
	config.StoreConfiguration["UPDATE-MULTIPLE"] = updateMultipleMap

	var updateSingleMap map[string]string
	updateSingleMap = make(map[string]string)
	//updateSingleMap["1"] = "COUCH"
	updateSingleMap["2"] = "ELASTIC"
	config.StoreConfiguration["UPDATE-SINGLE"] = updateSingleMap

	var deleteSingleMap map[string]string
	deleteSingleMap = make(map[string]string)
	//deleteSingleMap["1"] = "COUCH"
	deleteSingleMap["2"] = "ELASTIC"
	config.StoreConfiguration["DELETE-SINGLE"] = deleteSingleMap

	var deleteMultipleMap map[string]string
	deleteMultipleMap = make(map[string]string)
	//deleteMultipleMap["1"] = "COUCH"
	deleteMultipleMap["2"] = "ELASTIC"
	config.StoreConfiguration["DELETE-MULTIPLE"] = deleteMultipleMap

	var specialMap map[string]string
	specialMap = make(map[string]string)
	//specialMap["1"] = "COUCH"
	specialMap["2"] = "ELASTIC"
	config.StoreConfiguration["SPECIAL"] = specialMap

	return config
}
