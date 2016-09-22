package repositories

import (
	"database/sql"
	"duov6.com/common"
	"duov6.com/objectstore/cache"
	"duov6.com/objectstore/keygenerator"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/security"
	"duov6.com/queryparser"
	"encoding/base64"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
	"time"
)

type CloudSqlRepository struct {
}

var availableDbs map[string]interface{}
var availableTables map[string]interface{}
var tableCache map[string]map[string]string
var connection map[string]*sql.DB

func (repository CloudSqlRepository) GetRepositoryName() string {
	return "CloudSQL"
}

func (repository CloudSqlRepository) getConnection(request *messaging.ObjectRequest) (conn *sql.DB, err error) {

	if connection == nil {
		connection = make(map[string]*sql.DB)
	}
	mysqlConf := request.Configuration.ServerConfiguration["MYSQL"]

	username := mysqlConf["Username"]
	password := mysqlConf["Password"]
	url := mysqlConf["Url"]
	port := mysqlConf["Port"]
	IdleLimit := -1
	OpenLimit := 0
	TTL := 5

	poolPattern := url

	if mysqlConf["IdleLimit"] != "" {
		IdleLimit, err = strconv.Atoi(mysqlConf["IdleLimit"])
		if err != nil {
			request.Log("Error : " + err.Error())
		}
	}

	if mysqlConf["OpenLimit"] != "" {
		OpenLimit, err = strconv.Atoi(mysqlConf["OpenLimit"])
		if err != nil {
			request.Log("Error : " + err.Error())
		}
	}

	if mysqlConf["TTL"] != "" {
		TTL, err = strconv.Atoi(mysqlConf["TTL"])
		if err != nil {
			request.Log("Error : " + err.Error())
		}
	}

	if connection[poolPattern] == nil {
		conn, err = repository.CreateConnection(username, password, url, port, IdleLimit, OpenLimit, TTL)
		if err != nil {
			request.Log("Error : " + err.Error())
			return
		}
		connection[poolPattern] = conn
	} else {
		if connection[poolPattern].Ping(); err != nil {
			_ = connection[poolPattern].Close()
			connection[poolPattern] = nil
			conn, err = repository.CreateConnection(username, password, url, port, IdleLimit, OpenLimit, TTL)
			if err != nil {
				request.Log("Error : " + err.Error())
				return
			}
			connection[poolPattern] = conn
		} else {
			conn = connection[poolPattern]
		}
	}
	return conn, err
}

func (repository CloudSqlRepository) CreateConnection(username, password, url, port string, IdleLimit, OpenLimit, TTL int) (conn *sql.DB, err error) {
	conn, err = sql.Open("mysql", username+":"+password+"@tcp("+url+":"+port+")/")
	conn.SetMaxIdleConns(IdleLimit)
	conn.SetMaxOpenConns(OpenLimit)
	conn.SetConnMaxLifetime(time.Duration(TTL) * time.Minute)
	return
}

func (repository CloudSqlRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	isOrderByAsc := false
	isOrderByDesc := false
	orderbyfield := ""
	skip := "0"
	take := "100"

	if request.Extras["skip"] != nil {
		skip = request.Extras["skip"].(string)
	}

	if request.Extras["take"] != nil {
		take = request.Extras["take"].(string)
	}

	if request.Extras["orderby"] != nil {
		orderbyfield = request.Extras["orderby"].(string)
		isOrderByAsc = true
	} else if request.Extras["orderbydsc"] != nil {
		orderbyfield = request.Extras["orderbydsc"].(string)
		isOrderByDesc = true
	}

	query := "SELECT * FROM " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class

	if isOrderByAsc {
		query += " order by " + orderbyfield + " asc "
	} else if isOrderByDesc {
		query += " order by " + orderbyfield + " desc "
	}

	query += " limit " + take
	query += " offset " + skip

	response := repository.queryCommonMany(query, request)
	return response
}

func (repository CloudSqlRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	if request.Body.Query.Parameters != "*" {

		parameters := make(map[string]interface{})

		if request.Extras["skip"] != nil {
			parameters["skip"] = request.Extras["skip"].(string)
		} else {
			parameters["skip"] = ""
		}

		if request.Extras["take"] != nil {
			parameters["take"] = request.Extras["take"].(string)
		} else {
			parameters["take"] = ""
		}

		if request.Extras["orderby"] != nil {
			parameters["orderby"] = request.Extras["orderby"].(string)
		} else if request.Extras["orderbydsc"] != nil {
			parameters["orderbydsc"] = request.Extras["orderbydsc"].(string)
		}

		formattedQuery, err := queryparser.GetCloudSQLQuery(request.Body.Query.Parameters, request.Controls.Namespace, request.Controls.Class, parameters)
		if err != nil {
			request.Log("Error : " + err.Error())
			response.IsSuccess = false
			response.Message = err.Error()
			return response
		}

		query := formattedQuery
		//fmt.Println("Formatted Query : " + query)
		response = repository.queryCommonMany(query, request)
	} else {
		response = repository.GetAll(request)
	}
	return response
}

func (repository CloudSqlRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	if security.ValidateSecurity(getNoSqlKey(request)) {
		response.GetResponseWithBody(getEmptyByteObject())
		request.Log("Error! Security Violation of request detected. Aborting request with error!")
		return response
	}

	query := "SELECT * FROM " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class + " WHERE __os_id = '" + getNoSqlKey(request) + "';"
	response = repository.queryCommonOne(query, request)
	return response
}

func (repository CloudSqlRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	isOrderByAsc := false
	isOrderByDesc := false
	orderbyfield := ""
	skip := "0"
	take := "100"
	isFullTextSearch := false

	if request.Extras["skip"] != nil {
		skip = request.Extras["skip"].(string)
	}

	if request.Extras["take"] != nil {
		take = request.Extras["take"].(string)
	}

	if request.Extras["orderby"] != nil {
		orderbyfield = request.Extras["orderby"].(string)
		isOrderByAsc = true
	} else if request.Extras["orderbydsc"] != nil {
		orderbyfield = request.Extras["orderbydsc"].(string)
		isOrderByDesc = true
	}

	domain := repository.getDatabaseName(request.Controls.Namespace)

	response := RepositoryResponse{}
	query := ""
	if strings.Contains(request.Body.Query.Parameters, ":") {
		tokens := strings.Split(request.Body.Query.Parameters, ":")
		fieldName := tokens[0]
		fieldValue := tokens[1]

		if security.ValidateSecurity(fieldValue) {
			response.GetResponseWithBody(getEmptyByteObject())
			request.Log("Error! Security Violation of request detected. Aborting request with error!")
			return response
		}

		if len(tokens) > 2 {
			fieldValue = ""
			for x := 1; x < len(tokens); x++ {
				fieldValue += tokens[x] + " "
			}
		}

		fieldName = strings.TrimSpace(fieldName)
		fieldValue = strings.TrimSpace(fieldValue)
		if strings.HasPrefix(fieldValue, "*") && strings.HasSuffix(fieldValue, "*") {
			fieldValue = strings.TrimSuffix(fieldValue, "*")
			fieldValue = strings.TrimPrefix(fieldValue, "*")
			query = "select * from " + domain + "." + request.Controls.Class + " where " + fieldName + " LIKE '%" + fieldValue + "%'"
		} else if strings.HasPrefix(fieldValue, "*") {
			fieldValue = strings.TrimPrefix(fieldValue, "*")
			query = "select * from " + domain + "." + request.Controls.Class + " where " + fieldName + " LIKE '%" + fieldValue + "'"
		} else if strings.HasSuffix(fieldValue, "*") {
			fieldValue = strings.TrimSuffix(fieldValue, "*")
			query = "select * from " + domain + "." + request.Controls.Class + " where " + fieldName + " LIKE '" + fieldValue + "%'"
		} else {
			query = "select * from " + domain + "." + request.Controls.Class + " where " + fieldName + "='" + fieldValue + "'"
		}
	} else {
		if request.Body.Query.Parameters == "" || request.Body.Query.Parameters == "*" {
			//Get All Query
			query = "select * from " + domain + "." + request.Controls.Class
		} else {
			//Full Text Search Query
			query = repository.getFullTextSearchQuery(request)
			isFullTextSearch = true
		}
	}

	if !isFullTextSearch {
		if isOrderByAsc {
			query += " order by " + orderbyfield + " asc "
		} else if isOrderByDesc {
			query += " order by " + orderbyfield + " desc "
		}

		query += " limit " + take
		query += " offset " + skip

		query += ";"
	}

	response = repository.queryCommonMany(query, request)
	return response
}

