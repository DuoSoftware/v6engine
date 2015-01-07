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
	//getAllMap["1"] = "ELASTIC"
	//getAllMap["2"] = "COUCH"
	getAllMap["3"] = "REDIS"
	//getAllMap["4"] = "MONGO"
	//getAllMap["5"] = "CASSANDRA"
	//getAllMap["6"] = "HIVE"
	config.StoreConfiguration["GET-ALL"] = getAllMap

	var getSearchMap map[string]string
	getSearchMap = make(map[string]string)
	//getSearchMap["1"] = "ELASTIC"
	//getSearchMap["2"] = "COUCH"
	getSearchMap["3"] = "REDIS"
	//getSearchMap["4"] = "MONGO"
	//getSearchMap["5"] = "CASSANDRA"
	//getSearchMap["6"] = "HIVE"
	config.StoreConfiguration["GET-SEARCH"] = getSearchMap

	var getByKey map[string]string
	getByKey = make(map[string]string)
	//getByKey["1"] = "COUCH"
	//getByKey["2"] = "ELASTIC"
	getByKey["3"] = "REDIS"
	//getByKey["4"] = "MONGO"
	//getByKey["5"] = "CASSANDRA"
	//getByKey["6"] = "HIVE"
	config.StoreConfiguration["GET-KEY"] = getByKey

	var getByQuery map[string]string
	getByQuery = make(map[string]string)
	//getByQuery["1"] = "ELASTIC"
	//getByQuery["2"] = "COUCH"
	getByQuery["3"] = "REDIS"
	//getByQuery["4"] = "MONGO"
	//getByQuery["5"] = "CASSANDRA"
	//getByQuery["6"] = "HIVE"
	config.StoreConfiguration["GET-QUERY"] = getByQuery

	var insertMultipleMap map[string]string
	insertMultipleMap = make(map[string]string)
	//insertMultipleMap["1"] = "COUCH"
	//insertMultipleMap["2"] = "ELASTIC"
	insertMultipleMap["3"] = "REDIS"
	//insertMultipleMap["4"] = "MONGO"
	//insertMultipleMap["5"] = "CASSANDRA"
	//insertMultipleMap["6"] = "HIVE"
	config.StoreConfiguration["INSERT-MULTIPLE"] = insertMultipleMap

	var insertSingleMap map[string]string
	insertSingleMap = make(map[string]string)
	//insertSingleMap["1"] = "COUCH"
	//insertSingleMap["2"] = "ELASTIC"
	insertSingleMap["3"] = "REDIS"
	//insertSingleMap["4"] = "MONGO"
	//insertSingleMap["5"] = "CASSANDRA"
	//insertSingleMap["6"] = "HIVE"
	config.StoreConfiguration["INSERT-SINGLE"] = insertSingleMap

	var updateMultipleMap map[string]string
	updateMultipleMap = make(map[string]string)
	//updateMultipleMap["1"] = "COUCH"
	//updateMultipleMap["2"] = "ELASTIC"
	updateMultipleMap["3"] = "REDIS"
	//updateMultipleMap["4"] = "MONGO"
	//updateMultipleMap["5"] = "CASSANDRA"
	//updateMultipleMap["6"] = "HIVE"
	config.StoreConfiguration["UPDATE-MULTIPLE"] = updateMultipleMap

	var updateSingleMap map[string]string
	updateSingleMap = make(map[string]string)
	//updateSingleMap["1"] = "COUCH"
	//updateSingleMap["2"] = "ELASTIC"
	updateSingleMap["3"] = "REDIS"
	//updateSingleMap["4"] = "MONGO"
	//updateSingleMap["5"] = "CASSANDRA"
	//updateSingleMap["6"] = "HIVE"
	config.StoreConfiguration["UPDATE-SINGLE"] = updateSingleMap

	var deleteSingleMap map[string]string
	deleteSingleMap = make(map[string]string)
	//deleteSingleMap["1"] = "COUCH"
	//deleteSingleMap["2"] = "ELASTIC"
	deleteSingleMap["3"] = "REDIS"
	//deleteSingleMap["4"] = "MONGO"
	//deleteSingleMap["5"] = "CASSANDRA"
	//deleteSingleMap["6"] = "HIVE"
	config.StoreConfiguration["DELETE-SINGLE"] = deleteSingleMap

	var deleteMultipleMap map[string]string
	//deleteMultipleMap = make(map[string]string)
	deleteMultipleMap = make(map[string]string)
	//deleteMultipleMap["1"] = "COUCH"
	//deleteMultipleMap["2"] = "ELASTIC"
	deleteMultipleMap["3"] = "REDIS"
	//deleteMultipleMap["4"] = "MONGO"
	//deleteMultipleMap["5"] = "CASSANDRA"
	//deleteMultipleMap["6"] = "HIVE"
	config.StoreConfiguration["DELETE-MULTIPLE"] = deleteMultipleMap

	var specialMap map[string]string
	specialMap = make(map[string]string)
	//specialMap["1"] = "COUCH"
	//specialMap["2"] = "ELASTIC"
	specialMap["3"] = "REDIS"
	//specialMap["4"] = "MONGO"
	//specialMap["5"] = "CASSANDRA"
	//specialMap["6"] = "HIVE"
	config.StoreConfiguration["SPECIAL"] = specialMap

	//return config
	return config
}
