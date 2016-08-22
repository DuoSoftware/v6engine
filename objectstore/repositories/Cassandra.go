package repositories

import (
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
	"reflect"
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

// End of GET and SET methods

func (repository CassandraRepository) GetNamespace(namespace string) string {
	namespace = strings.Replace(namespace, ".", "", -1)
	namespace += "db_"
	return strings.ToLower(namespace)
}

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
	keyspace := repository.GetNamespace(request.Controls.Namespace)
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
	keyspace := repository.GetNamespace(request.Controls.Namespace)
	//Log to Default SYSTEM Keyspace
	cluster := gocql.NewCluster(request.Configuration.ServerConfiguration["CASSANDRA"]["Url"])
	cluster.Keyspace = "system"
	var conn *gocql.Session
	conn, err = cluster.CreateSession()
	if err != nil {
		request.Log("Error : Cassandra connection to SYSTEM keyspace initilizing failed!")
	} else {
		err = conn.Query("CREATE KEYSPACE " + keyspace + " WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };").Exec()
		if err != nil {
			request.Log("Error : Failed to create new " + keyspace + " Keyspace : " + err.Error())
		} else {
			request.Log("Debug : Created new " + keyspace + " Keyspace")
		}
	}
	return
}

func (repository CassandraRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-ALL!")
	response := RepositoryResponse{}
	session, err := repository.GetConnection(request)
	if err != nil {
		response.GetErrorResponse(err.Error())
	} else {
		iter2 := session.Query("SELECT * FROM " + strings.ToLower(request.Controls.Class)).Iter()

		my, isErr := iter2.SliceMap()

		if isErr != nil {
			response.IsSuccess = true
			response.Message = isErr.Error()
			fmt.Println(isErr.Error())
			response.Message = "No objects found in Cassandra"
			var emptyMap map[string]interface{}
			emptyMap = make(map[string]interface{})
			byte, _ := json.Marshal(emptyMap)
			response.GetResponseWithBody(byte)
			return response
		}

		iter2.Close()

		skip := 0

		if request.Extras["skip"] != nil {
			skip, _ = strconv.Atoi(request.Extras["skip"].(string))
		}

		take := len(my)

		if request.Extras["take"] != nil {
			take, _ = strconv.Atoi(request.Extras["take"].(string))
		}

		fmt.Println(reflect.TypeOf(my))

		if request.Controls.SendMetaData == "false" {

			for index, arrVal := range my {
				for key, _ := range arrVal {
					if key == "osheaders" {
						delete(my[index], key)
					}
				}
			}
		}

		byteValue, errMarshal := json.Marshal(my[skip:(skip + take)])
		if errMarshal != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting values for all objects in Cassandra" + isErr.Error())
		} else {
			//response.IsSuccess = true
			//response.GetResponseWithBody(byteValue)
			//response.Message = "Successfully retrieved values for all objects in Cassandra"
			//request.Log(response.Message)

			if len(my) == 0 {
				response.IsSuccess = true
				response.Message = "No objects found in Cassandra"
				var emptyMap map[string]interface{}
				emptyMap = make(map[string]interface{})
				byte, _ := json.Marshal(emptyMap)
				response.GetResponseWithBody(byte)
			} else {
				response.IsSuccess = true
				response.GetResponseWithBody(byteValue)
				response.Message = "Successfully retrieved values for one object in Cassandra"
				request.Log(response.Message)
			}
		}
	}

	return response
}

func (repository CassandraRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Get Search not implemented in Cassandra Db repository")
	return getDefaultNotImplemented()
}

func (repository CassandraRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-QUERY!")
	response := RepositoryResponse{}
	queryType := request.Body.Query.Type

	switch queryType {
	case "Query":
		if request.Body.Query.Parameters != "*" {
			request.Log("Support for SQL Query not implemented in Cassandra Db repository")
			return getDefaultNotImplemented()
		} else {
			return repository.GetAll(request)
		}
	default:
		request.Log(queryType + " not implemented in Cassandra Db repository")
		return getDefaultNotImplemented()
	}

	return response
}

func (repository CassandraRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting Get-BY-KEY!")
	response := RepositoryResponse{}
	session, err := repository.GetConnection(request)
	if err != nil {
		response.GetErrorResponse(err.Error())
	} else {

		//get primary key field name
		iter := session.Query("select type, column_name from system.schema_columns WHERE keyspace_name='" + repository.GetNamespace(request.Controls.Namespace) + "' AND columnfamily_name='" + strings.ToLower(request.Controls.Class) + "'").Iter()

		my1, isErr := iter.SliceMap()

		if isErr != nil {
			response.IsSuccess = true
			response.Message = isErr.Error()
			fmt.Println(isErr.Error())
			response.Message = "No objects found in Cassandra"
			var emptyMap map[string]interface{}
			emptyMap = make(map[string]interface{})
			byte, _ := json.Marshal(emptyMap)
			response.GetResponseWithBody(byte)
			return response
		}

		iter.Close()

		fieldName := ""

		for _, value := range my1 {

			if value["type"].(string) == "partition_key" {
				fieldName = value["column_name"].(string)
				break
			}
		}

		parameter := request.Controls.Id

		iter2 := session.Query("SELECT * FROM " + strings.ToLower(request.Controls.Class) + " where " + fieldName + " = '" + parameter + "'").Iter()

		my, isErr := iter2.SliceMap()

		iter2.Close()

		if request.Controls.SendMetaData == "false" {

			for index, arrVal := range my {
				for key, _ := range arrVal {
					if key == "osheaders" {
						delete(my[index], key)
					}
				}
			}
		}

		byteValue, errMarshal := json.Marshal(my)
		if errMarshal != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting values for one object in Cassandra" + isErr.Error())
		} else {
			if len(my) == 0 {
				response.IsSuccess = true
				response.Message = "No objects found in Cassandra"
				var emptyMap map[string]interface{}
				emptyMap = make(map[string]interface{})
				byte, _ := json.Marshal(emptyMap)
				response.GetResponseWithBody(byte)
			} else {
				response.IsSuccess = true
				response.GetResponseWithBody(byteValue)
				response.Message = "Successfully retrieved values for one object in Cassandra"
				request.Log(response.Message)
			}
		}
	}

	return response
}

