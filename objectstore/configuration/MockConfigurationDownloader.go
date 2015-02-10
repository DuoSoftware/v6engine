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
	getAllMap["priority1"] = "ELASTIC"
	//getAllMap["priority2"] = "COUCH"
	//getAllMap["priority3"] = "REDIS"
	//getAllMap["priority4"] = "MONGO"
	//getAllMap["priority5"] = "CASSANDRA"
	//getAllMap["priority6"] = "HIVE"
	//etAllMap["priority7"] = "REDISExcel"
	config.StoreConfiguration["GET-ALL"] = getAllMap

	var getSearchMap map[string]string
	getSearchMap = make(map[string]string)
	getSearchMap["priority1"] = "ELASTIC"
	//getSearchMap["priority2"] = "COUCH"
	//getSearchMap["priority3"] = "REDIS"
	//getSearchMap["priority4"] = "MONGO"
	//getSearchMap["priority5"] = "CASSANDRA"
	//getSearchMap["priority6"] = "HIVE"
	//getSearchMap["priority7"] = "REDISExcel"
	config.StoreConfiguration["GET-SEARCH"] = getSearchMap

	var getByKey map[string]string
	getByKey = make(map[string]string)
	//getByKey["priority1"] = "COUCH"
	getByKey["priority2"] = "ELASTIC"
	//getByKey["priority3"] = "REDIS"
	//getByKey["priority4"] = "MONGO"
	//getByKey["priority5"] = "CASSANDRA"
	//getByKey["priority6"] = "HIVE"
	//getByKey["priority7"] = "REDISExcel"
	config.StoreConfiguration["GET-KEY"] = getByKey

	var getByQuery map[string]string
	getByQuery = make(map[string]string)
	getByQuery["priority1"] = "ELASTIC"
	//getByQuery["priority2"] = "COUCH"
	//getByQuery["priority3"] = "REDIS"
	//getByQuery["priority4"] = "MONGO"
	//getByQuery["priority5"] = "CASSANDRA"
	//getByQuery["priority6"] = "HIVE"
	getByQuery["priority6"] = "REDISExcel"
	config.StoreConfiguration["GET-QUERY"] = getByQuery

	var insertMultipleMap map[string]string
	insertMultipleMap = make(map[string]string)
	//insertMultipleMap["priority1"] = "COUCH"
	insertMultipleMap["priority2"] = "ELASTIC"
	// insertMultipleMap["priority3"] = "REDIS"
	//insertMultipleMap["priority4"] = "MONGO"
	//insertMultipleMap["priority5"] = "CASSANDRA"
	//insertMultipleMap["priority6"] = "HIVE"
	//insertMultipleMap["priority7"] = "REDISExcel"
	config.StoreConfiguration["INSERT-MULTIPLE"] = insertMultipleMap

	var insertSingleMap map[string]string
	insertSingleMap = make(map[string]string)
	//insertSingleMap["priority1"] = "COUCH"
	insertSingleMap["priority2"] = "ELASTIC"
	// insertSingleMap["priority3"] = "REDIS"
	//insertSingleMap["priority4"] = "MONGO"
	//insertSingleMap["priority5"] = "CASSANDRA"
	//insertSingleMap["priority6"] = "HIVE"
	//insertSingleMap["priority7"] = "REDISExcel"
	config.StoreConfiguration["INSERT-SINGLE"] = insertSingleMap

	var updateMultipleMap map[string]string
	updateMultipleMap = make(map[string]string)
	//updateMultipleMap["priority1"] = "COUCH"
	updateMultipleMap["priority2"] = "ELASTIC"
	// updateMultipleMap["priority3"] = "REDIS"
	//updateMultipleMap["priority4"] = "MONGO"
	//updateMultipleMap["priority5"] = "CASSANDRA"
	//updateMultipleMap["priority6"] = "HIVE"
	//insertSingleMap["priority7"] = "REDISExcel"
	config.StoreConfiguration["UPDATE-MULTIPLE"] = updateMultipleMap

	var updateSingleMap map[string]string
	updateSingleMap = make(map[string]string)
	//updateSingleMap["priority1"] = "COUCH"
	updateSingleMap["priority2"] = "ELASTIC"
	// updateSingleMap["priority3"] = "REDIS"
	//updateSingleMap["priority4"] = "MONGO"
	//updateSingleMap["priority5"] = "CASSANDRA"
	//updateSingleMap["priority6"] = "HIVE"
	//updateSingleMap["priority7"] = "REDISExcel"
	config.StoreConfiguration["UPDATE-SINGLE"] = updateSingleMap

	var deleteSingleMap map[string]string
	deleteSingleMap = make(map[string]string)
	//deleteSingleMap["priority1"] = "COUCH"
	deleteSingleMap["priority2"] = "ELASTIC"
	// deleteSingleMap["priority3"] = "REDIS"
	//deleteSingleMap["priority4"] = "MONGO"
	//deleteSingleMap["priority5"] = "CASSANDRA"
	//deleteSingleMap["priority6"] = "HIVE"
	//deleteSingleMap["priority7"] = "REDISExcel"
	config.StoreConfiguration["DELETE-SINGLE"] = deleteSingleMap

	var deleteMultipleMap map[string]string
	//deleteMultipleMap = make(map[string]string)
	deleteMultipleMap = make(map[string]string)
	//deleteMultipleMap["priority1"] = "COUCH"
	deleteMultipleMap["priority2"] = "ELASTIC"
	// deleteMultipleMap["priority3"] = "REDIS"
	//deleteMultipleMap["priority4"] = "MONGO"
	//deleteMultipleMap["priority5"] = "CASSANDRA"
	//deleteMultipleMap["priority6"] = "HIVE"
	//deleteMultipleMap["priority7"] = "REDISExcel"
	config.StoreConfiguration["DELETE-MULTIPLE"] = deleteMultipleMap

	var specialMap map[string]string
	specialMap = make(map[string]string)
	//specialMap["priority1"] = "COUCH"
	specialMap["priority2"] = "ELASTIC"
	// specialMap["priority3"] = "REDIS"
	//specialMap["priority4"] = "MONGO"
	//specialMap["priority5"] = "CASSANDRA"
	//specialMap["priority6"] = "HIVE"
	//specialMap["priority7"] = "REDISExcel"
	config.StoreConfiguration["SPECIAL"] = specialMap

	//return config
	return config
}