func (repository CloudSqlRepository) getFullTextSearchQuery(request *messaging.ObjectRequest) (query string) {
	var fieldNames []string

	domain := repository.getDatabaseName(request.Controls.Namespace)

	indexedFields := repository.getFullTextIndexes(request)

	if len(indexedFields) > 0 {
		//Indexed Queries
		queryParam := request.Body.Query.Parameters
		queryParam = strings.TrimPrefix(queryParam, "*")
		queryParam = strings.TrimSuffix(queryParam, "*")

		query = "SELECT * FROM " + domain + "." + request.Controls.Class + " WHERE MATCH ("

		argumentCount := 0
		fullTextArguments := ""
		for _, field := range indexedFields {
			if argumentCount < 16 {
				fullTextArguments += field + ","
			} else {
				break
			}
			argumentCount += 1
		}
		fullTextArguments = strings.TrimSuffix(fullTextArguments, ",")
		query += fullTextArguments + ") AGAINST ('" + queryParam
		query += "*' IN BOOLEAN MODE);"
	} else {

		fieldsAndTypes := make(map[string]string)

		tableCacheRedisPattern := "CloudSqlTableCache." + domain + "." + request.Controls.Class

		IsRedis := false
		if CheckRedisAvailability(request) {
			IsRedis = true
		}

		if IsRedis && cache.ExistsKeyValue(request, tableCacheRedisPattern, cache.MetaData) {

			byteVal := cache.GetKeyValue(request, tableCacheRedisPattern, cache.MetaData)
			err := json.Unmarshal(byteVal, &fieldsAndTypes)
			if err != nil {
				request.Log("Error : " + err.Error())
				return
			}

			for name, typee := range fieldsAndTypes {
				if strings.EqualFold(typee, "TEXT") {
					fieldNames = append(fieldNames, name)
				}
			}
		} else if tableCache[domain+"."+request.Controls.Class] != nil {
			//Available in Table Cache
			for name, fieldType := range tableCache[domain+"."+request.Controls.Class] {
				if name != "__osHeaders" && strings.EqualFold(fieldType, "TEXT") {
					fieldNames = append(fieldNames, name)
				}
			}
		} else {
			//Get From Db
			query := "SELECT COLUMN_NAME, DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = '" + domain + "' AND TABLE_NAME = '" + request.Controls.Class + "';"
			repoResponse := repository.queryCommonMany(query, request)
			var mapArray []map[string]interface{}
			err := json.Unmarshal(repoResponse.Body, &mapArray)
			if err != nil {
				request.Log("Error : " + err.Error())
			} else {
				for _, value := range mapArray {
					if value["COLUMN_NAME"].(string) != "__osHeaders" && strings.EqualFold(value["DATA_TYPE"].(string), "TEXT") {
						fieldNames = append(fieldNames, value["COLUMN_NAME"].(string))
					}
				}
			}
		}

		//Non Indexed Queries
		query = "SELECT * FROM " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class + " WHERE Concat("

		//Make Argument Array
		fullTextArguments := ""
		for _, field := range fieldNames {
			fullTextArguments += "IFNULL(" + field + ",''), '',"
		}

		fullTextArguments = fullTextArguments[:(len(fullTextArguments) - 5)]

		queryParam := request.Body.Query.Parameters
		queryParam = strings.TrimPrefix(queryParam, "*")
		queryParam = strings.TrimSuffix(queryParam, "*")
		query += fullTextArguments + ") LIKE '%" + queryParam + "%' "
	}
	return
}

func (repository CloudSqlRepository) getFullTextIndexes(request *messaging.ObjectRequest) (fieldnames []string) {

	conn, err := repository.getConnection(request)
	if err != nil {
		request.Log("Error : " + err.Error())
		return
	}

	domain := repository.getDatabaseName(request.Controls.Class)
	getIndexesQuery := "show index from " + domain + "." + request.Controls.Class + " where Index_type = 'FULLTEXT'"

	indexResult, err := repository.executeQueryMany(request, conn, getIndexesQuery, "")
	if err != nil {
		request.Log("Error : " + err.Error())
	} else {
		for _, obj := range indexResult {
			keyName := obj["Column_name"].(string)
			fieldnames = append(fieldnames, keyName)
		}
	}
	return
}