func (repository CassandraRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting Insert-Multiple!")
	response := RepositoryResponse{}

	var idData map[string]interface{}
	idData = make(map[string]interface{})

	session, err := repository.GetConnection(request)
	if err != nil {
		response.GetErrorResponse(err.Error())
	} else {

		// if createCassandraTable(request, session) {
		// 	request.Log("Table Verified Successfully!")
		// } else {
		// 	response.IsSuccess = false
		// 	return response
		// }

		var DataObjects []map[string]interface{}
		DataObjects = make([]map[string]interface{}, len(request.Body.Objects))

		//change osheaders
		for i := 0; i < len(request.Body.Objects); i++ {
			var tempMapObject map[string]interface{}
			tempMapObject = make(map[string]interface{})

			for key, value := range request.Body.Objects[i] {
				if key == "__osHeaders" {
					tempMapObject["osheaders"] = value
				} else {
					tempMapObject[strings.ToLower(key)] = value
				}
			}

			DataObjects[i] = tempMapObject
		}

		for i := 0; i < len(DataObjects); i++ {

			keyValue := GetRecordID(request, DataObjects[i])
			DataObjects[i][strings.ToLower(request.Body.Parameters.KeyProperty)] = keyValue
			idData[strconv.Itoa(i)] = keyValue
			if keyValue == "" {
				response.IsSuccess = false
				response.Message = "Failed inserting multiple object in Cassandra"
				request.Log(response.Message)
				request.Log("Inavalid ID request")
				return response
			}

			noOfElements := len(DataObjects[i])

			var keyArray = make([]string, noOfElements)
			var valueArray = make([]string, noOfElements)

			// Process A :start identifying individual data in array and convert to string
			var startIndex int = 0

			for key, value := range DataObjects[i] {

				if key != "__osHeaders" {
					if _, ok := value.(string); ok {
						//Implement all MAP related logic here. All correct data are being caught in here
						keyArray[startIndex] = key
						valueArray[startIndex] = value.(string)
						startIndex = startIndex + 1

					} else {
						request.Log("Not String converting to string")
						keyArray[startIndex] = key
						valueArray[startIndex] = getStringByObject(value)
						startIndex = startIndex + 1
					}
				} else {
					//__osHeaders Catched!
					keyArray[startIndex] = "osHeaders"
					valueArray[startIndex] = ConvertOsheaders(value.(messaging.ControlHeaders))
					startIndex = startIndex + 1
				}
			}

			var argKeyList string
			var argValueList string

			//Build the query string
			for i := 0; i < noOfElements; i++ {
				if i != noOfElements-1 {
					argKeyList = argKeyList + keyArray[i] + ", "
					argValueList = argValueList + "'" + valueArray[i] + "'" + ", "
				} else {
					argKeyList = argKeyList + keyArray[i]
					argValueList = argValueList + "'" + valueArray[i] + "'"
				}
			}

			//DEBUG USE : Display Query information
			//fmt.Println("Table Name : " + request.Controls.Class)
			//fmt.Println("Key list : " + argKeyList)
			//fmt.Println("Value list : " + argValueList)
			//fmt.Println("INSERT INTO " + request.Controls.Class + " (" + argKeyList + ") VALUES (" + argValueList + ")")

			err := session.Query("INSERT INTO " + strings.ToLower(request.Controls.Class) + " (" + argKeyList + ") VALUES (" + argValueList + ")").Exec()
			if err != nil {
				response.IsSuccess = false
				response.Message = "Successfully inserted one object in Cassandra"
				request.Log(response.Message + " : " + err.Error())
			} else {
				response.IsSuccess = true
				response.Message = "Successfully inserted one object in Cassandra"
				request.Log(response.Message)

			}

		}
	}

	//Update Response
	var DataMap []map[string]interface{}
	DataMap = make([]map[string]interface{}, 1)
	var actualInput map[string]interface{}
	actualInput = make(map[string]interface{})
	actualInput["ID"] = idData
	DataMap[0] = actualInput
	response.Data = DataMap
	return response
}

// func (repository CassandraRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
// 	request.Log("Starting INSERT-SINGLE")
// 	response := RepositoryResponse{}
// 	session, err := repository.GetConnection(request)
// 	keyValue := GetRecordID(request, nil)
// 	if err != nil || keyValue == "" {
// 		response.GetErrorResponse(err.Error())
// 	} else {
// 		//change field names to Lower Case
// 		var DataObject map[string]interface{}
// 		DataObject = make(map[string]interface{})

// 		for key, value := range request.Body.Object {
// 			if key == "__osHeaders" {
// 				DataObject["osheaders"] = value
// 			} else {
// 				DataObject[strings.ToLower(key)] = value
// 			}
// 		}

// 		noOfElements := len(DataObject)
// 		DataObject[strings.ToLower(request.Body.Parameters.KeyProperty)] = keyValue

// 		// if createCassandraTable(request, session) {
// 		// 	request.Log("Table Verified Successfully!")
// 		// } else {
// 		// 	response.IsSuccess = false
// 		// 	return response
// 		// }

// 		//indexNames := getCassandraFieldOrder(request)
// 		indexNames := make([]string, 0)
// 		var argKeyList string
// 		var argValueList string

// 		//create keyvalue list

// 		for i := 0; i < len(indexNames); i++ {
// 			if i != len(indexNames)-1 {
// 				argKeyList = argKeyList + indexNames[i] + ", "
// 			} else {
// 				argKeyList = argKeyList + indexNames[i]
// 			}
// 		}

