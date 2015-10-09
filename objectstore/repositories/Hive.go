package repositories

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/queryparser"
	"encoding/json"
	"fmt"
	"github.com/mattbaird/hive"
	"github.com/twinj/uuid"
	"strconv"
	"strings"
)

type HiveRepository struct {
}

func (repository HiveRepository) GetRepositoryName() string {
	return "Hive"
}

func getHiveConnection(request *messaging.ObjectRequest) (conn *hive.HiveConnection, isError bool, errorMessage string) {
	isError = false
	hive.MakePool(request.Configuration.ServerConfiguration["HIVE"]["Host"] + ":" + request.Configuration.ServerConfiguration["HIVE"]["Port"])
	fmt.Println("Hive Server : " + request.Configuration.ServerConfiguration["HIVE"]["Host"] + ":" + request.Configuration.ServerConfiguration["HIVE"]["Port"])
	conn, err := hive.GetHiveConn()
	if err != nil {
		isError = true
		errorMessage = err.Error()
		request.Log("HIVE connection initilizing failed!")
	} else {
		request.Log("HIVE connection initilizing Successful!")
	}

	request.Log("Reusing existing HIVE connection")
	return
}

func (repository HiveRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-ALL")
	response := RepositoryResponse{}
	conn, isError, errorMessage := getHiveConnection(request)
	if isError == false {
		tableName := getSQLnamespace(request) + "." + request.Controls.Class

		er, err := conn.Client.Execute("SELECT * FROM " + tableName)
		if er == nil && err == nil {

			//Get Schema
			schema, _, _ := conn.Client.GetSchema()

			var allMaps []map[string]string

			for {
				row, _, _ := conn.Client.FetchOne()
				if row == "" {
					break
				} else {

					var myMap map[string]string
					myMap = make(map[string]string)

					temp := strings.Split(row, "\t")

					for key, _ := range temp {
						myMap[(schema.FieldSchemas[key].Name)] = temp[key]
					}

					delete(myMap, request.Controls.Class+".primarykey")
					if request.Controls.SendMetaData == "false" {
						delete(myMap, request.Controls.Class+".osheaders")
					}
					allMaps = append(allMaps, myMap)

				}
			}

			if len(allMaps) == 0 {
				response.IsSuccess = true
				response.Message = "No objects found in Hive"
				var emptyMap map[string]interface{}
				emptyMap = make(map[string]interface{})
				byte, _ := json.Marshal(emptyMap)
				response.GetResponseWithBody(byte)
			}

			byteValue, errMarshal := json.Marshal(allMaps)

			if errMarshal != nil {
				response.Message = "Conversion to JSON failed!"
				request.Log(response.Message)
			}
			response.IsSuccess = true
			response.GetResponseWithBody(byteValue)
			response.Message = "Successfully retrieved values for all objects in HIVE"
			request.Log(response.Message)
		} else {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting values for all objects in HIVE" + err.Error())
			response.GetErrorResponse(errorMessage)
		}
	} else {
		response.GetErrorResponse(errorMessage)
	}
	if conn != nil {
		conn.Checkin()
	}
	return response
}

func (repository HiveRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Get Search not implemented in Hive Db repository")
	return getDefaultNotImplemented()
}

func (repository HiveRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-QUERY")
	response := RepositoryResponse{}
	queryType := request.Body.Query.Type

	switch queryType {
	case "Query":
		if request.Body.Query.Parameters != "*" {
			fieldsInByte := executeHiveQuery(request)
			if fieldsInByte != nil {
				response.IsSuccess = true
				response.Message = "Successfully Retrieved Data For Custom Query"
				response.GetResponseWithBody(fieldsInByte)
			} else {
				response.IsSuccess = false
				response.Message = "Aborted! Unsuccessful Retrieving Data For Custom Query"
				errorMessage := response.Message
				response.GetErrorResponse(errorMessage)
			}
		} else {
			return repository.GetAll(request)
		}
	default:
		request.Log(queryType + " not implemented in Hive Db repository")
		return getDefaultNotImplemented()

	}

	return response
}

