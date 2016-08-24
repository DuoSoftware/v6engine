package repositories

import (
	"duov6.com/common"
	"duov6.com/objectstore/cache"
	"duov6.com/objectstore/keygenerator"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/security"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
	"strconv"
	"strings"
	"sync"
)

type CassandraRepository struct {
}

func (repository CassandraRepository) GetRepositoryName() string {
	return "Cassandra DB"
}

var cassandraConnections map[string]*gocql.Session
var cassandraConnectionLock = sync.RWMutex{}

var cassandraTableCache map[string]map[string]string
var cassandraTableCacheLock = sync.RWMutex{}

var cassandraAvailableTables map[string]interface{}
var cassandraAvailableTablesLock = sync.RWMutex{}

// Start of GET and SET methods

func (repository CassandraRepository) GetCassandraConnections(index string) (conn *gocql.Session) {
	cassandraConnectionLock.RLock()
	defer cassandraConnectionLock.RUnlock()
	conn = cassandraConnections[index]
	return
}

func (repository CassandraRepository) SetCassandraConnections(index string, conn *gocql.Session) {
	cassandraConnectionLock.Lock()
	defer cassandraConnectionLock.Unlock()
	cassandraConnections[index] = conn
}

func (repository CassandraRepository) GetCassandraAvailableTables(index string) (value interface{}) {
	cassandraAvailableTablesLock.RLock()
	defer cassandraAvailableTablesLock.RUnlock()
	value = cassandraAvailableTables[index]
	return
}

func (repository CassandraRepository) SetCassandraAvailabaleTables(index string, value interface{}) {
	cassandraAvailableTablesLock.Lock()
	defer cassandraAvailableTablesLock.Unlock()
	cassandraAvailableTables[index] = value
}

func (repository CassandraRepository) GetCassandraTableCache(index string) (value map[string]string) {
	cassandraTableCacheLock.RLock()
	defer cassandraTableCacheLock.RUnlock()
	value = cassandraTableCache[index]
	return
}

func (repository CassandraRepository) SetCassandraTableCache(index string, value map[string]string) {
	cassandraTableCacheLock.Lock()
	defer cassandraTableCacheLock.Unlock()
	cassandraTableCache[index] = value
}

// End of GET and SET methods

func (repository CassandraRepository) GetConnection(request *messaging.ObjectRequest) (conn *gocql.Session, err error) {

	if cassandraConnections == nil {
		cassandraConnections = make(map[string]*gocql.Session)
	}

	URL := request.Configuration.ServerConfiguration["CASSANDRA"]["Url"]

	poolPattern := URL + request.Controls.Namespace

	if repository.GetCassandraConnections(poolPattern) == nil {
		conn, err = repository.CreateConnection(request)
		if err != nil {
			request.Log("Error : " + err.Error())
			return
		}
		repository.SetCassandraConnections(poolPattern, conn)
	} else {
		conStatus := repository.GetCassandraConnections(poolPattern).Closed()
		if conStatus == true {
			repository.SetCassandraConnections(poolPattern, nil)
			conn, err = repository.CreateConnection(request)
			if err != nil {
				request.Log("Error : " + err.Error())
				return
			}
			repository.SetCassandraConnections(poolPattern, conn)
		} else {
			conn = repository.GetCassandraConnections(poolPattern)
		}
	}
	return
}

func (repository CassandraRepository) CreateConnection(request *messaging.ObjectRequest) (conn *gocql.Session, err error) {
	keyspace := repository.GetDatabaseName(request.Controls.Namespace)
	cluster := gocql.NewCluster(request.Configuration.ServerConfiguration["CASSANDRA"]["Url"])
	cluster.Keyspace = keyspace

	conn, err = cluster.CreateSession()
	if err != nil {
		request.Log("Error : Cassandra connection initilizing failed!")
		err = repository.CreateNewKeyspace(request)
		if err != nil {
			request.Log("Error : " + err.Error())
		} else {
			return repository.CreateConnection(request)
		}
	}
	return
}