// 		var keyArray = make([]string, noOfElements)
// 		var valueArray = make([]string, noOfElements)

// 		// Process A :start identifying individual data in array and convert to string
// 		for index := 0; index < len(indexNames); index++ {
// 			if indexNames[index] != "osheaders" {

// 				if _, ok := DataObject[indexNames[index]].(string); ok {
// 					keyArray[index] = indexNames[index]
// 					valueArray[index] = DataObject[indexNames[index]].(string)
// 				} else {
// 					fmt.Println("Non string value detected, Will be strigified!")
// 					keyArray[index] = indexNames[index]
// 					valueArray[index] = getStringByObject(DataObject[indexNames[index]])
// 				}
// 			} else {
// 				// __osHeaders Catched!
// 				keyArray[index] = "osheaders"
// 				valueArray[index] = ConvertOsheaders(DataObject[indexNames[index]].(messaging.ControlHeaders))
// 			}

// 		}

// 		//Build the query string
// 		for i := 0; i < noOfElements; i++ {
// 			if i != noOfElements-1 {
// 				argValueList = argValueList + "'" + valueArray[i] + "'" + ", "
// 			} else {
// 				argValueList = argValueList + "'" + valueArray[i] + "'"
// 			}
// 		}
// 		//..........................................

// 		//DEBUG USE : Display Query information
// 		//fmt.Println("Table Name : " + request.Controls.Class)
// 		//fmt.Println("Key list : " + argKeyList)
// 		//fmt.Println("Value list : " + argValueList)
// 		//request.Log("INSERT INTO " + request.Controls.Class + " (" + argKeyList + ") VALUES (" + argValueList + ")")
// 		request.Log("INSERT INTO " + strings.ToLower(request.Controls.Class) + " (" + argKeyList + ") VALUES (" + argValueList + ");")
// 		err := session.Query("INSERT INTO " + strings.ToLower(request.Controls.Class) + " (" + argKeyList + ") VALUES (" + argValueList + ");").Exec()
// 		if err != nil {
// 			response.IsSuccess = false
// 			response.GetErrorResponse("Error inserting one object in Cassandra" + err.Error())
// 			if strings.Contains(err.Error(), "duplicate key value") {
// 				response.IsSuccess = true
// 				response.Message = "No Change since record already Available!"
// 				request.Log(response.Message)
// 				return response
// 			}
// 		} else {
// 			response.IsSuccess = true
// 			response.Message = "Successfully inserted one object in Cassandra"
// 			request.Log(response.Message)
// 		}
// 	}

// 	//Update Response
// 	var Data []map[string]interface{}
// 	Data = make([]map[string]interface{}, 1)
// 	var actualData map[string]interface{}
// 	actualData = make(map[string]interface{})
// 	actualData["ID"] = keyValue
// 	Data[0] = actualData
// 	response.Data = Data
// 	return response
// }

func (repository CassandraRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	var response RepositoryResponse

	conn, err := repository.GetConnection(request)
	if err != nil {
		response.IsSuccess = false
		response.Message = err.Error()
		return response
	}

	_ = conn

	id := GetRecordID(request, request.Body.Object)
	request.Controls.Id = id
	request.Body.Object[request.Body.Parameters.KeyProperty] = id

	Data := make([]map[string]interface{}, 1)
	var idData map[string]interface{}
	idData = make(map[string]interface{})
	idData["ID"] = id
	Data[0] = idData

	//response = repository.queryStore(request)
	// if !response.IsSuccess {
	// 	response = repository.ReRun(request, conn, request.Body.Object)
	// }

	response.Data = Data
	return response
}

func (repository CassandraRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-MULTIPLE")
	response := RepositoryResponse{}
	session, err := repository.GetConnection(request)
	if err != nil {
		response.GetErrorResponse(err.Error())
	} else {

		for i := 0; i < len(request.Body.Objects); i++ {
			noOfElements := len(request.Body.Objects[i]) - 1
			var keyUpdate = make([]string, noOfElements)
			var valueUpdate = make([]string, noOfElements)

			var startIndex = 0
			for key, value := range request.Body.Objects[i] {
				if key != request.Body.Parameters.KeyProperty {
					if key != "__osHeaders" {
						if _, ok := value.(string); ok {
							//Implement all MAP related logic here. All correct data are being caught in here
							keyUpdate[startIndex] = key
							valueUpdate[startIndex] = value.(string)
							startIndex = startIndex + 1

						} else {
							request.Log("Not String converting to string")
							keyUpdate[startIndex] = key
							valueUpdate[startIndex] = getStringByObject(value)
							startIndex = startIndex + 1
						}
					} else {
						keyUpdate[startIndex] = "osHeaders"
						valueUpdate[startIndex] = ConvertOsheaders(value.(messaging.ControlHeaders))
						startIndex = startIndex + 1
					}
				}

			}

			var argValueList string

			//Build the query string
			for i := 0; i < noOfElements; i++ {
				if i != noOfElements-1 {
					argValueList = argValueList + keyUpdate[i] + " = " + "'" + valueUpdate[i] + "'" + ", "
				} else {
					argValueList = argValueList + keyUpdate[i] + " = " + "'" + valueUpdate[i] + "'"
				}
			}

			//DEBUG USE : Display Query information
			//fmt.Println("Table Name : " + request.Controls.Class)
			//fmt.Println("Value list : " + argValueList)
			obj := request.Body.Objects[i]
			err := session.Query("UPDATE " + strings.ToLower(request.Controls.Class) + " SET " + argValueList + " WHERE " + request.Body.Parameters.KeyProperty + " =" + "'" + obj[request.Body.Parameters.KeyProperty].(string) + "'").Exec()
			request.Log("UPDATE " + strings.ToLower(request.Controls.Class) + " SET " + argValueList + " WHERE " + request.Body.Parameters.KeyProperty + " =" + "'" + obj[request.Body.Parameters.KeyProperty].(string) + "'")

			//err := collection.Update(bson.M{key: value}, bson.M{"$set": request.Body.Object})
			if err != nil {
				response.IsSuccess = false
				request.Log("Error updating object in Cassandra  : " + getNoSqlKey(request) + ", " + err.Error())
				response.GetErrorResponse("Error updating one object in Cassandra because no match was found!" + err.Error())
			} else {
				response.IsSuccess = true
				response.Message = "Successfully updating one object in Cassandra "
				request.Log(response.Message)
			}
		}

	}
	return response
}