func executeHiveQuery(request *messaging.ObjectRequest) (returnByte []byte) {

	if checkIfTenantIsAllowed(request.Body.Query.Parameters, request.Controls.Namespace) {

		request.Log("This Tenent is ALLOWED to perform this Query!")

		request.Log("USER INPUT QUERY : " + request.Body.Query.Parameters)
		formattedQuery := queryparser.GetFormattedQuery(request.Body.Query.Parameters)
		request.Log("HIVE QUERY : " + formattedQuery)

		conn, isError, _ := getHiveConnection(request)
		if isError == false {

			request.Log("Executing Query...")
			er, err := conn.Client.Execute(formattedQuery)
			if er == nil && err == nil {

				//Get Schema
				schema, _, _ := conn.Client.GetSchema()

				var allMaps map[string]interface{}
				allMaps = make(map[string]interface{})

				recordIndex := 0

				for {
					row, _, _ := conn.Client.FetchOne()
					if row == "" {
						break
					} else {
						var myMap map[string]string
						myMap = make(map[string]string)

						temp := strings.Split(row, "\t")

						for key, _ := range temp {
							myMap[(schema.FieldSchemas[key].Name)] = temp[key]
						}
						allMaps[strconv.Itoa(recordIndex)] = myMap
						recordIndex++

					}

				}

				byteValue, errMarshal := json.Marshal(allMaps)

				if errMarshal != nil {
					byteValue = nil
				}

				returnByte = byteValue
			}
		}
	} else {
		return ([]byte("This Tenent is NOT ALLOWED to perform submitted Query!"))
	}

	return
}

func (repository HiveRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-BY-KEY")
	response := RepositoryResponse{}
	conn, isError, errorMessage := getHiveConnection(request)
	if isError == false {

		tableName := getSQLnamespace(request) + "." + request.Controls.Class
		key := request.Controls.Id
		fmt.Println("SELECT * FROM " + tableName + " where primarykey='" + key + "'")
		er, err := conn.Client.Execute("SELECT * FROM " + tableName + " where PrimaryKey='" + key + "'")
		if er == nil && err == nil {

			//Get Schema
			schema, _, _ := conn.Client.GetSchema()

			var myMap map[string]string
			myMap = make(map[string]string)

			for {
				row, _, _ := conn.Client.FetchOne()
				if row == "" {
					break
				} else {

					temp := strings.Split(row, "\t")

					for key, _ := range temp {
						myMap[(schema.FieldSchemas[key].Name)] = temp[key]
					}

					delete(myMap, request.Controls.Class+".primarykey")

					if request.Controls.SendMetaData == "false" {
						delete(myMap, request.Controls.Class+".osheaders")
					}
				}
			}

			if len(myMap) == 0 {
				response.IsSuccess = true
				response.Message = "No objects found in Hive"
				var emptyMap map[string]interface{}
				emptyMap = make(map[string]interface{})
				byte, _ := json.Marshal(emptyMap)
				response.GetResponseWithBody(byte)
			}

			byteValue, errMarshal := json.Marshal(myMap)

			if errMarshal != nil {
				response.Message = "Conversion to JSON failed!"
				request.Log(response.Message)
			}
			response.IsSuccess = true
			response.GetResponseWithBody(byteValue)
			response.Message = "Successfully retrieved values for all objects in HIVE"
			request.Log(response.Message)
		} else {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting values for all objects in HIVE" + err.Error())
		}
	} else {
		response.GetErrorResponse(errorMessage)
	}
	if conn != nil {
		conn.Checkin()
	}
	return response
}