func (repository CloudSqlRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {

	var response RepositoryResponse

	conn, err := repository.getConnection(request)
	if err != nil {
		response.IsSuccess = false
		response.Message = err.Error()
		return response
	}

	var idData map[string]interface{}
	idData = make(map[string]interface{})

	for index, obj := range request.Body.Objects {
		id := repository.getRecordID(request, obj)
		idData[strconv.Itoa(index)] = id
		request.Body.Objects[index][request.Body.Parameters.KeyProperty] = id
	}

	DataMap := make([]map[string]interface{}, 1)
	var idMap map[string]interface{}
	idMap = make(map[string]interface{})
	idMap["ID"] = idData
	DataMap[0] = idMap

	response = repository.queryStore(request)
	if !response.IsSuccess {
		response = repository.ReRun(request, conn, request.Body.Objects[0])
	}

	response.Data = DataMap
	return response
}

func (repository CloudSqlRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {

	var response RepositoryResponse

	conn, err := repository.getConnection(request)
	if err != nil {
		response.IsSuccess = false
		response.Message = err.Error()
		return response
	}

	id := repository.getRecordID(request, request.Body.Object)
	request.Controls.Id = id
	request.Body.Object[request.Body.Parameters.KeyProperty] = id

	Data := make([]map[string]interface{}, 1)
	var idData map[string]interface{}
	idData = make(map[string]interface{})
	idData["ID"] = id
	Data[0] = idData

	response = repository.queryStore(request)
	if !response.IsSuccess {
		response = repository.ReRun(request, conn, request.Body.Object)
	}

	response.Data = Data
	return response
}

func (repository CloudSqlRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {

	var response RepositoryResponse

	conn, err := repository.getConnection(request)
	if err != nil {
		response.IsSuccess = false
		response.Message = err.Error()
		return response
	}

	var idData map[string]interface{}
	idData = make(map[string]interface{})

	for index, obj := range request.Body.Objects {
		id := repository.getRecordID(request, obj)
		idData[strconv.Itoa(index)] = id
		request.Body.Objects[index][request.Body.Parameters.KeyProperty] = id
	}

	DataMap := make([]map[string]interface{}, 1)
	var idMap map[string]interface{}
	idMap = make(map[string]interface{})
	idMap["ID"] = idData
	DataMap[0] = idMap

	response = repository.queryStore(request)
	if !response.IsSuccess {
		response = repository.ReRun(request, conn, request.Body.Objects[0])
	}

	response.Data = DataMap
	return response
}

func (repository CloudSqlRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {

	var response RepositoryResponse

	conn, err := repository.getConnection(request)
	if err != nil {
		response.IsSuccess = false
		response.Message = err.Error()
		return response
	}

	id := repository.getRecordID(request, request.Body.Object)
	request.Controls.Id = id
	request.Body.Object[request.Body.Parameters.KeyProperty] = id

	Data := make([]map[string]interface{}, 1)
	var idData map[string]interface{}
	idData = make(map[string]interface{})
	idData["ID"] = id
	Data[0] = idData

	response = repository.queryStore(request)
	if !response.IsSuccess {
		response = repository.ReRun(request, conn, request.Body.Object)
	}

	response.Data = Data

	return response
}

func (repository CloudSqlRepository) ReRun(request *messaging.ObjectRequest, conn *sql.DB, obj map[string]interface{}) RepositoryResponse {
	var response RepositoryResponse

	repository.checkSchema(request, conn, request.Controls.Namespace, request.Controls.Class, obj)
	response = repository.queryStore(request)
	if !response.IsSuccess {
		if CheckRedisAvailability(request) {
			cache.FlushCache(request)
		} else {
			tableCache = make(map[string]map[string]string)
			availableDbs = make(map[string]interface{})
			availableTables = make(map[string]interface{})
		}
		repository.checkSchema(request, conn, request.Controls.Namespace, request.Controls.Class, obj)
		response = repository.queryStore(request)
	}

	return response
}

func (repository CloudSqlRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {

	response := RepositoryResponse{}
	conn, err := repository.getConnection(request)
	if err == nil {
		isError := false
		for _, obj := range request.Body.Objects {
			query := repository.getDeleteScript(request.Controls.Namespace, request.Controls.Class, getNoSqlKeyById(request, obj))
			err, message := repository.executeNonQuery(conn, query, request)
			if err != nil {
				isError = true
			} else {
				if message == "No Rows Changed" {
					request.Log("Information : No Rows Changed for : " + request.Body.Parameters.KeyProperty + " = " + obj[request.Body.Parameters.KeyProperty].(string))
				}
			}
		}
		if isError {
			response.IsSuccess = false
			response.Message = "Error deleting all objects. Please double check data!"
		} else {
			response.IsSuccess = true
			response.Message = "Successfully Deleted all objects from CloudSQL repository!"
		}
	} else {
		response.IsSuccess = false
		response.Message = "Error deleting all objects! : " + err.Error()
	}
	repository.closeConnection(conn)
	return response
}

func (repository CloudSqlRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {

	response := RepositoryResponse{}
	conn, err := repository.getConnection(request)
	if err == nil {
		query := repository.getDeleteScript(request.Controls.Namespace, request.Controls.Class, getNoSqlKey(request))
		err, message := repository.executeNonQuery(conn, query, request)
		if err != nil {
			response.IsSuccess = false
			response.Message = "Failed Deleting from CloudSQL repository : " + err.Error()
		} else {
			response.IsSuccess = true
			response.Message = "Successfully Deleted from CloudSQL repository!"
			if message == "No Rows Changed" {
				request.Log("Information : No Rows Changed for : " + request.Body.Parameters.KeyProperty + " = " + request.Body.Object[request.Body.Parameters.KeyProperty].(string))
			}
		}
	} else {
		response.IsSuccess = false
		response.Message = "Failed Deleting from CloudSQL repository : " + err.Error()
	}
	repository.closeConnection(conn)
	return response
}

func (repository CloudSqlRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {

	response := RepositoryResponse{}
	queryType := request.Body.Special.Type
	queryType = strings.ToLower(queryType)
	domain := repository.getDatabaseName(request.Controls.Namespace)

	switch queryType {
	case "getfields":
		request.Log("Debug : Starting GET-FIELDS sub routine!")
		query := "EXPLAIN " + domain + "." + request.Controls.Class + ";"
		var resultSet []map[string]interface{}
		repoResponse := repository.queryCommonMany(query, request)
		err := json.Unmarshal(repoResponse.Body, &resultSet)
		if err != nil {
			response.IsSuccess = false
			response.Message = err.Error()
		} else {
			for x := 0; x < len(resultSet); x++ {
				delete(resultSet[x], "Default")
				delete(resultSet[x], "Extra")
				delete(resultSet[x], "Key")
				delete(resultSet[x], "Null")
			}
			response.IsSuccess = true
			byteArray, _ := json.Marshal(resultSet)
			response.Body = byteArray
		}
		return response
	case "getclasses":
		request.Log("Error : Starting GET-CLASSES sub routine")
		query := "SELECT DISTINCT TABLE_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='" + domain + "';"
		repoResponse := repository.queryCommonMany(query, request)
		var mapArray []map[string]interface{}
		err := json.Unmarshal(repoResponse.Body, &mapArray)
		if err != nil {
			request.Log("Error : " + err.Error())
			repoResponse.Body = nil
			return repoResponse
		} else {
			valueArray := make([]string, len(mapArray))
			for index, value := range mapArray {
				valueArray[index] = value["TABLE_NAME"].(string)
			}
			repoResponse.Body, _ = json.Marshal(valueArray)
			return repoResponse
		}
	case "getnamespaces":
		request.Log("Debug : Starting GET-NAMESPACES sub routine")
		query := "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME != 'information_schema' AND SCHEMA_NAME !='mysql' AND SCHEMA_NAME !='performance_schema';"
		result := repository.queryCommonMany(query, request)
		var resultSet []map[string]interface{}
		if err := json.Unmarshal(result.Body, &resultSet); err != nil {
			response.IsSuccess = false
			response.Message = err.Error()
		} else {
			response.IsSuccess = true
			var resultArray []string
			for _, singleDB := range resultSet {
				resultArray = append(resultArray, singleDB["SCHEMA_NAME"].(string))
			}
			byteArray, _ := json.Marshal(resultArray)
			response.Body = byteArray
		}
		return response
	case "gettablemeta":
		recordCount := 0
		fieldNameList := make([]string, 0)
		isError := false
		request.Log("Debug : Starting GET-Table-Meta-Data sub routine")

		query := "SELECT count(*) as count FROM " + domain + "." + request.Controls.Class + ";"
		result := repository.queryCommonMany(query, request)
		//fmt.Println(string(result.Body))
		var resultSet []map[string]interface{}
		if err := json.Unmarshal(result.Body, &resultSet); err != nil {
			isError = true
		} else {
			if len(resultSet) > 0 {
				recordCount, _ = strconv.Atoi(resultSet[0]["count"].(string))
			}
		}

		//........... Get PK and Table Info .................

		query = "EXPLAIN " + domain + "." + request.Controls.Class + ";"
		var resultSet2 []map[string]interface{}
		repoResponse := repository.queryCommonMany(query, request)
		err := json.Unmarshal(repoResponse.Body, &resultSet2)
		if err != nil {
			isError = true
		} else {
			if len(resultSet2) > 0 {
				for x := 0; x < len(resultSet2); x++ {
					fieldNameList = append(fieldNameList, resultSet2[x]["Field"].(string))
				}
			}
		}

		if isError {
			response.IsSuccess = false
		} else {
			response.IsSuccess = true
			returnMap := make(map[string]interface{})
			returnMap["RecordCount"] = recordCount
			returnMap["FieldList"] = fieldNameList
			//fmt.Println(returnMap)
			byteArray, _ := json.Marshal(returnMap)
			response.Body = byteArray
		}

		return response
	case "getselected":
		fieldNames := strings.Split(strings.TrimSpace(request.Body.Special.Parameters), " ")
		query := "select " + fieldNames[0]
		for x := 1; x < len(fieldNames); x++ {
			query += "," + fieldNames[x]
		}
		query += " from " + domain + "." + request.Controls.Class
		return repository.queryCommonMany(query, request)
	case "dropclass":
		request.Log("Debug : Starting Delete-Class sub routine")
		conn, err := repository.getConnection(request)
		if err == nil {
			query := "DROP TABLE " + domain + "." + request.Controls.Class
			err, _ := repository.executeNonQuery(conn, query, request)
			if err != nil {
				response.IsSuccess = false
				response.Message = "Error Dropping Table in CloudSQL Repository : " + err.Error()
			} else {
				//Delete Class from availableTables and tablecache
				if CheckRedisAvailability(request) {
					_ = cache.DeleteKey(request, ("CloudSqlTableCache." + domain + "." + request.Controls.Class), cache.MetaData)
					_ = cache.DeleteKey(request, ("CloudSqlAvailableTables." + domain + "." + request.Controls.Class), cache.MetaData)
					_ = cache.DeletePattern(request, (domain + "." + request.Controls.Class + "*"), cache.Data)
				} else {
					delete(availableTables, (domain + "." + request.Controls.Class))
					delete(tableCache, (domain + "." + request.Controls.Class))
				}
				response.IsSuccess = true
				response.Message = "Successfully Dropped Table : " + request.Controls.Class
			}
		} else {
			response.IsSuccess = false
			response.Message = "Connection Failed to CloudSQL Server"
		}
		repository.closeConnection(conn)
	case "dropnamespace":
		request.Log("Debug : Starting Delete-Database sub routine")
		conn, err := repository.getConnection(request)
		if err == nil {
			query := "DROP SCHEMA " + domain
			err, _ := repository.executeNonQuery(conn, query, request)
			if err != nil {
				response.IsSuccess = false
				response.Message = "Error Dropping Table in CloudSQL Repository : " + err.Error()
			} else {
				if CheckRedisAvailability(request) {

					_ = cache.DeleteKey(request, ("CloudSqlTableCache." + domain + "." + request.Controls.Class), cache.MetaData)

					var availableTablesKeys []string
					availableTablesPattern := "CloudSqlAvailableTables." + domain + ".*"
					availableTablesKeys = cache.GetKeyListPattern(request, availableTablesPattern, cache.MetaData)
					if len(availableTablesKeys) > 0 {
						for _, name := range availableTablesKeys {
							_ = cache.DeleteKey(request, name, cache.MetaData)
						}
					}
					_ = cache.DeleteKey(request, ("CloudSqlAvailableDbs." + domain), cache.MetaData)
					_ = cache.DeletePattern(request, (domain + "*"), cache.Data)

				} else {
					//Delete Namespace from availableDbs
					delete(availableDbs, domain)
					//Delete all associated Classes from it's TableCache and availableTables
					for key, _ := range availableTables {
						if strings.Contains(key, domain) {
							delete(availableTables, key)
							delete(tableCache, key)
						}
					}
				}
				response.IsSuccess = true
				response.Message = "Successfully Dropped Table : " + request.Controls.Class
			}
		} else {
			response.IsSuccess = false
			response.Message = "Connection Failed to CloudSQL Server"
		}
		repository.closeConnection(conn)
	case "flushcache":
		if CheckRedisAvailability(request) {
			cache.FlushCache(request)
		} else {
			tableCache = make(map[string]map[string]string)
			availableDbs = make(map[string]interface{})
			availableTables = make(map[string]interface{})
		}

		response.IsSuccess = true
		response.Message = "Cache Cleared successfully!"
	case "idservice":
		var IsPattern bool
		var idServiceCommand string

		if request.Body.Special.Extras["Pattern"] != nil {
			IsPattern = request.Body.Special.Extras["Pattern"].(bool)
		}

		if request.Body.Special.Extras["Command"] != nil {
			idServiceCommand = strings.ToLower(request.Body.Special.Extras["Command"].(string))
		}

		conn, err := repository.getConnection(request)
		if err != nil {
			request.Log("Error : " + err.Error())
			response.IsSuccess = false
			response.Message = "Connection Error! : " + err.Error()
			return response
		}

		err = repository.checkAvailabilityDb(request, conn, domain)
		if err != nil {
			request.Log("Error : " + err.Error())
			response.IsSuccess = false
			response.Message = "Database Error! : " + err.Error()
			return response
		}

		switch idServiceCommand {
		case "getid":
			if IsPattern {
				//pattern code goes here
				prefix, valueInString := keygenerator.GetPatternAttributes(request)
				var value int
				value, _ = strconv.Atoi(valueInString)

				if CheckRedisAvailability(request) {
					id := keygenerator.GetIncrementID(request, "CLOUDSQL", value)

					for x := 0; x < len(request.Controls.Class); x++ {
						if (len(prefix) + len(id)) < len(request.Controls.Class) {
							prefix += "0"
						} else {
							break
						}
					}

					id = prefix + id
					response.Body = []byte(id)
					response.IsSuccess = true
					response.Message = "Successfully Completed!"
				} else {
					response.IsSuccess = false
					response.Message = "REDIS not Available!"
				}

			} else {
				//Get ID and Return
				if CheckRedisAvailability(request) {
					id := keygenerator.GetIncrementID(request, "CLOUDSQL", 0)
					response.Body = []byte(id)
					response.IsSuccess = true
					response.Message = "Successfully Completed!"
				} else {
					response.IsSuccess = false
					response.Message = "REDIS not Available!"
				}
			}
		case "readid":
			if IsPattern {
				//pattern code goes here
				prefix, valueInString := keygenerator.GetPatternAttributes(request)
				var value int
				value, _ = strconv.Atoi(valueInString)

				if CheckRedisAvailability(request) {
					id := keygenerator.GetTentativeID(request, "CLOUDSQL", value)

					for x := 0; x < len(request.Controls.Class); x++ {
						if (len(prefix) + len(id)) < len(request.Controls.Class) {
							prefix += "0"
						} else {
							break
						}
					}

					intVal, _ := strconv.Atoi(id)
					id = strconv.Itoa(intVal + 1)
					id = prefix + id
					response.Body = []byte(id)
					response.IsSuccess = true
					response.Message = "Successfully Completed!"
				} else {
					response.IsSuccess = false
					response.Message = "REDIS not Available!"
				}

			} else {
				//Get ID and Return
				if CheckRedisAvailability(request) {
					id := keygenerator.GetTentativeID(request, "CLOUDSQL", 0)
					intVal, _ := strconv.Atoi(id)
					id = strconv.Itoa(intVal + 1)
					response.Body = []byte(id)
					response.IsSuccess = true
					response.Message = "Successfully Completed!"
				} else {
					response.IsSuccess = false
					response.Message = "REDIS not Available!"
				}
			}
		default:
			response.IsSuccess = false
			response.Message = "No Such Command is facilitated!"
		}
	case "fulltextsearch":
		var FTH_command string

		if request.Body.Special.Extras["Command"] != nil {
			FTH_command = strings.ToLower(request.Body.Special.Extras["Command"].(string))
		}

		conn, err := repository.getConnection(request)
		if err != nil {
			response.IsSuccess = false
			response.Message = err.Error()
			return response
		}

		switch FTH_command {
		case "reset":
			getIndexesQuery := "show index from " + domain + "." + request.Controls.Class + " where Index_type = 'FULLTEXT'"

			indexResult, err := repository.executeQueryMany(request, conn, getIndexesQuery, "")
			if err != nil {
				request.Log("Error : " + err.Error())
			} else {
				executedList := ""
				for _, obj := range indexResult {
					keyName := obj["Key_name"].(string)
					if !strings.Contains(executedList, keyName) {
						alterQuery := "ALTER TABLE " + domain + "." + request.Controls.Class + " DROP INDEX " + keyName
						_, _ = repository.executeNonQuery(conn, alterQuery, request)
						executedList += " " + keyName
					}
				}
			}
			response.IsSuccess = true
			response.Message = "Successfully dropped Full Text Indexes!"
		case "index":
			fieldNames := strings.Split(strings.TrimSpace(request.Body.Special.Parameters), " ")
			alterQuery := "ALTER TABLE " + domain + "." + request.Controls.Class + " ADD FULLTEXT(" + fieldNames[0]
			for x := 1; x < len(fieldNames); x++ {
				alterQuery += ", " + fieldNames[x]
			}
			alterQuery += ");"
			err, _ = repository.executeNonQuery(conn, alterQuery, request)
			if err != nil {
				response.IsSuccess = false
				response.Message = err.Error()
			} else {
				response.IsSuccess = true
				response.Message = "Successfully added Full Text Indexes!"
			}
		default:
			response.IsSuccess = false
			response.Message = "No Such Command is facilitated!"
		}
	case "uniqueindex":
		var UIC_command string

		if request.Body.Special.Extras["Command"] != nil {
			UIC_command = strings.ToLower(request.Body.Special.Extras["Command"].(string))
		}

		conn, err := repository.getConnection(request)
		if err != nil {
			response.IsSuccess = false
			response.Message = err.Error()
			return response
		}

		switch UIC_command {
		case "reset":
			getIndexesQuery := "show index from " + domain + "." + request.Controls.Class + " where Index_type = 'BTREE' AND Key_name != 'PRIMARY';"

			indexResult, err := repository.executeQueryMany(request, conn, getIndexesQuery, "")
			if err != nil {
				request.Log("Error : " + err.Error())
			} else {
				executedList := ""
				for _, obj := range indexResult {
					keyName := obj["Key_name"].(string)
					if !strings.Contains(executedList, keyName) {
						alterQuery := "ALTER TABLE " + domain + "." + request.Controls.Class + " DROP INDEX " + keyName
						_, _ = repository.executeNonQuery(conn, alterQuery, request)
						executedList += " " + keyName
					}
				}
			}
			response.IsSuccess = true
			response.Message = "Successfully dropped UNIQUE Key Indexes!"
		case "index":
			indexNames := strings.Split(strings.TrimSpace(request.Body.Special.Parameters), " ")
			isAllDone := true

			for _, singleName := range indexNames {
				indexID := common.GetGUID()
				alterQuery := "CREATE UNIQUE INDEX " + indexID + " ON " + domain + "." + request.Controls.Class + " (" + singleName + ");"
				err, _ = repository.executeNonQuery(conn, alterQuery, request)
				if err != nil {
					//1170 - Non defined key length for indexable field
					if strings.Contains(err.Error(), "BLOB/TEXT") {
						modifyQuery := "ALTER TABLE " + domain + "." + request.Controls.Class + " MODIFY COLUMN " + singleName + " varchar(255);"
						err, _ = repository.executeNonQuery(conn, modifyQuery, request)
						if err != nil {
							request.Log("Error : " + err.Error())
						} else {
							err, _ = repository.executeNonQuery(conn, alterQuery, request)
							if err != nil {
								request.Log("Error : " + err.Error())
								isAllDone = false
							}
						}
					} else {
						request.Log("Error : " + err.Error())
					}
				}
			}

			if isAllDone {
				response.IsSuccess = true
				response.Message = "Successfully added UNIQUE Indexes!"
			} else {
				response.IsSuccess = false
				response.Message = "Creating UNIQUE Indexes Failed!"
			}
		default:
			response.IsSuccess = false
			response.Message = "No Such Command is facilitated!"
		}

	default:
		response.IsSuccess = false
		response.Message = "No such Special Type is Implemented!"

	}

	return response
}

func (repository CloudSqlRepository) Test(request *messaging.ObjectRequest) {
}

func (repository CloudSqlRepository) ClearCache(request *messaging.ObjectRequest) {
	if CheckRedisAvailability(request) {
		cache.FlushCache(request)
	} else {
		tableCache = make(map[string]map[string]string)
		availableDbs = make(map[string]interface{})
		availableTables = make(map[string]interface{})
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////SQL GENERATORS/////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////
func (repository CloudSqlRepository) queryCommon(query string, request *messaging.ObjectRequest, isOne bool) RepositoryResponse {
	response := RepositoryResponse{}

	request.Log("Info Query : " + query)

	conn, err := repository.getConnection(request)
	if err == nil {
		var err error
		dbName := repository.getDatabaseName(request.Controls.Namespace)
		err = repository.buildTableCache(request, conn, dbName, request.Controls.Class)
		if err != nil {
			request.Log("Error : " + err.Error())
		}

		var obj interface{}
		tableName := dbName + "." + request.Controls.Class
		if isOne {
			obj, err = repository.executeQueryOne(request, conn, query, tableName)
		} else {
			obj, err = repository.executeQueryMany(request, conn, query, tableName)
		}

		if err == nil {
			var bytes []byte
			if isOne {
				bytes, _ = json.Marshal(obj.(map[string]interface{}))
			} else {
				bytes, _ = json.Marshal(obj.([]map[string]interface{}))
			}

			if checkEmptyByteArray(bytes) {
				response.GetResponseWithBody(getEmptyByteObject())
			} else {
				response.GetResponseWithBody(bytes)
			}
		} else {
			response.GetResponseWithBody(getEmptyByteObject())
		}
	} else {
		response.GetErrorResponse("Error connecting to CloudSQL : " + err.Error())
	}
	repository.closeConnection(conn)
	return response
}

func (repository CloudSqlRepository) queryCommonMany(query string, request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryCommon(query, request, false)

}

func (repository CloudSqlRepository) queryCommonOne(query string, request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryCommon(query, request, true)
}

func (repository CloudSqlRepository) queryStore(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	conn, _ := repository.getConnection(request)

	domain := request.Controls.Namespace
	class := request.Controls.Class

	isOkay := true

	if request.Body.Object != nil || len(request.Body.Objects) == 1 {

		obj := make(map[string]interface{})

		if request.Body.Object != nil {
			obj = request.Body.Object
		} else {
			obj = request.Body.Objects[0]
		}

		insertScript := repository.getSingleObjectInsertQuery(request, domain, class, obj, conn)
		err, _ := repository.executeNonQuery(conn, insertScript, request)
		if err != nil {
			if !strings.Contains(err.Error(), "specified twice") {
				updateScript := repository.getSingleObjectUpdateQuery(request, domain, class, obj, conn)
				err, message := repository.executeNonQuery(conn, updateScript, request)
				if err != nil {
					isOkay = false
					request.Log("Error : " + err.Error())
				} else {
					if message == "No Rows Changed" {
						request.Log("Information : No Rows Changed for : " + request.Body.Parameters.KeyProperty + " = " + obj[request.Body.Parameters.KeyProperty].(string))
					}
					isOkay = true
				}
			} else {
				isOkay = false
			}
		} else {
			isOkay = true
		}

	} else {

		//execute insert queries
		scripts, err := repository.GetMultipleStoreScripts(conn, request)

		for x := 0; x < len(scripts); x++ {
			script := scripts[x]["query"].(string)
			if err == nil && script != "" {

				err, _ := repository.executeNonQuery(conn, script, request)
				if err != nil {
					request.Log("Error : " + err.Error())
					if strings.Contains(err.Error(), "Duplicate entry") {
						errorBlock := scripts[x]["queryObject"].([]map[string]interface{})
						for _, singleQueryObject := range errorBlock {
							insertScript := repository.getSingleObjectInsertQuery(request, domain, class, singleQueryObject, conn)
							err1, _ := repository.executeNonQuery(conn, insertScript, request)
							if err1 != nil {
								if !strings.Contains(err.Error(), "specified twice") {
									updateScript := repository.getSingleObjectUpdateQuery(request, domain, class, singleQueryObject, conn)
									err2, message := repository.executeNonQuery(conn, updateScript, request)
									if err2 != nil {
										request.Log("Error : " + err2.Error())
										isOkay = false
									} else {
										if message == "No Rows Changed" {
											request.Log("Information : No Rows Changed for : " + request.Body.Parameters.KeyProperty + " = " + singleQueryObject[request.Body.Parameters.KeyProperty].(string))
										}
									}
								}
							}
						}
					} else {
						//if strings.Contains(err.Error(), "doesn't exist") {
						isOkay = false
						break
						//}
					}
				}

			} else {
				isOkay = false
				request.Log("Error : " + err.Error())
			}
		}

	}

	if isOkay {
		response.IsSuccess = true
		response.Message = "Successfully stored object(s) in CloudSQL"
		request.Log("Debug : " + response.Message)
	} else {
		response.IsSuccess = false
		response.Message = "Error storing/updating all object(s) in CloudSQL."
		request.Log("Error : " + response.Message)
	}

	repository.closeConnection(conn)
	return response
}

func (repository CloudSqlRepository) getByKey(conn *sql.DB, namespace string, class string, id string, request *messaging.ObjectRequest) (obj map[string]interface{}) {

	isCacheable := false
	if request != nil {
		if CheckRedisAvailability(request) {
			isCacheable = true
		}
	}

	if isCacheable {
		result := cache.GetByKey(request, cache.Data)
		if result == nil {
			query := "SELECT * FROM " + repository.getDatabaseName(namespace) + "." + class + " WHERE __os_id = '" + id + "';"
			obj, _ = repository.executeQueryOne(request, conn, query, nil)
			if obj == nil || len(obj) == 0 {
				//Data not available.
			} else {
				err := cache.StoreOne(request, obj, cache.Data)
				if err != nil {
					request.Log("Error : " + err.Error())
				}
			}
		} else {
			err := json.Unmarshal(result, &obj)
			if err != nil {
				request.Log("Error : " + err.Error())
			}
		}
	} else {
		query := "SELECT * FROM " + repository.getDatabaseName(namespace) + "." + class + " WHERE __os_id = '" + id + "';"
		obj, _ = repository.executeQueryOne(request, conn, query, nil)
	}

	return
}

func (repository CloudSqlRepository) GetMultipleStoreScripts(conn *sql.DB, request *messaging.ObjectRequest) (query []map[string]interface{}, err error) {
	namespace := request.Controls.Namespace
	class := request.Controls.Class

	noOfElementsPerSet := 1000
	noOfSets := (len(request.Body.Objects) / noOfElementsPerSet)
	remainderFromSets := 0
	remainderFromSets = (len(request.Body.Objects) - (noOfSets * noOfElementsPerSet))

	startIndex := 0
	stopIndex := noOfElementsPerSet

	for x := 0; x < noOfSets; x++ {
		queryOutput := repository.GetMultipleInsertQuery(request, namespace, class, request.Body.Objects[startIndex:stopIndex], conn)
		query = append(query, queryOutput)
		startIndex += noOfElementsPerSet
		stopIndex += noOfElementsPerSet
	}

	if remainderFromSets > 0 {
		start := len(request.Body.Objects) - remainderFromSets
		queryOutput := repository.GetMultipleInsertQuery(request, namespace, class, request.Body.Objects[start:len(request.Body.Objects)], conn)
		query = append(query, queryOutput)
	}

	return
}

func (repository CloudSqlRepository) GetMultipleInsertQuery(request *messaging.ObjectRequest, namespace, class string, records []map[string]interface{}, conn *sql.DB) (queryData map[string]interface{}) {
	queryData = make(map[string]interface{})
	query := ""
	//create insert scripts
	isFirstRow := true
	var keyArray []string
	for _, obj := range records {
		if isFirstRow {
			query += ("INSERT INTO " + repository.getDatabaseName(namespace) + "." + class)
		}

		id := ""

		if obj["OriginalIndex"] == nil {
			id = getNoSqlKeyById(request, obj)
		} else {
			id = obj["OriginalIndex"].(string)
		}

		delete(obj, "OriginalIndex")

		keyList := ""
		valueList := ""

		if isFirstRow {
			for k, _ := range obj {
				keyList += ("," + k)
				keyArray = append(keyArray, k)
			}
		}
		//request.Log(keyArray)
		for _, k := range keyArray {
			v := obj[k]
			valueList += ("," + repository.getSqlFieldValue(v))
		}

		if isFirstRow {
			query += "(__os_id" + keyList + ") VALUES "
		} else {
			query += ","
		}
		query += ("(\"" + id + "\"" + valueList + ")")

		if isFirstRow {
			isFirstRow = false
		}
	}

	queryData["query"] = query
	queryData["queryObject"] = records
	return
}

func (repository CloudSqlRepository) getSingleObjectInsertQuery(request *messaging.ObjectRequest, namespace, class string, obj map[string]interface{}, conn *sql.DB) (query string) {
	var keyArray []string
	query = ""
	query = ("INSERT INTO " + repository.getDatabaseName(namespace) + "." + class)

	id := ""

	if obj["OriginalIndex"] == nil {
		id = getNoSqlKeyById(request, obj)
	} else {
		id = obj["OriginalIndex"].(string)
	}

	delete(obj, "OriginalIndex")

	keyList := ""
	valueList := ""

	for k, _ := range obj {
		keyList += ("," + k)
		keyArray = append(keyArray, k)
	}

	for _, k := range keyArray {
		v := obj[k]
		valueList += ("," + repository.getSqlFieldValue(v))
	}

	query += "(__os_id" + keyList + ") VALUES "
	query += ("(\"" + id + "\"" + valueList + ")")
	return
}

func (repository CloudSqlRepository) getSingleObjectUpdateQuery(request *messaging.ObjectRequest, namespace, class string, obj map[string]interface{}, conn *sql.DB) (query string) {

	updateValues := ""
	isFirst := true
	for k, v := range obj {
		if isFirst {
			isFirst = false
		} else {
			updateValues += ","
		}

		updateValues += (k + "=" + repository.getSqlFieldValue(v))
	}
	query = ("UPDATE " + repository.getDatabaseName(namespace) + "." + class + " SET " + updateValues + " WHERE __os_id=\"" + getNoSqlKeyById(request, obj) + "\";")
	return
}

func (repository CloudSqlRepository) getDeleteScript(namespace string, class string, id string) string {
	return "DELETE FROM " + repository.getDatabaseName(namespace) + "." + class + " WHERE __os_id = \"" + id + "\""
}

func (repository CloudSqlRepository) getCreateScript(namespace string, class string, obj map[string]interface{}) string {

	domain := repository.getDatabaseName(namespace)

	query := "CREATE TABLE IF NOT EXISTS " + domain + "." + class + "(__os_id varchar(255) primary key"

	var textFields []string

	for k, v := range obj {
		if k != "OriginalIndex" {
			dataType := repository.golangToSql(v)
			query += (", " + k + " " + dataType)

			if strings.EqualFold(dataType, "TEXT") {
				textFields = append(textFields, k)
			}
		}
	}

	query += ")"
	return query
}

func (repository CloudSqlRepository) checkAvailabilityDb(request *messaging.ObjectRequest, conn *sql.DB, dbName string) (err error) {
	if availableDbs == nil {
		availableDbs = make(map[string]interface{})
	}

	if CheckRedisAvailability(request) {
		if cache.ExistsKeyValue(request, ("CloudSqlAvailableDbs." + dbName), cache.MetaData) {
			return
		}
	} else {
		if availableDbs[dbName] != nil {
			return
		}
	}

	dbQuery := "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '" + dbName + "'"
	dbResult, err := repository.executeQueryOne(request, conn, dbQuery, nil)

	if err == nil {
		if dbResult["SCHEMA_NAME"] == nil {
			repository.executeNonQuery(conn, "CREATE DATABASE IF NOT EXISTS "+dbName, request)
			repository.executeNonQuery(conn, "create table "+dbName+".domainClassAttributes ( __os_id VARCHAR(255) primary key, class text, maxCount text, version text);", request)
		}

		if CheckRedisAvailability(request) {
			err = cache.StoreKeyValue(request, ("CloudSqlAvailableDbs." + dbName), "true", cache.MetaData)
		} else {
			if availableDbs[dbName] == nil {
				availableDbs[dbName] = true
			}
		}
	} else {
		request.Log("Error : " + err.Error())
	}

	return
}

func (repository CloudSqlRepository) checkAvailabilityTable(request *messaging.ObjectRequest, conn *sql.DB, dbName string, namespace string, class string, obj map[string]interface{}) (err error) {

	if availableTables == nil {
		availableTables = make(map[string]interface{})
	}

	isTableCreatedNow := false

	if CheckRedisAvailability(request) {
		if !cache.ExistsKeyValue(request, ("CloudSqlAvailableTables." + dbName + "." + class), cache.MetaData) {
			var tableResult map[string]interface{}
			tableResult, err = repository.executeQueryOne(request, conn, "SHOW TABLES FROM "+dbName+" LIKE \""+class+"\"", nil)
			if err == nil {
				//if tableResult["Tables_in_"+dbName] == nil {
				if len(tableResult) == 0 {
					script := repository.getCreateScript(namespace, class, obj)
					err, _ = repository.executeNonQuery(conn, script, request)
					if err != nil {
						return
					} else {
						isTableCreatedNow = true
						recordForIDService := "INSERT INTO " + dbName + ".domainClassAttributes (__os_id, class, maxCount,version) VALUES ('" + getDomainClassAttributesKey(request) + "','" + request.Controls.Class + "','0','" + common.GetGUID() + "')"
						_, _ = repository.executeNonQuery(conn, recordForIDService, request)
						keygenerator.CreateNewKeyGenBundle(request)
					}
				}
				if CheckRedisAvailability(request) {
					err = cache.StoreKeyValue(request, ("CloudSqlAvailableTables." + dbName + "." + class), "true", cache.MetaData)
				} else {
					if availableTables[dbName+"."+class] == nil || availableTables[dbName+"."+class] == false {
						availableTables[dbName+"."+class] = true
					}
				}

			} else {
				return
			}
		}
	} else {
		if availableTables[dbName+"."+class] == nil {
			var tableResult map[string]interface{}
			tableResult, err = repository.executeQueryOne(request, conn, "SHOW TABLES FROM "+dbName+" LIKE \""+class+"\"", nil)
			if err == nil {
				if tableResult["Tables_in_"+dbName] == nil {
					script := repository.getCreateScript(namespace, class, obj)
					err, _ = repository.executeNonQuery(conn, script, request)
					if err != nil {
						return
					} else {
						isTableCreatedNow = true
						recordForIDService := "INSERT INTO " + dbName + ".domainClassAttributes (__os_id, class, maxCount,version) VALUES ('" + getDomainClassAttributesKey(request) + "','" + request.Controls.Class + "','0','" + common.GetGUID() + "')"
						_, _ = repository.executeNonQuery(conn, recordForIDService, request)
					}
				}
				if availableTables[dbName+"."+class] == nil || availableTables[dbName+"."+class] == false {
					availableTables[dbName+"."+class] = true
				}

			} else {
				return
			}
		}
	}

	err = repository.buildTableCache(request, conn, dbName, class)

	if !isTableCreatedNow {
		alterColumns := ""

		cacheItem := make(map[string]string)

		if CheckRedisAvailability(request) {
			tableCachePattern := "CloudSqlTableCache." + dbName + "." + request.Controls.Class

			if IsTableCacheKeys := cache.ExistsKeyValue(request, tableCachePattern, cache.MetaData); IsTableCacheKeys {

				byteVal := cache.GetKeyValue(request, tableCachePattern, cache.MetaData)
				err = json.Unmarshal(byteVal, &cacheItem)
				if err != nil {
					request.Log("Error : " + err.Error())
					return
				}
			}

		} else {
			cacheItem = tableCache[dbName+"."+class]
		}

		isFirst := true
		for k, v := range obj {
			if !strings.EqualFold(k, "OriginalIndex") || !strings.EqualFold(k, "__osHeaders") {
				_, ok := cacheItem[k]
				if !ok {
					if isFirst {
						isFirst = false
					} else {
						alterColumns += ", "
					}

					alterColumns += ("ADD COLUMN " + k + " " + repository.golangToSql(v))
					repository.addColumnToTableCache(request, dbName, class, k, repository.golangToSql(v))
					cacheItem[k] = repository.golangToSql(v)
				}
			}
		}

		if len(alterColumns) != 0 && len(alterColumns) != len(obj) {

			alterQuery := "ALTER TABLE " + dbName + "." + class + " " + alterColumns
			err, _ = repository.executeNonQuery(conn, alterQuery, request)
			if err != nil {
				request.Log("Error : " + err.Error())
			}
		}

	}

	return
}

func (repository CloudSqlRepository) addColumnToTableCache(request *messaging.ObjectRequest, dbName string, class string, field string, datatype string) {
	if CheckRedisAvailability(request) {

		byteVal := cache.GetKeyValue(request, ("CloudSqlTableCache." + dbName + "." + request.Controls.Class), cache.MetaData)
		fieldsAndTypes := make(map[string]string)
		err := json.Unmarshal(byteVal, &fieldsAndTypes)
		if err != nil {
			request.Log("Error : " + err.Error())
			return
		}

		fieldsAndTypes[field] = datatype

		err = cache.StoreKeyValue(request, ("CloudSqlTableCache." + dbName + "." + request.Controls.Class), getStringByObject(fieldsAndTypes), cache.MetaData)
		if err != nil {
			request.Log("Error : " + err.Error())
		}
	} else {
		dataMap := make(map[string]string)
		dataMap = tableCache[dbName+"."+class]
		dataMap[field] = datatype
		tableCache[dbName+"."+class] = dataMap
	}
}

func (repository CloudSqlRepository) buildTableCache(request *messaging.ObjectRequest, conn *sql.DB, dbName string, class string) (err error) {
	if tableCache == nil {
		tableCache = make(map[string]map[string]string)
	}

	if !CheckRedisAvailability(request) {
		_, ok := tableCache[dbName+"."+class]

		if !ok {
			var exResult []map[string]interface{}
			exResult, err = repository.executeQueryMany(request, conn, "EXPLAIN "+dbName+"."+class, nil)
			if err == nil {
				newMap := make(map[string]string)

				for _, cRow := range exResult {
					newMap[cRow["Field"].(string)] = cRow["Type"].(string)
				}
				if tableCache[dbName+"."+class] == nil {
					tableCache[dbName+"."+class] = newMap
				}
			}
		} else {
			if len(tableCache[dbName+"."+class]) == 0 {
				var exResult []map[string]interface{}
				exResult, err = repository.executeQueryMany(request, conn, "EXPLAIN "+dbName+"."+class, nil)
				if err == nil {
					newMap := make(map[string]string)
					for _, cRow := range exResult {
						newMap[cRow["Field"].(string)] = cRow["Type"].(string)
					}
					tableCache[dbName+"."+class] = newMap

				}
			}
		}
	} else {
		tableCachePattern := ("CloudSqlTableCache." + dbName + "." + request.Controls.Class)
		IsTableCacheKeys := cache.ExistsKeyValue(request, tableCachePattern, cache.MetaData)
		if !IsTableCacheKeys {
			var exResult []map[string]interface{}
			exResult, err := repository.executeQueryMany(request, conn, "EXPLAIN "+dbName+"."+class, nil)
			if err == nil {
				fieldsAndTypes := make(map[string]string)
				key := "CloudSqlTableCache." + dbName + "." + request.Controls.Class
				for _, cRow := range exResult {
					fieldsAndTypes[cRow["Field"].(string)] = cRow["Type"].(string)
				}
				err = cache.StoreKeyValue(request, key, getStringByObject(fieldsAndTypes), cache.MetaData)
			}
		} else {
			cacheItem := make(map[string]string)
			byteVal := cache.GetKeyValue(request, tableCachePattern, cache.MetaData)
			err = json.Unmarshal(byteVal, &cacheItem)
			if err != nil || len(cacheItem) == 0 {
				var exResult []map[string]interface{}
				exResult, err := repository.executeQueryMany(request, conn, "EXPLAIN "+dbName+"."+class, nil)
				if err == nil {
					fieldsAndTypes := make(map[string]string)
					key := "CloudSqlTableCache." + dbName + "." + request.Controls.Class
					for _, cRow := range exResult {
						fieldsAndTypes[cRow["Field"].(string)] = cRow["Type"].(string)
					}
					err = cache.StoreKeyValue(request, key, getStringByObject(fieldsAndTypes), cache.MetaData)
				}
			}
		}
	}

	return
}

func (repository CloudSqlRepository) checkSchema(request *messaging.ObjectRequest, conn *sql.DB, namespace string, class string, obj map[string]interface{}) {
	dbName := repository.getDatabaseName(namespace)
	err := repository.checkAvailabilityDb(request, conn, dbName)

	if err == nil {
		err := repository.checkAvailabilityTable(request, conn, dbName, namespace, class, obj)

		if err != nil {
			request.Log("Error : " + err.Error())
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////Helper functions/////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////

func (repository CloudSqlRepository) getDatabaseName(namespace string) string {
	return "_" + strings.ToLower(strings.Replace(namespace, ".", "", -1))
}

func (repository CloudSqlRepository) getJson(m interface{}) string {
	bytes, _ := json.Marshal(m)
	return string(bytes[:len(bytes)])
}

func (repository CloudSqlRepository) getSqlFieldValue(value interface{}) string {
	var strValue string
	switch v := value.(type) {
	case bool:
		if value.(bool) == true {
			strValue = "b'1'"
		} else {
			strValue = "b'0'"
		}
		break
	case string:
		sval := fmt.Sprint(value)
		// if strings.ContainsAny(sval, "\"'\n\r\t") {
		if strings.ContainsAny(sval, "\"\n\r\t") {
			sEnc := base64.StdEncoding.EncodeToString([]byte(sval))
			strValue = "'^" + sEnc + "'"
		} else {
			strValue = "'" + sval + "'"
		}
		/*else if (strings.Contains(sval, "'")){
		  		    sEnc := base64.StdEncoding.EncodeToString([]byte(sval))
		      		strValue = "'^" + sEnc + "'";
		  		}*/
		break
	default:
		strValue = "'" + repository.getJson(v) + "'"
		break

	}

	return strValue
}

func (repository CloudSqlRepository) golangToSql(value interface{}) string {

	var strValue string

	//request.Log(reflect.TypeOf(value))
	switch value.(type) {
	case string:
		strValue = "TEXT"
	case bool:
		strValue = "BIT"
		break
	case uint:
		strValue = "INT (10)"
		break
	case int:
		strValue = "INT (10)"
		break
	//case uintptr:
	case uint8:
	case uint16:
	case uint32:
		strValue = "INT (10)"
		break
	case uint64:
		strValue = "INT (10)"
		break
	case int8:
	case int16:
	case int32:
		strValue = "INT (10)"
		break
	case int64:
		strValue = "INT (10)"
		break
	case float32:
		strValue = "DOUBLE"
		break
	case float64:
		strValue = "DOUBLE"
		break
	default:
		strValue = "LONGBLOB"
		break

	}

	return strValue
}

func (repository CloudSqlRepository) sqlToGolang(b []byte, t string) interface{} {

	if b == nil {
		return nil
	}

	if len(b) == 0 {
		return b
	}

	var outData interface{}
	tmp := string(b)
	tType := strings.ToLower(t)
	if strings.Contains(tType, "bit") {
		if len(b) == 0 {
			outData = false
		} else {
			if b[0] == 1 {
				outData = true
			} else {
				outData = false
			}
		}
	} else if strings.Contains(tType, "double") {
		fData, err := strconv.ParseFloat(tmp, 64)
		if err != nil {
			outData = tmp
		} else {
			outData = fData
		}
	} else if strings.Contains(tType, "int") {
		fData, err := strconv.Atoi(tmp)
		if err != nil {
			outData = tmp
		} else {
			outData = fData
		}
	} else {
		if len(tmp) == 4 {
			if strings.ToLower(tmp) == "null" {
				outData = nil
				return outData
			}
		}

		if string(tmp[0]) == "^" {
			byteData := []byte(tmp)
			bdata := string(byteData[1:])
			decData, _ := base64.StdEncoding.DecodeString(bdata)
			outData = repository.getInterfaceValue(string(decData))
		} else {
			outData = repository.getInterfaceValue(tmp)
		}
	}

	return outData
}

func (repository CloudSqlRepository) getInterfaceValue(tmp string) (outData interface{}) {
	var m interface{}
	if string(tmp[0]) == "{" || string(tmp[0]) == "[" {
		err := json.Unmarshal([]byte(tmp), &m)
		if err == nil {
			outData = m
		} else {
			outData = tmp
		}
	} else {
		//outData = tmp
		if tmp == "\u0000" {
			outData = false
		} else if tmp == "\u0001" {
			outData = true
		} else {
			outData = tmp
		}
	}
	return
}

func (repository CloudSqlRepository) rowsToMap(request *messaging.ObjectRequest, rows *sql.Rows, tableName interface{}) (tableMap []map[string]interface{}, err error) {

	columns, _ := rows.Columns()
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

	cacheItem := make(map[string]string)

	if tableName != nil {
		if CheckRedisAvailability(request) {
			tableCachePattern := "CloudSqlTableCache." + tableName.(string)

			if IsTableCacheKeys := cache.ExistsKeyValue(request, tableCachePattern, cache.MetaData); IsTableCacheKeys {

				byteVal := cache.GetKeyValue(request, tableCachePattern, cache.MetaData)
				err = json.Unmarshal(byteVal, &cacheItem)
				if err != nil {
					request.Log("Error : " + err.Error())
					return
				}
			}
		} else {
			tName := tableName.(string)
			cacheItem = tableCache[tName]
		}
	}

	for rows.Next() {

		for i, _ := range columns {
			valuePtrs[i] = &values[i]
		}

		rows.Scan(valuePtrs...)

		rowMap := make(map[string]interface{})

		for i, col := range columns {
			if col == "__os_id" || col == "__osHeaders" {
				continue
			}
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				if cacheItem != nil {
					t, ok := cacheItem[col]
					if ok {
						v = repository.sqlToGolang(b, t)
					}
				}

				if v == nil {
					if b == nil {
						v = nil
					} else if strings.ToLower(string(b)) == "null" {
						v = nil
					} else {
						v = string(b)
					}

				}
			} else {
				v = val
			}
			rowMap[col] = v
		}
		tableMap = append(tableMap, rowMap)
	}

	return
}

func (repository CloudSqlRepository) executeQueryMany(request *messaging.ObjectRequest, conn *sql.DB, query string, tableName interface{}) (result []map[string]interface{}, err error) {
	rows, err := conn.Query(query)

	if err == nil {
		result, err = repository.rowsToMap(request, rows, tableName)
	} else {
		if strings.HasPrefix(err.Error(), "Error 1146") {
			err = nil
			result = make([]map[string]interface{}, 0)
		}
	}

	return
}

func (repository CloudSqlRepository) executeQueryOne(request *messaging.ObjectRequest, conn *sql.DB, query string, tableName interface{}) (result map[string]interface{}, err error) {
	rows, err := conn.Query(query)

	if err == nil {
		var resultSet []map[string]interface{}
		resultSet, err = repository.rowsToMap(request, rows, tableName)
		if len(resultSet) > 0 {
			result = resultSet[0]
		}

	} else {
		if strings.HasPrefix(err.Error(), "Error 1146") {
			err = nil
			result = make(map[string]interface{})
		}
	}

	return
}

func (repository CloudSqlRepository) executeNonQuery(conn *sql.DB, query string, request *messaging.ObjectRequest) (err error, message string) {
	//request.Log("Info Query : " + query)
	tokens := strings.Split(strings.ToLower(query), " ")
	result, err := conn.Exec(query)
	if err == nil {
		val, _ := result.RowsAffected()
		if val <= 0 && (tokens[0] == "delete" || tokens[0] == "update") {
			message = "No Rows Changed"
		}
	}
	return
}

func (repository CloudSqlRepository) getRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
	isGUIDKey := false
	isAutoIncrementId := false //else MANUAL key from the user

	//multiple requests
	if (obj[request.Body.Parameters.KeyProperty].(string) == "-999") || (request.Body.Parameters.AutoIncrement == true) {
		isAutoIncrementId = true
	}

	if (obj[request.Body.Parameters.KeyProperty].(string) == "-888") || (request.Body.Parameters.GUIDKey == true) {
		isGUIDKey = true
	}

	if isGUIDKey {
		returnID = common.GetGUID()
	} else if isAutoIncrementId {
		if CheckRedisAvailability(request) {
			returnID = keygenerator.GetIncrementID(request, "CLOUDSQL", 0)
		} else {
			request.Log("Debug : WARNING! : Returning GUID since REDIS not available and not concurrent safe!")
			returnID = common.GetGUID()
		}
	} else {
		returnID = obj[request.Body.Parameters.KeyProperty].(string)
	}
	return
}

func (repository CloudSqlRepository) closeConnection(conn *sql.DB) {
	// err := conn.Close()
	// if err != nil {
	// 	request.Log(err.Error())
	// } else {
	// 	request.Log("Connection Closed!")
	// }
}