// Helper Function
func (repository CassandraRepository) CreateNewKeyspace(request *messaging.ObjectRequest) (err error) {
	keyspace := repository.GetDatabaseName(request.Controls.Namespace)
	//Log to Default SYSTEM Keyspace
	cluster := gocql.NewCluster(request.Configuration.ServerConfiguration["CASSANDRA"]["Url"])
	cluster.Keyspace = "system"
	var conn *gocql.Session
	conn, err = cluster.CreateSession()
	if err != nil {
		request.Log("Error : Cassandra connection to SYSTEM keyspace initilizing failed!")
	} else {
		err = conn.Query("CREATE KEYSPACE IF NOT EXISTS " + keyspace + " WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };").Exec()
		if err != nil {
			request.Log("Error : Failed to create new " + keyspace + " Keyspace : " + err.Error())
		} else {
			request.Log("Debug : Created new " + keyspace + " Keyspace")
			err = conn.Query("create table IF NOT EXISTS " + keyspace + ".domainClassAttributes (os_id text, class text, maxCount text, version text, PRIMARY KEY(os_id));").Exec()
			conn.Close()
		}
	}
	return
}

//.................................................................................

func (repository CassandraRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {

	take := "100"

	if request.Extras["take"] != nil {
		take = request.Extras["take"].(string)
	}

	query := "SELECT * FROM " + repository.GetDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class

	query += " limit " + take

	query += ";"

	response := repository.queryCommonMany(query, request)
	return response
}

func (repository CassandraRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	if request.Body.Query.Parameters != "*" {
		response = getDefaultNotImplemented()
	} else {
		response = repository.GetAll(request)
	}
	return response
}

func (repository CassandraRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	if security.ValidateSecurity(getNoSqlKey(request)) {
		response.GetResponseWithBody(getEmptyByteObject())
		request.Log("Error! Security Violation of request detected. Aborting request with error!")
		return response
	}

	query := "SELECT * FROM " + repository.GetDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class + " WHERE os_id = '" + getNoSqlKey(request) + "';"
	response = repository.queryCommonOne(query, request)
	return response
}

func (repository CassandraRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository CassandraRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {

	var response RepositoryResponse

	conn, err := repository.GetConnection(request)
	if err != nil {
		response.IsSuccess = false
		response.Message = err.Error()
		return response
	}

	var idData map[string]interface{}
	idData = make(map[string]interface{})

	for index, obj := range request.Body.Objects {
		id := repository.GetRecordID(request, obj)
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

func (repository CassandraRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {

	var response RepositoryResponse

	request.Body.Object["osHeaders"] = request.Body.Object["__osHeaders"]
	delete(request.Body.Object, "__osHeaders")

	conn, err := repository.GetConnection(request)
	if err != nil {
		response.IsSuccess = false
		response.Message = err.Error()
		return response
	}

	id := repository.GetRecordID(request, request.Body.Object)
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

	request.Body.Object["__osHeaders"] = request.Body.Object["osHeaders"]
	delete(request.Body.Object, "osHeaders")

	return response
}

func (repository CassandraRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {

	var response RepositoryResponse

	conn, err := repository.GetConnection(request)
	if err != nil {
		response.IsSuccess = false
		response.Message = err.Error()
		return response
	}

	var idData map[string]interface{}
	idData = make(map[string]interface{})

	for index, obj := range request.Body.Objects {
		id := repository.GetRecordID(request, obj)
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

func (repository CassandraRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {

	var response RepositoryResponse

	conn, err := repository.GetConnection(request)
	if err != nil {
		response.IsSuccess = false
		response.Message = err.Error()
		return response
	}

	id := repository.GetRecordID(request, request.Body.Object)
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

func (repository CassandraRepository) ReRun(request *messaging.ObjectRequest, conn *gocql.Session, obj map[string]interface{}) RepositoryResponse {
	var response RepositoryResponse

	repository.CheckSchema(request, conn, request.Controls.Namespace, request.Controls.Class, obj)
	response = repository.queryStore(request)
	if !response.IsSuccess {
		if CheckRedisAvailability(request) {
			cache.FlushCache(request)
		} else {
			cassandraTableCache = make(map[string]map[string]string)
			cassandraAvailableTables = make(map[string]interface{})
		}
		repository.CheckSchema(request, conn, request.Controls.Namespace, request.Controls.Class, obj)
		response = repository.queryStore(request)
	}

	return response
}

func (repository CassandraRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {

	response := RepositoryResponse{}
	conn, err := repository.GetConnection(request)
	if err == nil {
		isError := false
		for _, obj := range request.Body.Objects {
			query := repository.GetDeleteScript(request.Controls.Namespace, request.Controls.Class, getNoSqlKeyById(request, obj))
			err, message := repository.ExecuteNonQuery(conn, query, request)
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
	repository.CloseConnection(conn)
	return response
}

func (repository CassandraRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {

	response := RepositoryResponse{}
	conn, err := repository.GetConnection(request)
	if err == nil {
		query := repository.GetDeleteScript(request.Controls.Namespace, request.Controls.Class, getNoSqlKey(request))
		err, message := repository.ExecuteNonQuery(conn, query, request)
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
	repository.CloseConnection(conn)
	return response
}

func (repository CassandraRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {

	response := RepositoryResponse{}
	queryType := request.Body.Special.Type
	queryType = strings.ToLower(queryType)
	domain := repository.GetDatabaseName(request.Controls.Namespace)

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
	case "GetDatabaseNames":
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
		conn, err := repository.GetConnection(request)
		if err == nil {
			query := "DROP TABLE " + domain + "." + request.Controls.Class
			err, _ := repository.ExecuteNonQuery(conn, query, request)
			if err != nil {
				response.IsSuccess = false
				response.Message = "Error Dropping Table in CloudSQL Repository : " + err.Error()
			} else {
				//Delete Class from availableTables and tablecache
				if CheckRedisAvailability(request) {
					_ = cache.DeleteKey(request, ("CassandraTableCache." + domain + "." + request.Controls.Class), cache.MetaData)
					_ = cache.DeleteKey(request, ("CassandraAvailableTables." + domain + "." + request.Controls.Class), cache.MetaData)
					_ = cache.DeletePattern(request, (domain + "." + request.Controls.Class + "*"), cache.Data)
				} else {
					delete(cassandraAvailableTables, (domain + "." + request.Controls.Class))
					delete(cassandraTableCache, (domain + "." + request.Controls.Class))
				}
				response.IsSuccess = true
				response.Message = "Successfully Dropped Table : " + request.Controls.Class
			}
		} else {
			response.IsSuccess = false
			response.Message = "Connection Failed to CloudSQL Server"
		}
		repository.CloseConnection(conn)
	case "dropnamespace":
		request.Log("Debug : Starting Delete-Database sub routine")
		conn, err := repository.GetConnection(request)
		if err == nil {
			query := "DROP SCHEMA " + domain
			err, _ := repository.ExecuteNonQuery(conn, query, request)
			if err != nil {
				response.IsSuccess = false
				response.Message = "Error Dropping Table in CloudSQL Repository : " + err.Error()
			} else {
				if CheckRedisAvailability(request) {

					_ = cache.DeleteKey(request, ("CassandraTableCache." + domain + "." + request.Controls.Class), cache.MetaData)

					var availableTablesKeys []string
					availableTablesPattern := "CassandraAvailableTables." + domain + ".*"
					availableTablesKeys = cache.GetKeyListPattern(request, availableTablesPattern, cache.MetaData)
					if len(availableTablesKeys) > 0 {
						for _, name := range availableTablesKeys {
							_ = cache.DeleteKey(request, name, cache.MetaData)
						}
					}
					_ = cache.DeletePattern(request, (domain + "*"), cache.Data)

				} else {
					//Delete all associated Classes from it's TableCache and availableTables
					for key, _ := range cassandraAvailableTables {
						if strings.Contains(key, domain) {
							delete(cassandraAvailableTables, key)
							delete(cassandraTableCache, key)
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
		repository.CloseConnection(conn)
	case "flushcache":
		if CheckRedisAvailability(request) {
			cache.FlushCache(request)
		} else {
			cassandraTableCache = make(map[string]map[string]string)
			cassandraAvailableTables = make(map[string]interface{})
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
		response = getDefaultNotImplemented()
		return response
	case "uniqueindex":
		response = getDefaultNotImplemented()
		return response
	default:
		response = getDefaultNotImplemented()
	}

	return response
}

func (repository CassandraRepository) Test(request *messaging.ObjectRequest) {
}

func (repository CassandraRepository) ClearCache(request *messaging.ObjectRequest) {
	if CheckRedisAvailability(request) {
		cache.FlushCache(request)
	} else {
		cassandraTableCache = make(map[string]map[string]string)
		cassandraAvailableTables = make(map[string]interface{})
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////SQL GENERATORS/////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////
func (repository CassandraRepository) queryCommon(query string, request *messaging.ObjectRequest, isOne bool) RepositoryResponse {
	response := RepositoryResponse{}

	request.Log("Info Query : " + query)

	conn, err := repository.GetConnection(request)
	if err == nil {
		var err error
		dbName := repository.GetDatabaseName(request.Controls.Namespace)
		err = repository.BuildTableCache(request, conn, dbName, request.Controls.Class)
		if err != nil {
			request.Log("Error : " + err.Error())
		}

		var obj interface{}
		tableName := dbName + "." + request.Controls.Class
		if isOne {
			obj, err = repository.ExecuteQueryOne(request, conn, query, tableName)
		} else {
			obj, err = repository.ExecuteQueryMany(request, conn, query, tableName)
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
	repository.CloseConnection(conn)
	return response
}

func (repository CassandraRepository) queryCommonMany(query string, request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryCommon(query, request, false)

}

func (repository CassandraRepository) queryCommonOne(query string, request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryCommon(query, request, true)
}

func (repository CassandraRepository) queryStore(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	conn, _ := repository.GetConnection(request)

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

		insertScript := repository.GetSingleObjectInsertQuery(request, domain, class, obj, conn)
		err, _ := repository.ExecuteNonQuery(conn, insertScript, request)
		if err != nil {
			if !strings.Contains(err.Error(), "specified twice") {
				updateScript := repository.GetSingleObjectUpdateQuery(request, domain, class, obj, conn)
				err, message := repository.ExecuteNonQuery(conn, updateScript, request)
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

				err, _ := repository.ExecuteNonQuery(conn, script, request)
				if err != nil {
					request.Log("Error : " + err.Error())
					if strings.Contains(err.Error(), "Duplicate entry") {
						errorBlock := scripts[x]["queryObject"].([]map[string]interface{})
						for _, singleQueryObject := range errorBlock {
							insertScript := repository.GetSingleObjectInsertQuery(request, domain, class, singleQueryObject, conn)
							err1, _ := repository.ExecuteNonQuery(conn, insertScript, request)
							if err1 != nil {
								if !strings.Contains(err.Error(), "specified twice") {
									updateScript := repository.GetSingleObjectUpdateQuery(request, domain, class, singleQueryObject, conn)
									err2, message := repository.ExecuteNonQuery(conn, updateScript, request)
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

	repository.CloseConnection(conn)
	return response
}

func (repository CassandraRepository) GetMultipleStoreScripts(conn *gocql.Session, request *messaging.ObjectRequest) (query []map[string]interface{}, err error) {
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

func (repository CassandraRepository) GetMultipleInsertQuery(request *messaging.ObjectRequest, namespace, class string, records []map[string]interface{}, conn *gocql.Session) (queryData map[string]interface{}) {
	queryData = make(map[string]interface{})
	query := ""
	//create insert scripts
	isFirstRow := true
	var keyArray []string
	for _, obj := range records {
		if isFirstRow {
			query += ("INSERT INTO " + repository.GetDatabaseName(namespace) + "." + class)
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
			valueList += ("," + repository.GetSqlFieldValue(v))
		}

		if isFirstRow {
			query += "(os_id" + keyList + ") VALUES "
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

func (repository CassandraRepository) GetSingleObjectInsertQuery(request *messaging.ObjectRequest, namespace, class string, obj map[string]interface{}, conn *gocql.Session) (query string) {
	var keyArray []string
	query = ""
	query = ("INSERT INTO " + repository.GetDatabaseName(namespace) + "." + class)

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
		valueList += ("," + repository.GetSqlFieldValue(v))
	}

	query += "(os_id" + keyList + ") VALUES "
	query += ("('" + id + "'" + valueList + ");")
	return
}

func (repository CassandraRepository) GetSingleObjectUpdateQuery(request *messaging.ObjectRequest, namespace, class string, obj map[string]interface{}, conn *gocql.Session) (query string) {

	updateValues := ""
	isFirst := true
	for k, v := range obj {
		if isFirst {
			isFirst = false
		} else {
			updateValues += ","
		}

		updateValues += (k + "=" + repository.GetSqlFieldValue(v))
	}
	query = ("UPDATE " + repository.GetDatabaseName(namespace) + "." + class + " SET " + updateValues + " WHERE os_id=\"" + getNoSqlKeyById(request, obj) + "\";")
	return
}

func (repository CassandraRepository) GetDeleteScript(namespace string, class string, id string) string {
	return "DELETE FROM " + repository.GetDatabaseName(namespace) + "." + class + " WHERE os_id = '" + id + "'"
}

func (repository CassandraRepository) GetCreateScript(namespace string, class string, obj map[string]interface{}) string {

	domain := repository.GetDatabaseName(namespace)

	query := "CREATE TABLE IF NOT EXISTS " + domain + "." + class + " (os_id text"

	var textFields []string

	for k, v := range obj {
		if k != "OriginalIndex" {
			dataType := repository.GolangToSql(v)
			query += (", " + k + " " + dataType)

			if strings.EqualFold(dataType, "TEXT") {
				textFields = append(textFields, k)
			}
		}
	}

	query += ", PRIMARY KEY(os_id));"
	return query
}

func (repository CassandraRepository) CheckAvailabilityTable(request *messaging.ObjectRequest, conn *gocql.Session, dbName string, namespace string, class string, obj map[string]interface{}) (err error) {

	if cassandraAvailableTables == nil {
		cassandraAvailableTables = make(map[string]interface{})
	}

	isTableCreatedNow := false

	if CheckRedisAvailability(request) {
		if !cache.ExistsKeyValue(request, ("CassandraAvailableTables." + dbName + "." + class), cache.MetaData) {
			var tableResult map[string]interface{}
			tableResult, err = repository.ExecuteQueryOne(request, conn, ("select columnfamily_name from system.schema_columnfamilies WHERE keyspace_name='" + dbName + "' AND columnfamily_name='" + class + "';"), nil)
			if err == nil {
				//if tableResult["Tables_in_"+dbName] == nil {
				if len(tableResult) == 0 {
					script := repository.GetCreateScript(namespace, class, obj)
					err, _ = repository.ExecuteNonQuery(conn, script, request)
					if err != nil {
						return
					} else {
						isTableCreatedNow = true
						recordForIDService := "INSERT INTO " + dbName + ".domainClassAttributes (os_id, class, maxCount,version) VALUES ('" + getDomainClassAttributesKey(request) + "','" + request.Controls.Class + "','0','" + common.GetGUID() + "')"
						_, _ = repository.ExecuteNonQuery(conn, recordForIDService, request)
						keygenerator.CreateNewKeyGenBundle(request)
					}
				}
				if CheckRedisAvailability(request) {
					err = cache.StoreKeyValue(request, ("CassandraAvailableTables." + dbName + "." + class), "true", cache.MetaData)
				} else {
					keyword := dbName + "." + class
					availableTableValue := repository.GetCassandraAvailableTables(keyword)
					if availableTableValue == nil || availableTableValue.(bool) == false {
						repository.SetCassandraAvailabaleTables(keyword, true)
					}
				}

			} else {
				return
			}
		}
	} else {
		keyword := dbName + "." + class
		availableTableValue := repository.GetCassandraAvailableTables(keyword)
		if availableTableValue == nil {
			var tableResult map[string]interface{}
			tableResult, err = repository.ExecuteQueryOne(request, conn, ("select columnfamily_name from system.schema_columnfamilies WHERE keyspace_name='" + dbName + "' AND columnfamily_name='" + class + "';"), nil)
			if err == nil {
				if tableResult["Tables_in_"+dbName] == nil {
					script := repository.GetCreateScript(namespace, class, obj)
					err, _ = repository.ExecuteNonQuery(conn, script, request)
					if err != nil {
						return
					} else {
						isTableCreatedNow = true
						recordForIDService := "INSERT INTO " + dbName + ".domainClassAttributes (os_id, class, maxCount,version) VALUES ('" + getDomainClassAttributesKey(request) + "','" + request.Controls.Class + "','0','" + common.GetGUID() + "')"
						_, _ = repository.ExecuteNonQuery(conn, recordForIDService, request)
					}
				}
				if availableTableValue == nil || availableTableValue.(bool) == false {
					repository.SetCassandraAvailabaleTables(keyword, true)
				}

			} else {
				return
			}
		}
	}

	err = repository.BuildTableCache(request, conn, dbName, class)

	if !isTableCreatedNow {
		alterColumns := ""

		cacheItem := make(map[string]string)

		if CheckRedisAvailability(request) {
			tableCachePattern := "CassandraTableCache." + dbName + "." + request.Controls.Class

			if IsTableCacheKeys := cache.ExistsKeyValue(request, tableCachePattern, cache.MetaData); IsTableCacheKeys {

				byteVal := cache.GetKeyValue(request, tableCachePattern, cache.MetaData)
				err = json.Unmarshal(byteVal, &cacheItem)
				if err != nil {
					request.Log("Error : " + err.Error())
					return
				}
			}

		} else {
			cacheItem = repository.GetCassandraTableCache(dbName + "." + class)
		}

		isFirst := true
		for k, v := range obj {
			if !strings.EqualFold(k, "OriginalIndex") || !strings.EqualFold(k, "osHeaders") {
				_, ok := cacheItem[k]
				if !ok {
					if isFirst {
						isFirst = false
					} else {
						alterColumns += ", "
					}

					alterColumns += ("ADD COLUMN " + k + " " + repository.GolangToSql(v))
					repository.AddColumnToTableCache(request, dbName, class, k, repository.GolangToSql(v))
					cacheItem[k] = repository.GolangToSql(v)
				}
			}
		}

		if len(alterColumns) != 0 && len(alterColumns) != len(obj) {

			alterQuery := "ALTER TABLE " + dbName + "." + class + " " + alterColumns
			err, _ = repository.ExecuteNonQuery(conn, alterQuery, request)
			if err != nil {
				request.Log("Error : " + err.Error())
			}
		}

	}

	return
}

func (repository CassandraRepository) AddColumnToTableCache(request *messaging.ObjectRequest, dbName string, class string, field string, datatype string) {
	if CheckRedisAvailability(request) {

		byteVal := cache.GetKeyValue(request, ("CassandraTableCache." + dbName + "." + request.Controls.Class), cache.MetaData)
		fieldsAndTypes := make(map[string]string)
		err := json.Unmarshal(byteVal, &fieldsAndTypes)
		if err != nil {
			request.Log("Error : " + err.Error())
			return
		}

		fieldsAndTypes[field] = datatype

		err = cache.StoreKeyValue(request, ("CassandraTableCache." + dbName + "." + request.Controls.Class), getStringByObject(fieldsAndTypes), cache.MetaData)
		if err != nil {
			request.Log("Error : " + err.Error())
		}
	} else {
		dataMap := make(map[string]string)
		dataMap = repository.GetCassandraTableCache(dbName + "." + class)
		dataMap[field] = datatype
		repository.SetCassandraTableCache(dbName+"."+class, dataMap)
	}
}

func (repository CassandraRepository) BuildTableCache(request *messaging.ObjectRequest, conn *gocql.Session, dbName string, class string) (err error) {
	if cassandraTableCache == nil {
		cassandraTableCache = make(map[string]map[string]string)
	}

	if !CheckRedisAvailability(request) {
		var ok bool
		tableCacheLocalEntry := repository.GetCassandraTableCache(dbName + "." + class)
		if tableCacheLocalEntry != nil {
			ok = true
		}

		if !ok {
			var exResult []map[string]interface{}
			exResult, err = repository.ExplainTable(request, conn)
			if err == nil {
				newMap := make(map[string]string)

				for _, cRow := range exResult {
					newMap[cRow["Field"].(string)] = cRow["Type"].(string)
				}

				if repository.GetCassandraTableCache(dbName+"."+class) == nil {
					repository.SetCassandraTableCache(dbName+"."+class, newMap)
				}
			}
		} else {
			if len(tableCacheLocalEntry) == 0 {
				var exResult []map[string]interface{}
				exResult, err = repository.ExplainTable(request, conn)
				if err == nil {
					newMap := make(map[string]string)
					for _, cRow := range exResult {
						newMap[cRow["Field"].(string)] = cRow["Type"].(string)
					}

					repository.SetCassandraTableCache(dbName+"."+class, newMap)
				}
			}
		}
	} else {
		tableCachePattern := ("CassandraTableCache." + dbName + "." + request.Controls.Class)
		IsTableCacheKeys := cache.ExistsKeyValue(request, tableCachePattern, cache.MetaData)
		if !IsTableCacheKeys {
			var exResult []map[string]interface{}
			exResult, err := repository.ExplainTable(request, conn)
			if err == nil {
				fieldsAndTypes := make(map[string]string)
				key := "CassandraTableCache." + dbName + "." + request.Controls.Class
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
				exResult, err := repository.ExplainTable(request, conn)
				if err == nil {
					fieldsAndTypes := make(map[string]string)
					key := "CassandraTableCache." + dbName + "." + request.Controls.Class
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

func (repository CassandraRepository) ExplainTable(request *messaging.ObjectRequest, conn *gocql.Session) (retMap []map[string]interface{}, err error) {
	retMap = make([]map[string]interface{}, 0)
	query := "select column_name,validator from system.schema_columns WHERE keyspace_name='" + repository.GetDatabaseName(request.Controls.Namespace) + "' AND columnfamily_name='" + request.Controls.Class + "';"

	resultSet, err := repository.ExecuteQueryMany(request, conn, query, nil)
	if err != nil {
		request.Log("Error : " + err.Error())
	} else {

		for _, mapVal := range resultSet {
			validator := strings.Replace(mapVal["validator"].(string), "org.apache.cassandra.db.marshal.", "", -1)

			columnMap := make(map[string]interface{})
			columnName := mapVal["column_name"].(string)
			dataType := ""
			switch validator {
			case "Int32Type":
				dataType = "INT"
			case "BytesType":
				dataType = "BLOB"
			case "BooleanType":
				dataType = "BOOLEAN"
			case "DoubleType":
				dataType = "DOUBLE"
			case "UTF8Type":
				dataType = "TEXT"
			default:
				dataType = "BLOB"
			}
			columnMap["Field"] = columnName
			columnMap["Type"] = dataType
			retMap = append(retMap, columnMap)
		}

	}

	fmt.Println(retMap)

	return
}

func (repository CassandraRepository) CheckSchema(request *messaging.ObjectRequest, conn *gocql.Session, namespace string, class string, obj map[string]interface{}) {
	dbName := repository.GetDatabaseName(namespace)

	err := repository.CheckAvailabilityTable(request, conn, dbName, namespace, class, obj)

	if err != nil {
		request.Log("Error : " + err.Error())
	}

}

////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////Helper functions/////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////

func (repository CassandraRepository) GetDatabaseName(namespace string) string {
	namespace = strings.Replace(namespace, ".", "", -1)
	namespace = "db_" + namespace
	return strings.ToLower(namespace)
}

func (repository CassandraRepository) GetJson(m interface{}) string {
	bytes, _ := json.Marshal(m)
	return string(bytes[:len(bytes)])
}

func (repository CassandraRepository) GetSqlFieldValue(value interface{}) string {
	var strValue string
	switch v := value.(type) {
	case bool:
		if value.(bool) == true {
			strValue = "true"
		} else {
			strValue = "false"
		}
		break
	case float64:
		strValue = strconv.FormatFloat(value.(float64), 'f', -1, 64)
		break
	case float32:
		strValue = strconv.FormatFloat(value.(float64), 'f', -1, 32)
		break
	case int64:
		strValue = strconv.FormatInt(value.(int64), 10)
		break
	case int32:
		intVal := value.(int)
		strValue = strconv.Itoa(intVal)
		break
	case int:
		intVal := value.(int)
		strValue = strconv.Itoa(intVal)
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
		strValue = "textAsBlob('" + repository.GetJson(v) + "')"
		break

	}

	return strValue
}

func (repository CassandraRepository) GolangToSql(value interface{}) string {

	var strValue string

	//request.Log(reflect.TypeOf(value))
	switch value.(type) {
	case string:
		strValue = "TEXT"
	case bool:
		strValue = "BOOLEAN"
		break
	case uint:
		strValue = "INT"
		break
	case int:
		strValue = "INT"
		break
	//case uintptr:
	case uint8:
	case uint16:
	case uint32:
		strValue = "INT"
		break
	case uint64:
		strValue = "INT"
		break
	case int8:
	case int16:
	case int32:
		strValue = "INT"
		break
	case int64:
		strValue = "INT"
		break
	case float32:
		strValue = "DOUBLE"
		break
	case float64:
		strValue = "DOUBLE"
		break
	default:
		strValue = "BLOB"
		break

	}

	return strValue
}

func (repository CassandraRepository) SqlToGolang(b []byte, t string) interface{} {

	if b == nil {
		return nil
	}

	if len(b) == 0 {
		return b
	}

	var outData interface{}
	tmp := string(b)
	tType := strings.ToLower(t)
	if strings.Contains(tType, "bool") {
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
			outData = repository.GetInterfaceValue(string(decData))
		} else {
			outData = repository.GetInterfaceValue(tmp)
		}
	}

	return outData
}

func (repository CassandraRepository) GetInterfaceValue(tmp string) (outData interface{}) {
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

func (repository CassandraRepository) RowsToMap(request *messaging.ObjectRequest, rows []map[string]interface{}, tableName interface{}) (tableMap []map[string]interface{}, err error) {

	tableMap = make([]map[string]interface{}, 0)

	for _, miniMap := range rows {
		tempMap := make(map[string]interface{})
		for key, value := range miniMap {
			switch value.(type) {
			case []uint8:
				var data interface{}
				err = json.Unmarshal(value.([]byte), &data)
				tempMap[key] = data
				break
			default:
				tempMap[key] = value
				break
			}
		}
		tableMap = append(tableMap, tempMap)
	}

	return
}

func (repository CassandraRepository) ExecuteQueryMany(request *messaging.ObjectRequest, conn *gocql.Session, query string, tableName interface{}) (result []map[string]interface{}, err error) {
	result = make([]map[string]interface{}, 0)

	iter := conn.Query(query).Iter()
	result, err = iter.SliceMap()
	iter.Close()
	result, err = repository.RowsToMap(request, result, nil)
	return
}

func (repository CassandraRepository) ExecuteQueryOne(request *messaging.ObjectRequest, conn *gocql.Session, query string, tableName interface{}) (result map[string]interface{}, err error) {
	resultSet := make([]map[string]interface{}, 0)
	result = make(map[string]interface{})

	iter := conn.Query(query).Iter()
	resultSet, err = iter.SliceMap()
	resultSet, err = repository.RowsToMap(request, resultSet, nil)
	if err == nil {
		if len(resultSet) > 0 {
			result = resultSet[0]
		}

	} else {
		// if strings.HasPrefix(err.Error(), "Error 1146") {
		// 	err = nil
		// 	result = make(map[string]interface{})
		// }
	}

	iter.Close()

	return
}

func (repository CassandraRepository) ExecuteNonQuery(conn *gocql.Session, query string, request *messaging.ObjectRequest) (err error, message string) {
	request.Log("Debug Query : " + query)
	//tokens := strings.Split(strings.ToLower(query), " ")
	err = conn.Query(query).Exec()
	// if err == nil {
	// 	val, _ := result.RowsAffected()
	// 	if val <= 0 && (tokens[0] == "delete" || tokens[0] == "update") {
	// 		message = "No Rows Changed"
	// 	}
	// }
	return
}

func (repository CassandraRepository) GetRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
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

func (repository CassandraRepository) CloseConnection(conn *gocql.Session) {
	// err := conn.Close()
	// if err != nil {
	// 	request.Log(err.Error())
	// } else {
	// 	request.Log("Connection Closed!")
	// }
}