func (repository CassandraRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-SINGLE")
	response := RepositoryResponse{}
	session, err := repository.GetConnection(request)
	if err != nil {
		response.GetErrorResponse(err.Error())
	} else {

		noOfElements := len(request.Body.Object) - 1
		var keyUpdate = make([]string, noOfElements)
		var valueUpdate = make([]string, noOfElements)

		var startIndex = 0
		for key, value := range request.Body.Object {
			if key != request.Body.Parameters.KeyProperty {
				if key != "__osHeaders" {
					if _, ok := value.(string); ok {
						//Implement all MAP related logic here. All correct data are being caught in here
						keyUpdate[startIndex] = key
						valueUpdate[startIndex] = value.(string)
						startIndex = startIndex + 1

					} else {
						fmt.Println("Not String.. Converting to string before storing")
						keyUpdate[startIndex] = key
						valueUpdate[startIndex] = getStringByObject(value)
						startIndex = startIndex + 1
					}
				} else {
					keyUpdate[startIndex] = "osHeaders"
					valueUpdate[startIndex] = ConvertOsheaders(value.(messaging.ControlHeaders))
					startIndex = startIndex + 1
				}
			}

		}

		var argValueList string

		//Build the query string
		for i := 0; i < noOfElements; i++ {
			if i != noOfElements-1 {
				argValueList = argValueList + keyUpdate[i] + " = " + "'" + valueUpdate[i] + "'" + ", "
			} else {
				argValueList = argValueList + keyUpdate[i] + " = " + "'" + valueUpdate[i] + "'"
			}
		}

		//DEBUG USE : Display Query information
		//fmt.Println("Table Name : " + request.Controls.Class)
		//fmt.Println("Value list : " + argValueList)

		err := session.Query("UPDATE " + strings.ToLower(request.Controls.Class) + " SET " + argValueList + " WHERE " + request.Body.Parameters.KeyProperty + " =" + "'" + getNoSqlKey(request) + "'").Exec()
		//request.Log("UPDATE " + request.Controls.Class + " SET " + argValueList + " WHERE " + request.Body.Parameters.KeyProperty + " =" + "'" + request.Controls.Id + "'")
		if err != nil {
			response.IsSuccess = false
			request.Log("Error updating object in Cassandra  : " + getNoSqlKey(request) + ", " + err.Error())
			response.GetErrorResponse("Error updating one object in Cassandra because no match was found!" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully updating one object in Cassandra "
			request.Log(response.Message)
		}

	}
	return response
}

func (repository CassandraRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-MULTIPLE")
	response := RepositoryResponse{}
	session, err := repository.GetConnection(request)

	if err != nil {
		response.GetErrorResponse(err.Error())
	} else {

		for _, obj := range request.Body.Objects {

			err := session.Query("DELETE FROM " + strings.ToLower(request.Controls.Class) + " WHERE " + request.Body.Parameters.KeyProperty + " = '" + obj[request.Body.Parameters.KeyProperty].(string) + "'").Exec()
			if err != nil {
				response.IsSuccess = false
				request.Log("Error deleting object in Cassandra  : " + err.Error())
				response.GetErrorResponse("Error deleting one object in Cassandra because no match was found!" + err.Error())
			} else {
				response.IsSuccess = true
				response.Message = "Successfully deleted one object in Cassandra"
				request.Log(response.Message)
			}
		}
	}

	return response
}

func (repository CassandraRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-SINGLE")
	response := RepositoryResponse{}
	session, err := repository.GetConnection(request)

	if err != nil {
		response.GetErrorResponse(err.Error())
	} else {

		err := session.Query("DELETE FROM " + strings.ToLower(request.Controls.Class) + " WHERE " + request.Body.Parameters.KeyProperty + " = '" + request.Controls.Id + "'").Exec()
		if err != nil {
			response.IsSuccess = false
			request.Log("Error deleting object in Cassandra  : " + err.Error())
			response.GetErrorResponse("Error deleting one object in Cassandra because no match was found!" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully deleted one object in Cassandra"
			request.Log(response.Message)
		}
	}

	return response
}

func (repository CassandraRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	// request.Log("Starting SPECIAL!")
	// queryType := request.Body.Special.Type

	// switch queryType {
	// case "getFields":
	// 	request.Log("Starting GET-FIELDS sub routine!")
	// 	fieldsInByte := executeCassandraGetFields(request)
	// 	if fieldsInByte != nil {
	// 		response.IsSuccess = true
	// 		response.Message = "Successfully Retrieved Fileds on Class : " + request.Controls.Class
	// 		response.GetResponseWithBody(fieldsInByte)
	// 	} else {
	// 		response.IsSuccess = false
	// 		response.Message = "Aborted! Unsuccessful Retrieving Fileds on Class : " + request.Controls.Class
	// 		err.Error() := response.Message
	// 		response.GetErrorResponse(err.Error())
	// 	}
	// case "getClasses":
	// 	request.Log("Starting GET-CLASSES sub routine")
	// 	fieldsInByte := executeCassandraGetClasses(request)
	// 	if fieldsInByte != nil {
	// 		response.IsSuccess = true
	// 		response.Message = "Successfully Retrieved Fileds on Class : " + request.Controls.Class
	// 		response.GetResponseWithBody(fieldsInByte)
	// 	} else {
	// 		response.IsSuccess = false
	// 		response.Message = "Aborted! Unsuccessful Retrieving Fileds on Class : " + request.Controls.Class
	// 		err.Error() := response.Message
	// 		response.GetErrorResponse(err.Error())
	// 	}
	// case "getNamespaces":
	// 	request.Log("Starting GET-NAMESPACES sub routine")
	// 	fieldsInByte := executeCassandraGetNamespaces(request)
	// 	if fieldsInByte != nil {
	// 		response.IsSuccess = true
	// 		response.Message = "Successfully Retrieved All Namespaces"
	// 		response.GetResponseWithBody(fieldsInByte)
	// 	} else {
	// 		response.IsSuccess = false
	// 		response.Message = "Aborted! Unsuccessful Retrieving All Namespaces"
	// 		err.Error() := response.Message
	// 		response.GetErrorResponse(err.Error())
	// 	}
	// case "getSelected":
	// 	request.Log("Starting GET-SELECTED_FIELDS sub routine")
	// 	fieldsInByte := executeCassandraGetSelectedFields(request)
	// 	if fieldsInByte != nil {
	// 		response.IsSuccess = true
	// 		response.Message = "Successfully Retrieved All selected Field data"
	// 		response.GetResponseWithBody(fieldsInByte)
	// 	} else {
	// 		response.IsSuccess = false
	// 		response.Message = "Aborted! Unsuccessful Retrieving All selected field data"
	// 		err.Error() := response.Message
	// 		response.GetErrorResponse(err.Error())
	// 	}
	// }

	return response
}

func (repository CassandraRepository) Test(request *messaging.ObjectRequest) {

}

//.....................................

// func (repository CassandraRepository) queryStore(request *messaging.ObjectRequest) RepositoryResponse {
// 	response := RepositoryResponse{}

// 	conn, _ := repository.GetConnection(request)

// 	domain := request.Controls.Namespace
// 	class := request.Controls.Class

// 	isOkay := true

// 	if request.Body.Object != nil || len(request.Body.Objects) == 1 {

// 		obj := make(map[string]interface{})

// 		if request.Body.Object != nil {
// 			obj = request.Body.Object
// 		} else {
// 			obj = request.Body.Objects[0]
// 		}

// 		insertScript := repository.GetSingleObjectInsertQuery(request, domain, class, obj, conn)
// 		err, _ := repository.ExecuteNonQuery(conn, insertScript, request)
// 		if err != nil {
// 			if !strings.Contains(err.Error(), "specified twice") {
// 				updateScript := repository.GetSingleObjectUpdateQuery(request, domain, class, obj, conn)
// 				err, message := repository.ExecuteNonQuery(conn, updateScript, request)
// 				if err != nil {
// 					isOkay = false
// 					request.Log("Error : " + err.Error())
// 				} else {
// 					if message == "No Rows Changed" {
// 						request.Log("Information : No Rows Changed for : " + request.Body.Parameters.KeyProperty + " = " + obj[request.Body.Parameters.KeyProperty].(string))
// 					}
// 					isOkay = true
// 				}
// 			} else {
// 				isOkay = false
// 			}
// 		} else {
// 			isOkay = true
// 		}

// 	} else {

// 		//execute insert queries
// 		scripts, err := repository.GetMultipleStoreScripts(conn, request)

// 		for x := 0; x < len(scripts); x++ {
// 			script := scripts[x]["query"].(string)
// 			if err == nil && script != "" {

// 				err, _ := repository.ExecuteNonQuery(conn, script, request)
// 				if err != nil {
// 					request.Log("Error : " + err.Error())
// 					if strings.Contains(err.Error(), "Duplicate entry") {
// 						errorBlock := scripts[x]["queryObject"].([]map[string]interface{})
// 						for _, singleQueryObject := range errorBlock {
// 							insertScript := repository.GetSingleObjectInsertQuery(request, domain, class, singleQueryObject, conn)
// 							err1, _ := repository.ExecuteNonQuery(conn, insertScript, request)
// 							if err1 != nil {
// 								if !strings.Contains(err.Error(), "specified twice") {
// 									updateScript := repository.GetSingleObjectUpdateQuery(request, domain, class, singleQueryObject, conn)
// 									err2, message := repository.ExecuteNonQuery(conn, updateScript, request)
// 									if err2 != nil {
// 										request.Log("Error : " + err2.Error())
// 										isOkay = false
// 									} else {
// 										if message == "No Rows Changed" {
// 											request.Log("Information : No Rows Changed for : " + request.Body.Parameters.KeyProperty + " = " + singleQueryObject[request.Body.Parameters.KeyProperty].(string))
// 										}
// 									}
// 								}
// 							}
// 						}
// 					} else {
// 						//if strings.Contains(err.Error(), "doesn't exist") {
// 						isOkay = false
// 						break
// 						//}
// 					}
// 				}

// 			} else {
// 				isOkay = false
// 				request.Log("Error : " + err.Error())
// 			}
// 		}

// 	}

// 	if isOkay {
// 		response.IsSuccess = true
// 		response.Message = "Successfully stored object(s) in CloudSQL"
// 		request.Log("Debug : " + response.Message)
// 	} else {
// 		response.IsSuccess = false
// 		response.Message = "Error storing/updating all object(s) in CloudSQL."
// 		request.Log("Error : " + response.Message)
// 	}

// 	repository.CloseConnection(conn)
// 	return response
// }

// func (repository CassandraRepository) GetSingleObjectInsertQuery(request *messaging.ObjectRequest, namespace, class string, obj map[string]interface{}, conn *sql.DB) (query string) {
// 	var keyArray []string
// 	query = ""
// 	query = ("INSERT INTO " + repository.GetNamespace(namespace) + "." + class)

// 	id := ""

// 	if obj["OriginalIndex"] == nil {
// 		id = getNoSqlKeyById(request, obj)
// 	} else {
// 		id = obj["OriginalIndex"].(string)
// 	}

// 	delete(obj, "OriginalIndex")

// 	keyList := ""
// 	valueList := ""

// 	for k, _ := range obj {
// 		keyList += ("," + k)
// 		keyArray = append(keyArray, k)
// 	}

// 	for _, k := range keyArray {
// 		v := obj[k]
// 		valueList += ("," + repository.GetSqlFieldValue(v))
// 	}

// 	query += "(__os_id" + keyList + ") VALUES "
// 	query += ("(\"" + id + "\"" + valueList + ")")
// 	return
// }

// func (repository CassandraRepository) GetSingleObjectUpdateQuery(request *messaging.ObjectRequest, namespace, class string, obj map[string]interface{}, conn *sql.DB) (query string) {

// 	updateValues := ""
// 	isFirst := true
// 	for k, v := range obj {
// 		if isFirst {
// 			isFirst = false
// 		} else {
// 			updateValues += ","
// 		}

// 		updateValues += (k + "=" + repository.GetSqlFieldValue(v))
// 	}
// 	query = ("UPDATE " + repository.GetNamespace(namespace) + "." + class + " SET " + updateValues + " WHERE __os_id=\"" + getNoSqlKeyById(request, obj) + "\";")
// 	return
// }

// func (repository CassandraRepository) ExecuteNonQuery(conn *gocql.Session, query string, request *messaging.ObjectRequest) (err error, message string) {
// 	request.Log("Debug Query : " + query)
// 	tokens := strings.Split(strings.ToLower(query), " ")
// 	result, err := conn.Exec(query)
// 	if err == nil {
// 		val, _ := result.RowsAffected()
// 		if val <= 0 && (tokens[0] == "delete" || tokens[0] == "update") {
// 			message = "No Rows Changed"
// 		}
// 	}
// 	return
// }

// func (repository CassandraRepository) GetMultipleStoreScripts(conn *gocql.Session, request *messaging.ObjectRequest) (query []map[string]interface{}, err error) {
// 	namespace := request.Controls.Namespace
// 	class := request.Controls.Class

// 	noOfElementsPerSet := 1000
// 	noOfSets := (len(request.Body.Objects) / noOfElementsPerSet)
// 	remainderFromSets := 0
// 	remainderFromSets = (len(request.Body.Objects) - (noOfSets * noOfElementsPerSet))

// 	startIndex := 0
// 	stopIndex := noOfElementsPerSet

// 	for x := 0; x < noOfSets; x++ {
// 		queryOutput := repository.GetMultipleInsertQuery(request, namespace, class, request.Body.Objects[startIndex:stopIndex], conn)
// 		query = append(query, queryOutput)
// 		startIndex += noOfElementsPerSet
// 		stopIndex += noOfElementsPerSet
// 	}

// 	if remainderFromSets > 0 {
// 		start := len(request.Body.Objects) - remainderFromSets
// 		queryOutput := repository.GetMultipleInsertQuery(request, namespace, class, request.Body.Objects[start:len(request.Body.Objects)], conn)
// 		query = append(query, queryOutput)
// 	}

// 	return
// }

// func (repository CassandraRepository) GetMultipleInsertQuery(request *messaging.ObjectRequest, namespace, class string, records []map[string]interface{}, conn *sql.DB) (queryData map[string]interface{}) {
// 	queryData = make(map[string]interface{})
// 	query := ""
// 	//create insert scripts
// 	isFirstRow := true
// 	var keyArray []string
// 	for _, obj := range records {
// 		if isFirstRow {
// 			query += ("INSERT INTO " + repository.GetNamespace(namespace) + "." + class)
// 		}

// 		id := ""

// 		if obj["OriginalIndex"] == nil {
// 			id = getNoSqlKeyById(request, obj)
// 		} else {
// 			id = obj["OriginalIndex"].(string)
// 		}

// 		delete(obj, "OriginalIndex")

// 		keyList := ""
// 		valueList := ""

// 		if isFirstRow {
// 			for k, _ := range obj {
// 				keyList += ("," + k)
// 				keyArray = append(keyArray, k)
// 			}
// 		}
// 		//request.Log(keyArray)
// 		for _, k := range keyArray {
// 			v := obj[k]
// 			valueList += ("," + repository.GetSqlFieldValue(v))
// 		}

// 		if isFirstRow {
// 			query += "(__os_id" + keyList + ") VALUES "
// 		} else {
// 			query += ","
// 		}
// 		query += ("(\"" + id + "\"" + valueList + ")")

// 		if isFirstRow {
// 			isFirstRow = false
// 		}
// 	}

// 	queryData["query"] = query
// 	queryData["queryObject"] = records
// 	return
// }

// func (repository CassandraRepository) GetSqlFieldValue(value interface{}) string {
// 	var strValue string
// 	switch v := value.(type) {
// 	case bool:
// 		if value.(bool) == true {
// 			strValue = "b'1'"
// 		} else {
// 			strValue = "b'0'"
// 		}
// 		break
// 	case string:
// 		sval := fmt.Sprint(value)
// 		// if strings.ContainsAny(sval, "\"'\n\r\t") {
// 		if strings.ContainsAny(sval, "\"\n\r\t") {
// 			sEnc := base64.StdEncoding.EncodeToString([]byte(sval))
// 			strValue = "'^" + sEnc + "'"
// 		} else {
// 			strValue = "'" + sval + "'"
// 		}
// 		/*else if (strings.Contains(sval, "'")){
// 		  		    sEnc := base64.StdEncoding.EncodeToString([]byte(sval))
// 		      		strValue = "'^" + sEnc + "'";
// 		  		}
// 		break
// 	default:
// 		strValue = "'" + repository.GetJson(v) + "'"
// 		break

// 	}

// 	return strValue
// }

// func (repository CassandraRepository) CloseConnection(conn *gocql.Session) {
// 	// err := conn.Close()
// 	// if err != nil {
// 	// 	request.Log(err.Error())
// 	// } else {
// 	// 	request.Log("Connection Closed!")
// 	// }
// }

//SUB ROUTINES

// func executeCassandraGetFields(request *messaging.ObjectRequest) (returnByte []byte) {
// 	session, isError, _ := repository.GetConnection(request)
// 	if isError == true {
// 		request.Log("Cassandra connection failed")
// 		returnByte = nil
// 	} else {
// 		isError = false

// 		iter2 := session.Query("select column_name from system.schema_columns WHERE keyspace_name='" + repository.GetNamespace(request.Controls.Namespace) + "' AND columnfamily_name='" + strings.ToLower(request.Controls.Class) + "'").Iter()

// 		my, _ := iter2.SliceMap()

// 		iter2.Close()

// 		var fields []string
// 		fields = make([]string, len(my))

// 		for key, value := range my {
// 			for _, fieldname := range value {
// 				fields[key] = fieldname.(string)
// 			}
// 		}

// 		byteValue, errMarshal := json.Marshal(fields)
// 		if errMarshal != nil {
// 			request.Log("Error getting values for all objects in Cassandra")
// 			returnByte = []byte("Error JSON marshalling to BYTE array")
// 		} else {
// 			request.Log("Successfully retrieved values for all objects in Cassandra")
// 			returnByte = byteValue
// 		}

// 	}
// 	return
// }

// func executeCassandraGetClasses(request *messaging.ObjectRequest) (returnByte []byte) {
// 	session, isError, _ := repository.GetConnection(request)
// 	if isError == true {
// 		request.Log("Cassandra connection failed")
// 		returnByte = nil
// 	} else {
// 		isError = false

// 		iter2 := session.Query("select columnfamily_name from system.schema_columnfamilies WHERE keyspace_name='" + getSQLnamespace(request) + "'").Iter()

// 		my, _ := iter2.SliceMap()

// 		iter2.Close()

// 		var fields []string
// 		fields = make([]string, len(my))

// 		for key, value := range my {
// 			for _, fieldname := range value {
// 				fields[key] = fieldname.(string)
// 			}
// 		}

// 		byteValue, errMarshal := json.Marshal(fields)
// 		if errMarshal != nil {
// 			request.Log("Error getting values for all objects in Cassandra")
// 			returnByte = []byte("Error JSON marshalling to BYTE array")
// 		} else {
// 			request.Log("Successfully retrieved values for all objects in Cassandra")
// 			returnByte = byteValue
// 		}

// 	}
// 	return
// }

// func executeCassandraGetNamespaces(request *messaging.ObjectRequest) (returnByte []byte) {
// 	session, isError, _ := repository.GetConnection(request)
// 	if isError == true {
// 		request.Log("Cassandra connection failed")
// 		returnByte = nil
// 	} else {
// 		isError = false

// 		iter2 := session.Query("select keyspace_name from system.schema_keyspaces").Iter()

// 		my, _ := iter2.SliceMap()

// 		iter2.Close()

// 		var fields []string
// 		fields = make([]string, len(my))

// 		for key, value := range my {
// 			for _, fieldname := range value {
// 				fields[key] = fieldname.(string)
// 			}
// 		}

// 		byteValue, errMarshal := json.Marshal(fields)
// 		if errMarshal != nil {
// 			request.Log("Error getting values for all objects in Cassandra")
// 			returnByte = []byte("Error JSON marshalling to BYTE array")
// 		} else {
// 			request.Log("Successfully retrieved values for all objects in Cassandra")
// 			returnByte = byteValue
// 		}

// 	}
// 	return
// }

// func executeCassandraGetSelectedFields(request *messaging.ObjectRequest) (returnByte []byte) {
// 	session, isError, _ := repository.GetConnection(request)
// 	if isError == true {
// 		request.Log("Cassandra connection failed")
// 		returnByte = nil
// 	} else {
// 		isError = false

// 		var selectedItemsQuery string

// 		var requestedFields []string
// 		request.Log("Requested Field List : " + request.Body.Special.Parameters)
// 		if request.Body.Special.Parameters == "*" {
// 			request.Log("All fields requested")
// 			requestedFields = make([]string, 1)
// 			requestedFields[0] = "*"
// 			selectedItemsQuery = "*"
// 		} else {
// 			requestedFields = strings.Split(request.Body.Special.Parameters, " ")

// 			for key, value := range requestedFields {
// 				if key == len(requestedFields)-1 {
// 					selectedItemsQuery += value
// 				} else {
// 					selectedItemsQuery += (value + ",")
// 				}
// 			}
// 		}

// 		iter2 := session.Query("select " + selectedItemsQuery + " from " + strings.ToLower(request.Controls.Class)).Iter()

// 		my, _ := iter2.SliceMap()

// 		iter2.Close()

// 		byteValue, errMarshal := json.Marshal(my)
// 		if errMarshal != nil {
// 			request.Log("Error getting values for all objects in Cassandra")
// 			returnByte = []byte("Error JSON marshalling to BYTE array")
// 		} else {
// 			request.Log("Successfully retrieved values for all objects in Cassandra")
// 			returnByte = byteValue
// 		}
// 	}

// 	return
// }
// func getCassandraDataType(item interface{}) (datatype string) {
// 	datatype = reflect.TypeOf(item).Name()
// 	if datatype == "bool" {
// 		datatype = "text"
// 	} else if datatype == "float64" {
// 		datatype = "text"
// 	} else if datatype == "string" {
// 		datatype = "text"
// 	} else if datatype == "" || datatype == "ControlHeaders" {
// 		datatype = "text"
// 	}
// 	return datatype
// }

// func getCassandraFieldOrder(request *messaging.ObjectRequest) []string {
// 	var returnArray []string
// 	//read fields
// 	byteValue := executeCassandraGetFields(request)

// 	err := json.Unmarshal(byteValue, &returnArray)
// 	fmt.Print("Field List from DB : ")
// 	fmt.Println(returnArray)
// 	if err != nil {
// 		request.Log("Converstion of Json Failed!")
// 		returnArray = make([]string, 1)
// 		returnArray[0] = "nil"
// 		return returnArray
// 	}

// 	return returnArray
// }

// func createCassandraTable(request *messaging.ObjectRequest, session *gocql.Session) (status bool) {
// 	status = false

// 	//get table list
// 	classBytes := executeCassandraGetClasses(request)
// 	var classList []string
// 	err := json.Unmarshal(classBytes, &classList)
// 	fmt.Print("Recieved Table List : ")
// 	fmt.Println(classList)
// 	if err != nil {
// 		status = false
// 	} else {
// 		for _, className := range classList {
// 			if strings.ToLower(request.Controls.Class) == className {
// 				fmt.Println("Table Already Available")
// 				status = true
// 				//Get all fields
// 				classBytes := executeCassandraGetFields(request)
// 				var tableFieldList []string
// 				_ = json.Unmarshal(classBytes, &tableFieldList)
// 				//Check For missing fields. If any ALTER TABLE
// 				var recordFieldList []string
// 				var recordFieldType []string
// 				if request.Body.Object == nil {
// 					recordFieldList = make([]string, len(request.Body.Objects[0]))
// 					recordFieldType = make([]string, len(request.Body.Objects[0]))
// 					index := 0
// 					for key, value := range request.Body.Objects[0] {
// 						if key == "__osHeaders" {
// 							recordFieldList[index] = "osheaders"
// 							recordFieldType[index] = "text"
// 						} else {
// 							recordFieldList[index] = strings.ToLower(key)
// 							recordFieldType[index] = getCassandraDataType(value)
// 						}
// 						index++
// 					}
// 				} else {
// 					recordFieldList = make([]string, len(request.Body.Object))
// 					recordFieldType = make([]string, len(request.Body.Object))
// 					index := 0
// 					for key, value := range request.Body.Object {
// 						if key == "__osHeaders" {
// 							recordFieldList[index] = "osheaders"
// 							recordFieldType[index] = "text"
// 						} else {
// 							recordFieldList[index] = strings.ToLower(key)
// 							recordFieldType[index] = getCassandraDataType(value)
// 						}
// 						index++
// 					}
// 				}

// 				var newFields []string
// 				var newTypes []string

// 				//check for new Fields
// 				for key, fieldName := range recordFieldList {
// 					isAvailable := false
// 					for _, tableField := range tableFieldList {
// 						if fieldName == tableField {
// 							isAvailable = true
// 							break
// 						}
// 					}

// 					if !isAvailable {
// 						newFields = append(newFields, fieldName)
// 						newTypes = append(newTypes, recordFieldType[key])
// 					}
// 				}

// 				//ALTER TABLES

// 				for key, _ := range newFields {
// 					request.Log("ALTER TABLE " + strings.ToLower(request.Controls.Class) + " ADD " + newFields[key] + " " + newTypes[key] + ";")
// 					er := session.Query("ALTER TABLE " + strings.ToLower(request.Controls.Class) + " ADD " + newFields[key] + " " + newTypes[key] + ";").Exec()
// 					if er != nil {
// 						status = false
// 						request.Log("Table Alter Failed : " + er.Error())
// 						return
// 					} else {
// 						status = true
// 						request.Log("Table Alter Success!")
// 					}
// 				}

// 				return
// 			}
// 		}

// 		// if not available
// 		//get one object
// 		var dataObject map[string]interface{}
// 		dataObject = make(map[string]interface{})

// 		if request.Body.Object != nil {
// 			for key, value := range request.Body.Object {
// 				if key == "__osHeaders" {
// 					dataObject["osheaders"] = value
// 				} else {
// 					dataObject[key] = value
// 				}
// 			}
// 		} else {
// 			for key, value := range request.Body.Objects[0] {
// 				if key == "__osHeaders" {
// 					dataObject["osheaders"] = value
// 				} else {
// 					dataObject[key] = value
// 				}
// 			}
// 		}
// 		//read fields
// 		noOfElements := len(dataObject)
// 		var keyArray = make([]string, noOfElements)
// 		var dataTypeArray = make([]string, noOfElements)

// 		var startIndex int = 0

// 		for key, value := range dataObject {
// 			keyArray[startIndex] = key
// 			dataTypeArray[startIndex] = getCassandraDataType(value)
// 			startIndex = startIndex + 1

// 		}

// 		//Create Table

// 		var argKeyList2 string

// 		for i := 0; i < noOfElements; i++ {
// 			if i != noOfElements-1 {
// 				if keyArray[i] == request.Body.Parameters.KeyProperty {
// 					argKeyList2 = argKeyList2 + keyArray[i] + " text PRIMARY KEY, "
// 				} else {
// 					argKeyList2 = argKeyList2 + keyArray[i] + " " + dataTypeArray[i] + ", "
// 				}

// 			} else {
// 				if keyArray[i] == request.Body.Parameters.KeyProperty {
// 					argKeyList2 = argKeyList2 + keyArray[i] + " text PRIMARY KEY"
// 				} else {
// 					argKeyList2 = argKeyList2 + keyArray[i] + " " + dataTypeArray[i]
// 				}

// 			}
// 		}

// 		request.Log("create table " + strings.ToLower(request.Controls.Class) + " (" + argKeyList2 + ");")

// 		er := session.Query("create table " + strings.ToLower(request.Controls.Class) + " (" + argKeyList2 + ");").Exec()
// 		if er != nil {
// 			status = false
// 			request.Log("Table Creation Failed : " + er.Error())
// 			return
// 		}

// 		status = true

// 	}

// 	return
// }

func (repository CassandraRepository) ClearCache(request *messaging.ObjectRequest) {
}