func (repository HiveRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-MULTIPLE")
	response := RepositoryResponse{}
	conn, isError, errorMessage := getHiveConnection(request)

	var idData map[string]interface{}
	idData = make(map[string]interface{})
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		tableName := getSQLnamespace(request) + "." + request.Controls.Class

		isDBReady := checkDBAndCreateOnNotAvailbale(getSQLnamespace(request), conn)

		if isDBReady {
			request.Log("Database Ready! Clear to proceed with TSQL transactions...")
		} else {
			request.Log("Database Not Ready! Check Hadoop Configuration. Following TSQL transaction WILL FAIL...")
		}

		isTableChecked := false

		for i := 0; i < len(request.Body.Objects); i++ {

			keyValue := getHiveRecordID(request, request.Body.Objects[i])
			request.Body.Objects[i][request.Body.Parameters.KeyProperty] = keyValue
			idData[strconv.Itoa(i)] = keyValue
			if keyValue == "" {
				response.IsSuccess = false
				response.Message = "Failed inserting multiple object in Cassandra"
				request.Log(response.Message)
				request.Log("Inavalid ID request")
				return response
			}

			//Get method count
			noOfElements := len(request.Body.Objects[i]) + 1

			var keyArray = make([]string, noOfElements)
			var valueArray = make([]string, noOfElements)

			var recordID string
			var keyProperty string
			recordID = keyValue
			keyProperty = request.Body.Parameters.KeyProperty

			request.Body.Objects[i]["primarykey"] = keyValue

			//duplicate fields and then lowercase field names

			var tempMapObject map[string]interface{}
			tempMapObject = make(map[string]interface{})

			for key, value := range request.Body.Objects[i] {
				if key == "__osHeaders" {
					tempMapObject["osheaders"] = value
				} else {
					tempMapObject[strings.ToLower(key)] = value
				}
			}

			request.Body.Objects[i] = tempMapObject

			indexNames := getHiveFieldOrder(request)

			if indexNames[0] != "nil" {
				request.Log("Old class, Prior format is available!")

				for index := 0; index < len(indexNames); index++ {
					if indexNames[index] != "osheaders" {

						if _, ok := request.Body.Objects[i][indexNames[index]].(string); ok {
							keyArray[index] = indexNames[index]
							valueArray[index] = request.Body.Objects[i][indexNames[index]].(string)
						} else {
							request.Log("Non string value detected, Will be strigified!")
							keyArray[index] = indexNames[index]
							valueArray[index] = getStringByObject(request.Body.Objects[i][indexNames[index]])
						}
					} else {
						// __osHeaders Catched!
						keyArray[index] = "osheaders"
						valueArray[index] = ConvertOsheaders(request.Body.Objects[i][indexNames[index]].(messaging.ControlHeaders))
					}

				}

			} else {
				request.Log("New class. No prior format defined.")

				// read all the fields
				var startIndex int = 0

				for key, value := range request.Body.Objects[i] {

					if key != "__osheaders" {

						if _, ok := value.(string); ok {
							keyArray[startIndex] = key
							valueArray[startIndex] = value.(string)
							startIndex = startIndex + 1
						} else {
							request.Log("Non string value detected, Will be strigified!")
							keyArray[startIndex] = key
							valueArray[startIndex] = getStringByObject(value)
							startIndex = startIndex + 1
						}
					} else {
						// __osHeaders Catched!
						keyArray[startIndex] = "osHeaders"
						valueArray[startIndex] = ConvertOsheaders(value.(messaging.ControlHeaders))
						startIndex = startIndex + 1
					}
				}
			}

			//check if data is already available.. IF so send to UPDATE

			isAvailable := checkIfRecordAvailable(recordID, keyProperty, tableName, conn)

			if isAvailable {
				request.Log("RECORD ALREADY AVAILABLE!")
				//response = repository.UpdateMultiple(request)
			} else {
				request.Log("NEW RECORD! Starting INSERT process!")

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

				if !isTableChecked {
					//check if table exists if not create new one

					if checkTableAvailability(request) == false {
						request.Log("There is no Table. Creating New Table..")
						var argKeyList2 string
						clusteredBy := ""
						for i := 0; i < noOfElements; i++ {
							if i != noOfElements-1 {
								argKeyList2 = argKeyList2 + keyArray[i] + " string, "
							} else {
								argKeyList2 = argKeyList2 + keyArray[i] + " string"
								clusteredBy = keyArray[i]
							}
						}

						request.Log("Argument List for New Table : " + argKeyList2)
						er, err := conn.Client.Execute("create table " + tableName + " (" + argKeyList2 + ") clustered by (" + clusteredBy + ") into 1 buckets stored as orc TBLPROPERTIES ('transactional'='true')")
						if er != nil || err != nil {
							request.Log("Table Creation Error! Check Hadoop Hive Configuration")
						}

					}
				}

				isTableChecked = true
				//DEBUG USE : Display Query information
				//fmt.Println("Table Name : " + request.Controls.Class)
				//fmt.Println("Key list : " + argKeyList)
				//fmt.Println("Value list : " + argValueList)
				fmt.Println("insert into table " + tableName + " values (" + argValueList + ")")
				er, err := conn.Client.Execute("insert into table " + tableName + " values (" + argValueList + ")")
				if er == nil && err == nil {
					response.IsSuccess = true
					response.Message = "Successfully inserted a single object in to HIVE"
					request.Log(response.Message)
				} else {
					response.IsSuccess = false
					response.GetErrorResponse("Error inserting a single object in to HIVE" + err.Error())
				}
			}

		}

	}

	if conn != nil {
		conn.Checkin()
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

func (repository HiveRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-SINGLE")
	keyValue := getHiveRecordID(request, nil)
	response := RepositoryResponse{}
	conn, isError, errorMessage := getHiveConnection(request)

	if isError == false && keyValue != "" {
		request.Body.Object[request.Body.Parameters.KeyProperty] = keyValue

		tableName := getSQLnamespace(request) + "." + request.Controls.Class

		isDBReady := checkDBAndCreateOnNotAvailbale(getSQLnamespace(request), conn)

		if isDBReady {
			request.Log("Database Ready! Clear to proceed with TSQL transactions...")
		} else {
			request.Log("Database Not Ready! Check Hadoop Configuration. Following TSQL transaction WILL FAIL...")
		}

		var recordID string
		var keyProperty string
		recordID = keyValue
		keyProperty = request.Body.Parameters.KeyProperty

		noOfElements := len(request.Body.Object) + 1
		var keyArray = make([]string, noOfElements)
		var valueArray = make([]string, noOfElements)

		request.Body.Object["primarykey"] = keyValue

		//duplicate fields and then lowercase field names

		var tempMapObject map[string]interface{}
		tempMapObject = make(map[string]interface{})

		for key, value := range request.Body.Object {
			if key == "__osHeaders" {
				tempMapObject["osheaders"] = value
			} else {
				tempMapObject[strings.ToLower(key)] = value
			}
		}

		request.Body.Object = tempMapObject

		indexNames := getHiveFieldOrder(request)

		if indexNames[0] != "nil" {
			request.Log("Old class, Prior format is available!")

			for index := 0; index < len(indexNames); index++ {
				if indexNames[index] != "osheaders" {

					if _, ok := request.Body.Object[indexNames[index]].(string); ok {
						keyArray[index] = indexNames[index]
						valueArray[index] = request.Body.Object[indexNames[index]].(string)
					} else {
						request.Log("Non string value detected, Will be strigified!")
						keyArray[index] = indexNames[index]
						valueArray[index] = getStringByObject(request.Body.Object[indexNames[index]])
					}
				} else {
					// __osHeaders Catched!
					keyArray[index] = "osheaders"
					valueArray[index] = ConvertOsheaders(request.Body.Object[indexNames[index]].(messaging.ControlHeaders))
				}

			}

		} else {
			request.Log("New class. No prior format defined.")

			// read all the fields
			var startIndex int = 0

			for key, value := range request.Body.Object {

				if key != "__osheaders" {

					if _, ok := value.(string); ok {
						keyArray[startIndex] = key
						valueArray[startIndex] = value.(string)
						startIndex = startIndex + 1
					} else {
						request.Log("Non string value detected, Will be strigified!")
						keyArray[startIndex] = key
						valueArray[startIndex] = getStringByObject(value)
						startIndex = startIndex + 1
					}
				} else {
					// __osHeaders Catched!
					keyArray[startIndex] = "osHeaders"
					valueArray[startIndex] = ConvertOsheaders(value.(messaging.ControlHeaders))
					startIndex = startIndex + 1
				}
			}
		}

		//check if data is already available.. IF so send to UPDATE

		isAvailable := checkIfRecordAvailable(recordID, keyProperty, tableName, conn)

		if isAvailable {
			request.Log("RECORD ALREADY AVAILABLE!")
			//response = repository.UpdateSingle(request)
		} else {
			request.Log("NEW RECORD! Staring INSERT process!")
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

			//check if table exists if not create new one

			if checkTableAvailability(request) == false {
				request.Log("There is no Table. Creating New Table..")
				var argKeyList2 string
				clusteredBy := ""
				for i := 0; i < noOfElements; i++ {
					if i != noOfElements-1 {
						argKeyList2 = argKeyList2 + keyArray[i] + " string, "
					} else {
						argKeyList2 = argKeyList2 + keyArray[i] + " string"
						clusteredBy = keyArray[i]
					}
				}

				//DEBUG USE : Display Query Information
				//fmt.Println("Argument List for New Table : " + argKeyList2)
				//fmt.Println("Executing Query : ")
				fmt.Println("create table " + request.Controls.Class + " (" + argKeyList2 + ") clustered by (" + clusteredBy + ") into 1 buckets stored as orc TBLPROPERTIES ('transactional'='true')")
				er, err := conn.Client.Execute("create table " + tableName + " (" + argKeyList2 + ") clustered by (" + clusteredBy + ") into 1 buckets stored as orc TBLPROPERTIES ('transactional'='true')")
				if er != nil || err != nil {
					request.Log("Table Creation Error! Check Hadoop Hive Configuration")
				}

			}

			//DEBUG USE : Display Query information
			//fmt.Println("Table Name : " + request.Controls.Class)
			//fmt.Println("Key list : " + argKeyList)
			//fmt.Println("Value list : " + argValueList)
			//fmt.Println("Inserting Query : ")
			//fmt.Println("insert into table " + request.Controls.Class + " values (" + argValueList + ")")
			fmt.Println("insert into table " + tableName + " values (" + argValueList + ")")
			_, err := conn.Client.Execute("insert into table " + tableName + " values (" + argValueList + ")")
			if err == nil {
				response.IsSuccess = true
				response.Message = "Successfully inserted a single object in to HIVE"
				request.Log(response.Message)
			} else {
				response.IsSuccess = false
				response.GetErrorResponse("Error inserting a single object in to HIVE" + err.Error())
			}
		}
	} else {
		response.GetErrorResponse(errorMessage)
	}
	if conn != nil {
		conn.Checkin()
	}

	var Data []map[string]interface{}
	Data = make([]map[string]interface{}, 1)
	var actualData map[string]interface{}
	actualData = make(map[string]interface{})
	actualData["ID"] = keyValue
	Data[0] = actualData
	response.Data = Data

	return response
}

func (repository HiveRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-MULTIPLE")
	response := RepositoryResponse{}
	conn, isError, errorMessage := getHiveConnection(request)
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		for i := 0; i < len(request.Body.Objects); i++ {

			tableName := getSQLnamespace(request) + "." + request.Controls.Class

			noOfElements := len(request.Body.Objects[i])
			var keyUpdate = make([]string, noOfElements)
			var valueUpdate = make([]string, noOfElements)
			fmt.Println(noOfElements)
			var startIndex = 0
			for key, value := range request.Body.Objects[i] {

				if key != "__osHeaders" {

					//if str, ok := value.(string); ok {
					if _, ok := value.(string); ok {
						//Implement all MAP related logic here. All correct data are being caught in here
						keyUpdate[startIndex] = key
						valueUpdate[startIndex] = value.(string)
						startIndex = startIndex + 1

					} else {
						request.Log("Non string value detected, Will be strigified!")
						keyUpdate[startIndex] = key
						valueUpdate[startIndex] = getStringByObject(value)
						startIndex = startIndex + 1
					}
				} else {
					// __osHeaders Catched!
					keyUpdate[startIndex] = "osHeaders"
					valueUpdate[startIndex] = ConvertOsheaders(value.(messaging.ControlHeaders))
					startIndex = startIndex + 1
				}
			}

			//Convert them to LOWERCASE

			primaryKeyWord := strings.ToLower(request.Body.Parameters.KeyProperty)

			for key, value := range keyUpdate {
				keyUpdate[key] = strings.ToLower(value)
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

			er, err := conn.Client.Execute("UPDATE " + tableName + " SET " + argValueList + " WHERE " + primaryKeyWord + " =" + "'" + request.Body.Objects[i][request.Body.Parameters.KeyProperty].(string) + "'")
			if er == nil && err == nil {
				response.IsSuccess = true
				response.Message = "Successfully Updated a single object in to HIVE"
				request.Log(response.Message)
			} else {
				response.IsSuccess = false
				response.GetErrorResponse("Error deleting a single object in to HIVE" + err.Error())
			}
		}

		if conn != nil {
			conn.Checkin()
		}

	}
	return response
}

func (repository HiveRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-SINGLE")
	response := RepositoryResponse{}
	conn, isError, errorMessage := getHiveConnection(request)
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		tableName := getSQLnamespace(request) + "." + request.Controls.Class

		noOfElements := len(request.Body.Object)
		var keyUpdate = make([]string, noOfElements)
		var valueUpdate = make([]string, noOfElements)

		var startIndex = 0
		for key, value := range request.Body.Object {

			if key != "__osHeaders" {

				if _, ok := value.(string); ok {
					//Implement all MAP related logic here. All correct data are being caught in here
					keyUpdate[startIndex] = key
					valueUpdate[startIndex] = value.(string)
					startIndex = startIndex + 1

				} else {
					request.Log("Non string value detected, Will be strigified!")
					keyUpdate[startIndex] = key
					valueUpdate[startIndex] = getStringByObject(value)
					startIndex = startIndex + 1
				}
			} else {
				//  __osHeaders Catched!
				keyUpdate[startIndex] = "osHeaders"
				valueUpdate[startIndex] = ConvertOsheaders(value.(messaging.ControlHeaders))
				startIndex = startIndex + 1
			}
		}

		//Convert them to LOWERCASE

		primaryKeyWord := strings.ToLower(request.Body.Parameters.KeyProperty)

		for key, value := range keyUpdate {
			keyUpdate[key] = strings.ToLower(value)
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
		//fmt.Println("ID to look : " + request.Body.Parameters.KeyProperty)
		//fmt.Println("Value for ID : " + getNoSqlKey(request))
		request.Log("Query : " + "UPDATE " + tableName + " SET " + argValueList + " WHERE " + primaryKeyWord + " =" + "'" + request.Controls.Id + "'")
		er, err := conn.Client.Execute("UPDATE " + tableName + " SET " + argValueList + " WHERE " + primaryKeyWord + " =" + "'" + request.Controls.Id + "'")

		if er == nil && err == nil {
			response.IsSuccess = true
			response.Message = "Successfully Updated a single object in to HIVE"
			request.Log(response.Message)
		} else {
			response.IsSuccess = false
			response.GetErrorResponse("Error Updating a single object in to HIVE" + err.Error())
		}
	}
	if conn != nil {
		conn.Checkin()
	}
	return response

}

func (repository HiveRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-MULTIPLE")
	request.Log("Get Search not implemented in Hive Db repository")
	return getDefaultNotImplemented()
}

func (repository HiveRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-SINGLE")
	request.Log("Get Search not implemented in Hive Db repository")
	return getDefaultNotImplemented()
}

func (repository HiveRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	request.Log("Starting SPECIAL!")
	queryType := request.Body.Special.Type

	switch queryType {
	case "getFields":
		request.Log("Starting GET-FIELDS sub routine!")
		fieldsInByte := executeHiveGetFields(request)
		if fieldsInByte != nil {
			response.IsSuccess = true
			response.Message = "Successfully Retrieved Fileds on Class : " + request.Controls.Class
			response.GetResponseWithBody(fieldsInByte)
		} else {
			response.IsSuccess = false
			response.Message = "Aborted! Unsuccessful Retrieving Fileds on Class : " + request.Controls.Class
			errorMessage := response.Message
			response.GetErrorResponse(errorMessage)
		}
	case "getClasses":
		request.Log("Starting GET-CLASSES sub routine")
		fieldsInByte := executeHiveGetClasses(request)
		if fieldsInByte != nil {
			response.IsSuccess = true
			response.Message = "Successfully Retrieved Fileds on Class : " + request.Controls.Class
			response.GetResponseWithBody(fieldsInByte)
		} else {
			response.IsSuccess = false
			response.Message = "Aborted! Unsuccessful Retrieving Fileds on Class : " + request.Controls.Class
			errorMessage := response.Message
			response.GetErrorResponse(errorMessage)
		}
	case "getNamespaces":
		request.Log("Starting GET-NAMESPACES sub routine")
		fieldsInByte := executeHiveGetNamespaces(request)
		if fieldsInByte != nil {
			response.IsSuccess = true
			response.Message = "Successfully Retrieved All Namespaces"
			response.GetResponseWithBody(fieldsInByte)
		} else {
			response.IsSuccess = false
			response.Message = "Aborted! Unsuccessful Retrieving All Namespaces"
			errorMessage := response.Message
			response.GetErrorResponse(errorMessage)
		}
	case "getSelected":
		request.Log("Starting GET-SELECTED sub routine")
		fieldsInByte := executeHiveGetSelected(request)
		if fieldsInByte != nil {
			response.IsSuccess = true
			response.Message = "Successfully Retrieved All Namespaces"
			response.GetResponseWithBody(fieldsInByte)
		} else {
			response.IsSuccess = false
			response.Message = "Aborted! Unsuccessful Retrieving All Namespaces"
			errorMessage := response.Message
			response.GetErrorResponse(errorMessage)
		}
	default:
		return search(request, request.Body.Special.Parameters)

	}

	return response
}

func (repository HiveRepository) Test(request *messaging.ObjectRequest) {

}

//Sub Routines

func executeHiveGetFields(request *messaging.ObjectRequest) (returnByte []byte) {

	namespace := getSQLnamespace(request)
	class := request.Controls.Class

	conn, isError, _ := getHiveConnection(request)
	if isError == false {

		er, err := conn.Client.Execute("describe " + namespace + "." + class)
		if er == nil && err == nil {

			//Get Schema
			schema, _, _ := conn.Client.GetSchema()

			var allMaps map[string]interface{}
			allMaps = make(map[string]interface{})

			recordIndex := 0

			for {
				row, _, _ := conn.Client.FetchOne()
				if row == "" {
					break
				} else {
					var myMap map[string]string
					myMap = make(map[string]string)

					temp := strings.Split(row, "\t")
					index := 0
					for key, value := range temp {
						value = strings.TrimSpace(value)
						if value != "" && value != " " && schema.FieldSchemas[key].Name == "col_name" {
							myMap[strconv.Itoa(index)] = value
						}
						index++
					}
					allMaps[strconv.Itoa(recordIndex)] = myMap
					recordIndex++
				}
			}
			numberOfFields := len(allMaps)
			var allFields []string
			allFields = make([]string, numberOfFields)

			for key, value := range allMaps {
				for _, value2 := range value.(map[string]string) {
					index, _ := strconv.Atoi(key)
					allFields[index] = value2
				}
			}

			byteValue, errMarshal := json.Marshal(allFields)

			if errMarshal != nil {
				byteValue = nil
			}

			returnByte = byteValue
		} else {
			returnByte = nil
		}
	}

	return
}

func executeHiveGetClasses(request *messaging.ObjectRequest) (returnByte []byte) {

	namespace := getSQLnamespace(request)

	conn, isError, _ := getHiveConnection(request)
	if isError == false {
		er, err := conn.Client.Execute("use " + namespace)
		er, err = conn.Client.Execute("show tables")
		if er == nil && err == nil {

			var allMaps map[string]interface{}
			allMaps = make(map[string]interface{})

			recordIndex := 0

			for {
				row, _, _ := conn.Client.FetchOne()
				if row == "" {
					break
				} else {
					var myMap map[string]string
					myMap = make(map[string]string)

					temp := strings.Split(row, "\t")
					index := 0
					for _, value := range temp {
						myMap[strconv.Itoa(index)] = value
						index++
					}
					allMaps[strconv.Itoa(recordIndex)] = myMap
					recordIndex++
				}
			}
			numberOfFields := len(allMaps)
			var allClasses []string
			allClasses = make([]string, numberOfFields)

			for key, value := range allMaps {
				for _, value2 := range value.(map[string]string) {
					index, _ := strconv.Atoi(key)
					allClasses[index] = value2
				}
			}

			byteValue, errMarshal := json.Marshal(allClasses)

			if errMarshal != nil {
				byteValue = nil
			}

			returnByte = byteValue
		}
	}

	return
}

func executeHiveGetNamespaces(request *messaging.ObjectRequest) (returnByte []byte) {
	conn, isError, _ := getHiveConnection(request)
	if isError == false {
		er, err := conn.Client.Execute("show databases")
		if er == nil && err == nil {

			var allMaps map[string]interface{}
			allMaps = make(map[string]interface{})

			recordIndex := 0

			for {
				row, _, _ := conn.Client.FetchOne()
				if row == "" {
					break
				} else {
					var myMap map[string]string
					myMap = make(map[string]string)

					temp := strings.Split(row, "\t")
					index := 0
					for _, value := range temp {
						myMap[strconv.Itoa(index)] = value
						index++
					}
					allMaps[strconv.Itoa(recordIndex)] = myMap
					recordIndex++
				}
			}
			numberOfFields := len(allMaps)
			var allClasses []string
			allClasses = make([]string, numberOfFields)

			for key, value := range allMaps {
				for _, value2 := range value.(map[string]string) {
					index, _ := strconv.Atoi(key)
					allClasses[index] = value2
				}
			}

			byteValue, errMarshal := json.Marshal(allClasses)

			if errMarshal != nil {
				byteValue = nil
			}

			returnByte = byteValue
		}
	}

	return
}

func executeHiveGetSelected(request *messaging.ObjectRequest) (returnByte []byte) {
	conn, isError, _ := getHiveConnection(request)
	if isError == false {

		tableName := getSQLnamespace(request) + "." + request.Controls.Class

		var selectedItemsQuery string

		var requestedFields []string
		request.Log("Requested Field List : " + request.Body.Special.Parameters)
		if request.Body.Special.Parameters == "*" {
			request.Log("All fields requested")
			requestedFields = make([]string, 1)
			requestedFields[0] = "*"
			selectedItemsQuery = "*"
		} else {
			requestedFields = strings.Split(request.Body.Special.Parameters, " ")

			for key, value := range requestedFields {
				if key == len(requestedFields)-1 {
					selectedItemsQuery += value
				} else {
					selectedItemsQuery += (value + ",")
				}
			}
		}

		er, err := conn.Client.Execute("SELECT " + selectedItemsQuery + " FROM " + tableName)
		request.Log("SELECT " + selectedItemsQuery + " FROM " + tableName)
		if er == nil && err == nil {

			//Get Schema
			schema, _, _ := conn.Client.GetSchema()

			var allMaps map[string]interface{}
			allMaps = make(map[string]interface{})

			var recordIndex int
			recordIndex = 0

			for {
				row, _, _ := conn.Client.FetchOne()
				if row == "" {
					break
				} else {

					var myMap map[string]string
					myMap = make(map[string]string)

					temp := strings.Split(row, "\t")

					for key, _ := range temp {
						myMap[(schema.FieldSchemas[key].Name)] = temp[key]
					}

					allMaps[strconv.Itoa(recordIndex)] = myMap

					recordIndex++

				}
			}

			byteValue, _ := json.Marshal(allMaps)
			returnByte = byteValue

		} else {
			returnByte = []byte("Error fetching values from HIVE server!")
		}

	} else {
		returnByte = []byte("Error Connecting to HIVE server!")
	}

	return
}

//Supplimentary Functions

func checkDBAndCreateOnNotAvailbale(databaseName string, conn *hive.HiveConnection) (IsSuccess bool) {
	IsSuccess = true
	fmt.Println("THIS PROCESS WILL CHECK FOR REQUIRED DATABASE AND CREATE NEW IF NECESSERY!")
	fmt.Println("Getting All Databases in the Hive Metastore....")

	er, err := conn.Client.Execute("show databases")
	if er == nil && err == nil {

		var myMap map[string]string
		myMap = make(map[string]string)
		recordNumber := 1

		for {
			row, _, _ := conn.Client.FetchOne()
			if row == "" {
				break
			} else {
				var temp []string
				temp = strings.Split(row, "\t")
				var temp2 string
				for i := 0; i < len(temp); i++ {
					temp2 = temp2 + " " + temp[i]
				}
				myMap[strconv.Itoa(recordNumber)] = temp2
				recordNumber = recordNumber + 1
			}
		}

		fmt.Print("All available Databases are : ")
		fmt.Println(myMap)

		fmt.Println("Begining checking for required Database....")

		isFound := false

		for _, value := range myMap {
			//fmt.Println("Checking database : " + value + " against : " + databaseName)
			if strings.TrimSpace(value) == databaseName {
				fmt.Println("Found DB!")
				isFound = true
				IsSuccess = true
				break
			}
		}

		if isFound {
			fmt.Println("Database Found! No need to Recreate!")
		} else {
			fmt.Println("Database Not Found! Creating new Database!")

			er, err = conn.Client.Execute("create database " + databaseName)
			if er != nil || err != nil {
				IsSuccess = false
				fmt.Println("Database Creation Error! Check Hadoop Hive Configuration - " + err.Error())
			} else {
				IsSuccess = true
				fmt.Println("Database Succsessfully Created!")
			}
		}

	}

	return
}

func checkIfRecordAvailable(recordId string, keyProperty string, tableDB string, conn *hive.HiveConnection) (isAvailable bool) {
	isAvailable = false

	fmt.Println("Checking if Record is already avaiable.")
	fmt.Println("Executing : " + ("select " + keyProperty + " from " + tableDB + " where " + keyProperty + "='" + recordId + "'"))
	er, err := conn.Client.Execute("select " + keyProperty + " from " + tableDB + " where " + keyProperty + "='" + recordId + "'")
	if er == nil && err == nil {

		for {
			row, _, _ := conn.Client.FetchOne()
			if row == "" {
				break
			} else {
				var temp []string
				temp = strings.Split(row, "\t")

				for _, value := range temp {
					fmt.Println("|" + strings.TrimSpace(value) + "| --- |" + recordId + "|")
					if strings.TrimSpace(value) == recordId {
						isAvailable = true
						break
					}
				}
			}
		}
	}

	return

}

func checkIfTenantIsAllowed(query string, tenent string) (isAllowed bool) {

	isAllowed = false
	tables, isException := queryparser.GetTablesInQuery(query)
	if isException {
		isAllowed = true
	} else {
		for _, value := range tables {
			tableElements := strings.Split(value, ".")
			value = tableElements[0] + "." + tableElements[1] + "." + tableElements[2]
			fmt.Println("Checking -> " + tenent + " vs " + value)
			if tenent == value {

				isAllowed = true
			} else {
			}
		}
	}

	return
}

func getHiveRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
	isGUIDKey := false
	isAutoIncrementId := false //else MANUAL key from the user

	if obj == nil {
		//single request
		if (request.Controls.Id == "-999") || (request.Body.Parameters.AutoIncrement == true) {
			isAutoIncrementId = true
		}

		if (request.Controls.Id == "-888") || (request.Body.Parameters.GUIDKey == true) {
			isGUIDKey = true
		}

	} else {
		//multiple requests
		if (obj[request.Body.Parameters.KeyProperty].(string) == "-999") || (request.Body.Parameters.AutoIncrement == true) {
			isAutoIncrementId = true
		}

		if (obj[request.Body.Parameters.KeyProperty].(string) == "-888") || (request.Body.Parameters.GUIDKey == true) {
			isGUIDKey = true
		}

	}

	if isGUIDKey {
		request.Log("GUID Key generation requested!")
		returnID = uuid.NewV1().String()
	} else if isAutoIncrementId {
		request.Log("Automatic Increment Key generation requested!")
		returnID = uuid.NewV1().String()
	} else {
		request.Log("Manual Key requested!")
		if obj == nil {
			returnID = request.Controls.Id
		} else {
			returnID = obj[request.Body.Parameters.KeyProperty].(string)
		}
	}

	return
}

func getHiveFieldOrder(request *messaging.ObjectRequest) []string {
	var returnArray []string
	//read fields
	byteValue := executeHiveGetFields(request)
	if byteValue == nil {
		request.Log("Table is not available")
		returnArray = make([]string, 1)
		returnArray[0] = "nil"
	} else {
		err := json.Unmarshal(byteValue, &returnArray)
		if err != nil {
			request.Log("Converstion of Json Failed!")
			returnArray = make([]string, 1)
			returnArray[0] = "nil"
			return returnArray
		}
	}

	return returnArray
}

func checkTableAvailability(request *messaging.ObjectRequest) (status bool) {
	status = false

	byteValue := executeHiveGetFields(request)

	if byteValue == nil {
		request.Log("Table is not available")
		status = false
	} else {
		status = true
	}

	return
}
