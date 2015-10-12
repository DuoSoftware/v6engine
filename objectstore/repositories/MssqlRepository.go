package repositories

import (
	"database/sql"
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/twinj/uuid"
	"reflect"
	"strconv"
	"strings"
)

type MssqlRepository struct {
}

func (repository MssqlRepository) GetRepositoryName() string {
	return "Mssql DB"
}

func getMssqlSQLnamespace(request *messaging.ObjectRequest) string {
	fmt.Println()
	namespace := strings.Replace(request.Controls.Namespace, ".", "", -1)
	return namespace
}

func getMssqlConnection(request *messaging.ObjectRequest) (session *sql.DB, isError bool, errorMessage string) {

	isError = false
	username := request.Configuration.ServerConfiguration["MSSQL"]["Username"]
	password := request.Configuration.ServerConfiguration["MSSQL"]["Password"]
	dbUrl := request.Configuration.ServerConfiguration["MSSQL"]["Server"]
	dbPort := request.Configuration.ServerConfiguration["MSSQL"]["Port"]

	session, err := sql.Open("mssql", "server="+dbUrl+";port="+dbPort+";user id="+username+";password="+password+";encrypt=disable")

	if err != nil {
		isError = true
		request.Log("There is an error")
		errorMessage = err.Error()
		request.Log("MsSql connection initilizing failed!")
	} else {
		if !verifyMsSqlDatabase(session, request) {
			return
		}
	}
	request.Log("Reusing existing MsSql connection")
	return
}

func verifyMsSqlDatabase(session *sql.DB, request *messaging.ObjectRequest) bool {
	//get list of databases
	rows, err := session.Query("SELECT name FROM sys.databases;")
	if err != nil {
		request.Log("Query Error : " + err.Error())
	} else {
		columns, _ := rows.Columns()
		count := len(columns)
		values := make([]interface{}, count)
		valuePtrs := make([]interface{}, count)

		var tempMap map[int]interface{}
		tempMap = make(map[int]interface{})
		index := 0
		for rows.Next() {

			for i, _ := range columns {
				valuePtrs[i] = &values[i]
			}

			rows.Scan(valuePtrs...)

			for i, _ := range columns {

				var v interface{}

				val := values[i]

				b, ok := val.([]byte)

				if ok {
					v = string(b)
				} else {
					v = val
				}

				tempMap[index] = v
				index++

			}
		}

		fmt.Print("Available Databases : ")
		fmt.Println(tempMap)
		requestedDB := getMssqlSQLnamespace(request)
		for _, dbName := range tempMap {
			if requestedDB == dbName.(string) {
				fmt.Println("Comparing : " + requestedDB + " with : " + dbName.(string))
				request.Log("Database Already Available! Nothing to do....")
				return true
			}
		}

		//else create the database
		_, err := session.Query("create database [" + getMssqlSQLnamespace(request) + "]")
		if err != nil {
			request.Log("Database Creation Failed : " + err.Error())
			return false
		} else {
			request.Log("New " + getMssqlSQLnamespace(request) + " database created successfully!")
			return true
		}

	}

	return true
}

func (repository MssqlRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-ALL")
	response := RepositoryResponse{}
	session, isError, errorMessage := getMssqlConnection(request)
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		isError = false
		skip := "0"
		if request.Extras["skip"] != nil {
			skip = request.Extras["skip"].(string)
		}

		take := "10000"
		if request.Extras["take"] != nil {
			take = request.Extras["take"].(string)
		}

		//get Primary Key of database
		fieldName := ""
		request.Log("Getting Primary Key")
		rows, err := session.Query("SELECT KCU.COLUMN_NAME AS COLUMN_NAME FROM [" + getMssqlSQLnamespace(request) + "].INFORMATION_SCHEMA.TABLE_CONSTRAINTS TC JOIN [" + getMssqlSQLnamespace(request) + "].INFORMATION_SCHEMA.KEY_COLUMN_USAGE KCU ON KCU.CONSTRAINT_SCHEMA = TC.CONSTRAINT_SCHEMA AND KCU.CONSTRAINT_NAME = TC.CONSTRAINT_NAME AND KCU.TABLE_SCHEMA = TC.TABLE_SCHEMA AND KCU.TABLE_NAME = TC.TABLE_NAME WHERE TC.CONSTRAINT_TYPE = 'PRIMARY KEY' AND KCU.TABLE_NAME='" + request.Controls.Class + "';")
		var keyMap map[string]interface{}
		keyMap = make(map[string]interface{})
		if err != nil {
			response.IsSuccess = false
			response.Message = "Error reading primary key"
			request.Log("Error reading primary key")
			return response
		} else {

			request.Log("Successfully returned primary Key for table")

			columns, _ := rows.Columns()
			count := len(columns)
			values := make([]interface{}, count)
			valuePtrs := make([]interface{}, count)

			for rows.Next() {
				for i, _ := range columns {
					valuePtrs[i] = &values[i]
				}

				rows.Scan(valuePtrs...)

				for i, col := range columns {

					var v interface{}

					val := values[i]

					b, ok := val.([]byte)

					if ok {
						v = string(b)
					} else {
						v = val
					}

					keyMap[col] = v

				}
			}
		}

		if keyMap["COLUMN_NAME"] == nil {
			var emptymap map[string]interface{}
			emptymap = make(map[string]interface{})
			finalBytes, _ := json.Marshal(emptymap)
			response.GetResponseWithBody(finalBytes)
			return response
		}

		fieldName = keyMap["COLUMN_NAME"].(string)

		request.Log("Primary Key Field : " + fieldName)
		//Getting All records
		var returnMap []map[string]interface{}

		//request.Log("select * from [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "];")
		//rows, err := session.Query("select * from [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "];")
		request.Log("select top " + take + " * from (select *, ROW_NUMBER() over (order by " + fieldName + ") as r_n_n from [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "]) xx where r_n_n >=" + skip + ";")
		rowss, err := session.Query("select top " + take + " * from (select *, ROW_NUMBER() over (order by " + fieldName + ") as r_n_n from [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "]) xx where r_n_n >=" + skip + ";")

		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting values for all objects in MsSql" + err.Error())
			response.Message = "Table Not Found in Database : " + getMssqlSQLnamespace(request)
		} else {
			response.IsSuccess = true
			response.Message = "Successfully retrieved values for all objects in MsSql"
			request.Log(response.Message)

			columns, _ := rowss.Columns()
			count := len(columns)
			values := make([]interface{}, count)
			valuePtrs := make([]interface{}, count)

			for rowss.Next() {

				var tempMap map[string]interface{}
				tempMap = make(map[string]interface{})

				for i, _ := range columns {
					valuePtrs[i] = &values[i]
				}

				rowss.Scan(valuePtrs...)

				for i, col := range columns {

					var v interface{}

					val := values[i]

					b, ok := val.([]byte)

					if ok {
						v = string(b)
					} else {
						v = val
					}

					tempMap[col] = v
				}

				returnMap = append(returnMap, tempMap)

			}

			if request.Controls.SendMetaData == "false" {

				for index, arrVal := range returnMap {
					for key, _ := range arrVal {
						if key == "osheaders" {
							delete(returnMap[index], key)
						}
					}
				}
			}

			for index, arrVal := range returnMap {
				for key, _ := range arrVal {
					if key == "r_n_n" {
						delete(returnMap[index], key)
					}
				}
			}

			byteValue, errMarshal := json.Marshal(returnMap)
			if errMarshal != nil {
				response.IsSuccess = false
				response.GetErrorResponse("Error getting values for all objects in MsSql" + err.Error())
			} else {
				response.IsSuccess = true
				response.GetResponseWithBody(byteValue)
				response.Message = "Successfully retrieved values for all objects in MsSql"
				request.Log(response.Message)
			}
		}
	}
	session.Close()
	return response
}

func (repository MssqlRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Get Search not implemented in Mssql Db repository")
	return getDefaultNotImplemented()
}

func (repository MssqlRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-QUERY")
	response := RepositoryResponse{}

	queryType := request.Body.Query.Type
	switch queryType {
	case "Query":
		if request.Body.Query.Parameters != "*" {
			fieldsInByte := executeMssqlQuery(request)
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
		request.Log(queryType + " not implemented in Mssql Db repository")
		return getDefaultNotImplemented()

	}

	return response
}

func executeMssqlQuery(request *messaging.ObjectRequest) (returnByte []byte) {
	session, isError, _ := getMssqlConnection(request)
	if isError == true {
		returnByte = nil
	} else {
		isError = false
		//Process A : Get Count of DB
		var returnMap map[string]interface{}
		returnMap = make(map[string]interface{})

		rows, err := session.Query(request.Body.Query.Parameters)

		if err != nil {
			request.Log("Error executing query in Mssql")
		} else {
			request.Log("Successfully executed query in Mssql")
			columns, _ := rows.Columns()
			count := len(columns)
			values := make([]interface{}, count)
			valuePtrs := make([]interface{}, count)

			index := 0
			for rows.Next() {

				var tempMap map[string]interface{}
				tempMap = make(map[string]interface{})

				for i, _ := range columns {
					valuePtrs[i] = &values[i]
				}

				rows.Scan(valuePtrs...)

				for i, col := range columns {

					var v interface{}

					val := values[i]

					b, ok := val.([]byte)

					if ok {
						v = string(b)
					} else {
						v = val
					}

					tempMap[col] = v

				}

				returnMap[strconv.Itoa(index)] = tempMap
				index++
			}

			byteValue, errMarshal := json.Marshal(returnMap)
			if errMarshal != nil {
				request.Log("Error converting to byte array")
				byteValue = nil
			} else {
				request.Log("Successfully converted result to byte array")
			}

			returnByte = byteValue
		}

	}
	session.Close()
	return returnByte
}

func (repository MssqlRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {

	request.Log("Starting GET-BY-KEY")
	response := RepositoryResponse{}
	session, isError, errorMessage := getMssqlConnection(request)
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		isError = false
		request.Log("Id key : " + request.Controls.Id)

		var myMap map[string]interface{}
		myMap = make(map[string]interface{})

		var keyMap map[string]interface{}
		keyMap = make(map[string]interface{})

		fieldName := ""
		parameter := request.Controls.Id
		if request.Extras["fieldName"] != nil {
			fieldName = request.Extras["fieldName"].(string)
			parameter = request.Controls.Id
		} else {
			request.Log("Getting Primary Key")
			rows, err := session.Query("SELECT KCU.COLUMN_NAME AS COLUMN_NAME FROM [" + getMssqlSQLnamespace(request) + "].INFORMATION_SCHEMA.TABLE_CONSTRAINTS TC JOIN [" + getMssqlSQLnamespace(request) + "].INFORMATION_SCHEMA.KEY_COLUMN_USAGE KCU ON KCU.CONSTRAINT_SCHEMA = TC.CONSTRAINT_SCHEMA AND KCU.CONSTRAINT_NAME = TC.CONSTRAINT_NAME AND KCU.TABLE_SCHEMA = TC.TABLE_SCHEMA AND KCU.TABLE_NAME = TC.TABLE_NAME WHERE TC.CONSTRAINT_TYPE = 'PRIMARY KEY' AND KCU.TABLE_NAME='" + request.Controls.Class + "';")

			if err != nil {
				response.IsSuccess = false
				response.Message = "Error reading primary key"
				request.Log("Error reading primary key")
				return response
			} else {

				request.Log("Successfully returned primary Key for table")

				columns, _ := rows.Columns()
				count := len(columns)
				values := make([]interface{}, count)
				valuePtrs := make([]interface{}, count)

				for rows.Next() {
					for i, _ := range columns {
						valuePtrs[i] = &values[i]
					}

					rows.Scan(valuePtrs...)

					for i, col := range columns {

						var v interface{}

						val := values[i]

						b, ok := val.([]byte)

						if ok {
							v = string(b)
						} else {
							v = val
						}

						keyMap[col] = v

					}
				}
			}

			if keyMap["COLUMN_NAME"] == nil {
				var emptymap map[string]interface{}
				emptymap = make(map[string]interface{})
				finalBytes, _ := json.Marshal(emptymap)
				response.GetResponseWithBody(finalBytes)
				return response
			}

			fieldName = keyMap["COLUMN_NAME"].(string)
		}

		request.Log("KeyProperty : " + fieldName)
		request.Log("KeyValue : " + request.Controls.Id)
		rows, err := session.Query("SELECT * FROM [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "] where " + fieldName + " = '" + parameter + "';")

		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting values for all objects in MsSql" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully retrieved values for all objects in MsSql"
			request.Log(response.Message)

			columns, _ := rows.Columns()
			count := len(columns)
			values := make([]interface{}, count)
			valuePtrs := make([]interface{}, count)

			for rows.Next() {
				for i, _ := range columns {
					valuePtrs[i] = &values[i]
				}

				rows.Scan(valuePtrs...)

				for i, col := range columns {

					var v interface{}

					val := values[i]

					b, ok := val.([]byte)

					if ok {
						v = string(b)
					} else {
						v = val
					}

					myMap[col] = v

				}
			}

			if request.Controls.SendMetaData == "false" {
				for index, _ := range myMap {
					if index == "osheaders" {
						delete(myMap, index)
					}
				}
			}

			for index, _ := range myMap {
				if index == "r_n_n" {
					delete(myMap, index)
				}
			}

			byteValue, errMarshal := json.Marshal(myMap)
			if errMarshal != nil {
				response.IsSuccess = false
				response.GetErrorResponse("Error getting values for all objects in MsSql" + err.Error())
			} else {
				response.IsSuccess = true
				response.GetResponseWithBody(byteValue)
				response.Message = "Successfully retrieved values for all objects in MsSql"
				request.Log(response.Message)
			}
		}
	}
	session.Close()
	return response

	//request.Log("Get By Key not implemented in Mssql Db repository")
	//return getDefaultNotImplemented()
}

func (repository MssqlRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-MULTIPLE")
	response := RepositoryResponse{}
	session, isError, errorMessage := getMssqlConnection(request)

	var idData map[string]interface{}
	idData = make(map[string]interface{})

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

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
					tempMapObject[key] = value
				}
			}

			DataObjects[i] = tempMapObject
		}
		//check for table in MsSql

		if createMssqlTable(request, session) {
			request.Log("Table Verified Successfully!")
		} else {
			response.IsSuccess = false
			return response
		}

		indexNames := getMssqlFieldOrder(request)

		var argKeyList string
		var argValueList string

		//create keyvalue list

		for i := 0; i < len(indexNames); i++ {
			if i != len(indexNames)-1 {
				argKeyList = argKeyList + indexNames[i] + ", "
			} else {
				argKeyList = argKeyList + indexNames[i]
			}
		}

		noOf500Sets := (len(DataObjects) / 500)
		remainderFromSets := 0
		statusCount := noOf500Sets
		remainderFromSets = (len(DataObjects) - (noOf500Sets * 500))
		if remainderFromSets > 0 {
			statusCount++
		}
		var setStatus []bool
		setStatus = make([]bool, statusCount)

		startIndex := 0
		stopIndex := 500
		statusIndex := 0

		for x := 0; x < noOf500Sets; x++ {
			argValueList = ""

			for i, _ := range DataObjects[startIndex:stopIndex] {
				i += startIndex
				noOfElements := len(DataObjects[i])
				request.Log("Serializing Object no : " + strconv.Itoa(i))
				keyValue := getMssqlSqlRecordID(request, DataObjects[i])
				DataObjects[i][request.Body.Parameters.KeyProperty] = keyValue
				idData[strconv.Itoa(i)] = keyValue
				if keyValue == "" {
					response.IsSuccess = false
					response.Message = "Failed inserting multiple object in MsSql"
					request.Log(response.Message)
					request.Log("Inavalid ID request")
					return response
				}

				var keyArray = make([]string, noOfElements)
				var valueArray = make([]string, noOfElements)

				for index := 0; index < len(indexNames); index++ {
					if indexNames[index] != "osheaders" {

						if _, ok := DataObjects[i][indexNames[index]].(string); ok {
							keyArray[index] = indexNames[index]
							valueArray[index] = DataObjects[i][indexNames[index]].(string)
						} else {
							fmt.Println("Non string value detected, Will be strigified!")
							keyArray[index] = indexNames[index]
							valueArray[index] = getStringByObject(DataObjects[i][indexNames[index]])
						}
					} else {
						// __osHeaders Catched!
						keyArray[index] = "osheaders"
						valueArray[index] = ConvertOsheaders(DataObjects[i][indexNames[index]].(messaging.ControlHeaders))
					}

				}
				argValueList += "("

				//Build the query string
				for i := 0; i < noOfElements; i++ {
					if i != noOfElements-1 {
						argValueList = argValueList + "'" + valueArray[i] + "'" + ", "
					} else {
						argValueList = argValueList + "'" + valueArray[i] + "'"
					}
				}

				i -= startIndex
				if i != len(DataObjects[startIndex:stopIndex])-1 {
					argValueList += "),"
				} else {
					argValueList += ")"
				}

			}

			//DEBUG USE : Display Query information
			//	fmt.Println("Table Name : " + request.Controls.Class)
			//	fmt.Println("Key list : " + argKeyList)
			//fmt.Println("Value list : " + argValueList)
			//request.Log("INSERT INTO " + request.Controls.Class + " (" + argKeyList + ") VALUES " + argValueList + ";")
			//request.Log("INSERT INTO [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "] (" + argKeyList + ") VALUES " + argValueList + ";")
			_, err := session.Query("INSERT INTO [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "] (" + argKeyList + ") VALUES " + argValueList + ";")
			if err != nil {
				setStatus[statusIndex] = false
				request.Log("ERROR : " + err.Error())
			} else {
				request.Log("INSERTED SUCCESSFULLY")
				setStatus[statusIndex] = true
			}

			statusIndex++
			startIndex += 500
			stopIndex += 500
		}

		if remainderFromSets > 0 {
			argValueList = ""
			start := len(DataObjects) - remainderFromSets

			for i, _ := range DataObjects[start:len(DataObjects)] {
				i += start
				noOfElements := len(DataObjects[i])
				request.Log("Serializing Object no : " + strconv.Itoa(i))
				keyValue := getMssqlSqlRecordID(request, DataObjects[i])
				DataObjects[i][request.Body.Parameters.KeyProperty] = keyValue
				idData[strconv.Itoa(i)] = keyValue
				if keyValue == "" {
					response.IsSuccess = false
					response.Message = "Failed inserting multiple object in MsSql"
					request.Log(response.Message)
					request.Log("Inavalid ID request")
					return response
				}

				var keyArray = make([]string, noOfElements)
				var valueArray = make([]string, noOfElements)

				for index := 0; index < len(indexNames); index++ {
					if indexNames[index] != "osheaders" {

						if _, ok := DataObjects[i][indexNames[index]].(string); ok {
							keyArray[index] = indexNames[index]
							valueArray[index] = DataObjects[i][indexNames[index]].(string)
						} else {
							fmt.Println("Non string value detected, Will be strigified!")
							keyArray[index] = indexNames[index]
							valueArray[index] = getStringByObject(DataObjects[i][indexNames[index]])
						}
					} else {
						// __osHeaders Catched!
						keyArray[index] = "osheaders"
						valueArray[index] = ConvertOsheaders(DataObjects[i][indexNames[index]].(messaging.ControlHeaders))
					}

				}

				argValueList += "("

				//Build the query string
				for i := 0; i < noOfElements; i++ {
					if i != noOfElements-1 {
						argValueList = argValueList + "'" + valueArray[i] + "'" + ", "
					} else {
						argValueList = argValueList + "'" + valueArray[i] + "'"
					}
				}

				i -= start
				if i != len(DataObjects[start:len(DataObjects)])-1 {
					argValueList += "),"
				} else {
					argValueList += ")"
				}

			}

			//DEBUG USE : Display Query information
			//	fmt.Println("Table Name : " + request.Controls.Class)
			//	fmt.Println("Key list : " + argKeyList)
			//fmt.Println("Value list : " + argValueList)
			//request.Log("INSERT INTO " + request.Controls.Class + " (" + argKeyList + ") VALUES " + argValueList + ";")
			//request.Log("INSERT INTO [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "] (" + argKeyList + ") VALUES " + argValueList + ";")
			_, err := session.Query("INSERT INTO [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "] (" + argKeyList + ") VALUES " + argValueList + ";")
			if err != nil {
				setStatus[statusIndex] = false
				request.Log("ERROR : " + err.Error())
			} else {
				request.Log("INSERTED SUCCESSFULLY")
				setStatus[statusIndex] = true
			}
		}

		isAllCompleted := true
		for _, value := range setStatus {
			if value == false {
				isAllCompleted = false
				break
			}
		}

		if isAllCompleted {
			response.IsSuccess = true
			response.Message = "Successfully inserted many objects in to Mssql"
			request.Log(response.Message)
		} else {
			response.IsSuccess = false
			response.Message = "Error inserting many objects in to Mssql"
			request.Log(response.Message)
			response.GetErrorResponse("Error inserting many objects in to Mssql")
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

	session.Close()
	return response
}

func (repository MssqlRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-SINGLE")
	response := RepositoryResponse{}

	keyValue := getMssqlSqlRecordID(request, nil)
	session, isError, errorMessage := getMssqlConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else if keyValue != "" {
		//change field names to Lower Case
		var DataObject map[string]interface{}
		DataObject = make(map[string]interface{})

		for key, value := range request.Body.Object {
			if key == "__osHeaders" {
				DataObject["osheaders"] = value
			} else {
				DataObject[key] = value
			}
		}

		noOfElements := len(DataObject)
		DataObject[request.Body.Parameters.KeyProperty] = keyValue

		if createMssqlTable(request, session) {
			request.Log("Table Verified Successfully!")
		} else {
			response.IsSuccess = false
			return response
		}

		indexNames := getMssqlFieldOrder(request)

		var argKeyList string
		var argValueList string

		//create keyvalue list

		for i := 0; i < len(indexNames); i++ {
			if i != len(indexNames)-1 {
				argKeyList = argKeyList + indexNames[i] + ", "
			} else {
				argKeyList = argKeyList + indexNames[i]
			}
		}

		var keyArray = make([]string, noOfElements)
		var valueArray = make([]string, noOfElements)

		// Process A :start identifying individual data in array and convert to string
		for index := 0; index < len(indexNames); index++ {
			if indexNames[index] != "osheaders" {

				if _, ok := DataObject[indexNames[index]].(string); ok {
					keyArray[index] = indexNames[index]
					valueArray[index] = DataObject[indexNames[index]].(string)
				} else {
					fmt.Println("Non string value detected, Will be strigified!")
					keyArray[index] = indexNames[index]
					valueArray[index] = getStringByObject(DataObject[indexNames[index]])
				}
			} else {
				// __osHeaders Catched!
				keyArray[index] = "osheaders"
				valueArray[index] = ConvertOsheaders(DataObject[indexNames[index]].(messaging.ControlHeaders))
			}

		}

		//Build the query string
		for i := 0; i < noOfElements; i++ {
			if i != noOfElements-1 {
				argValueList = argValueList + "'" + valueArray[i] + "'" + ", "
			} else {
				argValueList = argValueList + "'" + valueArray[i] + "'"
			}
		}
		//..........................................

		//DEBUG USE : Display Query information
		//fmt.Println("Table Name : " + request.Controls.Class)
		//fmt.Println("Key list : " + argKeyList)
		//fmt.Println("Value list : " + argValueList)
		//request.Log("INSERT INTO " + request.Controls.Class + " (" + argKeyList + ") VALUES (" + argValueList + ")")
		request.Log("INSERT INTO [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "] (" + argKeyList + ") VALUES (" + argValueList + ")")
		_, err := session.Query("INSERT INTO [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "] (" + argKeyList + ") VALUES (" + argValueList + ")")
		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error inserting one object in MsSql" + err.Error())
			if strings.Contains(err.Error(), "duplicate key value") {
				response.IsSuccess = true
				response.Message = "No Change since record already Available!"
				request.Log(response.Message)
				return response
			}
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted one object in MsSql"
			request.Log(response.Message)
		}
	}

	var Data []map[string]interface{}
	Data = make([]map[string]interface{}, 1)
	var actualData map[string]interface{}
	actualData = make(map[string]interface{})
	actualData["ID"] = keyValue
	Data[0] = actualData
	response.Data = Data

	session.Close()
	return response
}

func createMssqlTable(request *messaging.ObjectRequest, session *sql.DB) (status bool) {
	status = false

	//get table list
	classBytes := executeMssqlGetClasses(request)
	var classList []string
	err := json.Unmarshal(classBytes, &classList)
	fmt.Print("Recieved Table List : ")
	fmt.Println(classList)
	if err != nil {
		status = false
	} else {
		for _, className := range classList {
			if request.Controls.Class == className {
				fmt.Println("Table Already Available")
				status = true
				//Get all fields
				classBytes := executeMssqlGetFields(request)
				var tableFieldList []string
				_ = json.Unmarshal(classBytes, &tableFieldList)
				//Check For missing fields. If any ALTER TABLE
				var recordFieldList []string
				var recordFieldType []string
				if request.Body.Object == nil {
					recordFieldList = make([]string, len(request.Body.Objects[0]))
					recordFieldType = make([]string, len(request.Body.Objects[0]))
					index := 0
					for key, value := range request.Body.Objects[0] {
						if key == "__osHeaders" {
							recordFieldList[index] = "osheaders"
							recordFieldType[index] = "text"
						} else {
							recordFieldList[index] = key
							recordFieldType[index] = getMssqlDataType(value)
						}
						index++
					}
				} else {
					recordFieldList = make([]string, len(request.Body.Object))
					recordFieldType = make([]string, len(request.Body.Object))
					index := 0
					for key, value := range request.Body.Object {
						if key == "__osHeaders" {
							recordFieldList[index] = "osheaders"
							recordFieldType[index] = "text"
						} else {
							recordFieldList[index] = key
							recordFieldType[index] = getMssqlDataType(value)
						}
						index++
					}
				}

				var newFields []string
				var newTypes []string

				//check for new Fields
				for key, fieldName := range recordFieldList {
					isAvailable := false
					for _, tableField := range tableFieldList {
						if fieldName == tableField {
							isAvailable = true
							break
						}
					}

					if !isAvailable {
						newFields = append(newFields, fieldName)
						newTypes = append(newTypes, recordFieldType[key])
					}
				}

				//ALTER TABLES

				for key, _ := range newFields {
					request.Log("ALTER TABLE [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "] ADD " + newFields[key] + " " + newTypes[key] + ";")
					_, er := session.Query("ALTER TABLE [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "] ADD " + newFields[key] + " " + newTypes[key] + ";")
					if er != nil {
						status = false
						request.Log("Table Alter Failed : " + er.Error())
						return
					} else {
						status = true
						request.Log("Table Alter Success!")
					}
				}

				return
			}
		}

		// if not available
		//get one object
		var dataObject map[string]interface{}
		dataObject = make(map[string]interface{})

		if request.Body.Object != nil {
			for key, value := range request.Body.Object {
				if key == "__osHeaders" {
					dataObject["osheaders"] = value
				} else {
					dataObject[key] = value
				}
			}
		} else {
			for key, value := range request.Body.Objects[0] {
				if key == "__osHeaders" {
					dataObject["osheaders"] = value
				} else {
					dataObject[key] = value
				}
			}
		}
		//read fields
		noOfElements := len(dataObject)
		var keyArray = make([]string, noOfElements)
		var dataTypeArray = make([]string, noOfElements)

		var startIndex int = 0

		for key, value := range dataObject {
			keyArray[startIndex] = key
			dataTypeArray[startIndex] = getMssqlDataType(value)
			startIndex = startIndex + 1

		}

		//Create Table

		var argKeyList2 string

		for i := 0; i < noOfElements; i++ {
			if i != noOfElements-1 {
				if keyArray[i] == request.Body.Parameters.KeyProperty {
					argKeyList2 = argKeyList2 + keyArray[i] + " varchar(255) PRIMARY KEY, "
				} else {
					argKeyList2 = argKeyList2 + keyArray[i] + " " + dataTypeArray[i] + ", "
				}

			} else {
				if keyArray[i] == request.Body.Parameters.KeyProperty {
					argKeyList2 = argKeyList2 + keyArray[i] + " varchar(255) PRIMARY KEY"
				} else {
					argKeyList2 = argKeyList2 + keyArray[i] + " " + dataTypeArray[i]
				}

			}
		}

		request.Log("create table [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "] (" + argKeyList2 + ");")

		_, er := session.Query("create table [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "](" + argKeyList2 + ");")
		if er != nil {
			status = false
			request.Log("Table Creation Failed : " + er.Error())
			return
		}

		status = true

	}

	return
}

func getMssqlFieldOrder(request *messaging.ObjectRequest) []string {
	var returnArray []string
	//read fields
	byteValue := executeMssqlGetFields(request)

	err := json.Unmarshal(byteValue, &returnArray)
	fmt.Print("Field List from DB : ")
	fmt.Println(returnArray)
	if err != nil {
		request.Log("Converstion of Json Failed!")
		returnArray = make([]string, 1)
		returnArray[0] = "nil"
		return returnArray
	}

	return returnArray
}

func (repository MssqlRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-MULTIPLE")
	response := RepositoryResponse{}
	session, isError, errorMessage := getMssqlConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		if createMssqlTable(request, session) {
			request.Log("Table Verified Successfully!")
		} else {
			response.IsSuccess = false
			return response
		}

		for i := 0; i < len(request.Body.Objects); i++ {
			noOfElements := len(request.Body.Objects[i])
			var keyUpdate = make([]string, noOfElements)
			var valueUpdate = make([]string, noOfElements)

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
					keyUpdate[startIndex] = "osheaders"
					valueUpdate[startIndex] = ConvertOsheaders(value.(messaging.ControlHeaders))
					startIndex = startIndex + 1
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
			//	fmt.Println("Table Name : " + request.Controls.Class)
			//	fmt.Println("Value list : " + argValueList)
			_, err := session.Query("UPDATE [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "] SET " + argValueList + " WHERE " + request.Body.Parameters.KeyProperty + " =" + "'" + request.Body.Objects[i][request.Body.Parameters.KeyProperty].(string) + "'")
			if err != nil {
				response.IsSuccess = false
				request.Log("Error updating object in MsSql  : " + getNoSqlKey(request) + ", " + err.Error())
				response.GetErrorResponse("Error updating one object in MsSql because no match was found!" + err.Error())
			} else {
				response.IsSuccess = true
				response.Message = "Successfully updating one object in MsSql "
				request.Log(response.Message)
			}
		}

	}

	session.Close()
	return response
}

func (repository MssqlRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-SINGLE")
	response := RepositoryResponse{}
	session, isError, errorMessage := getMssqlConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		if createMssqlTable(request, session) {
			request.Log("Table Verified Successfully!")
		} else {
			response.IsSuccess = false
			return response
		}

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
				keyUpdate[startIndex] = "osheaders"
				valueUpdate[startIndex] = ConvertOsheaders(value.(messaging.ControlHeaders))
				startIndex = startIndex + 1
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

		_, err := session.Query("UPDATE [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "] SET " + argValueList + " WHERE " + request.Body.Parameters.KeyProperty + " =" + "'" + request.Controls.Id + "'")
		if err != nil {
			response.IsSuccess = false
			request.Log("Error updating object in MsSql  : " + getNoSqlKey(request) + ", " + err.Error())
			response.GetErrorResponse("Error updating one object in MsSql because no match was found!" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully updating one object in MsSql "
			request.Log(response.Message)
		}

	}

	session.Close()
	return response
}

func (repository MssqlRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-MULTIPLE")
	response := RepositoryResponse{}
	session, isError, errorMessage := getMssqlConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		for _, obj := range request.Body.Objects {
			_, err := session.Query("DELETE FROM [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "] WHERE " + request.Body.Parameters.KeyProperty + " = '" + obj[request.Body.Parameters.KeyProperty].(string) + "'")
			if err != nil {
				response.IsSuccess = false
				request.Log("Error deleting object in MsSql : " + err.Error())
				response.GetErrorResponse("Error deleting one object in MsSql because no match was found!" + err.Error())
			} else {
				response.IsSuccess = true
				response.Message = "Successfully deleted one object in MsSql"
				request.Log(response.Message)
			}
		}
	}

	session.Close()
	return response
}

func (repository MssqlRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-SINGLE")
	response := RepositoryResponse{}
	session, isError, errorMessage := getMssqlConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		_, err := session.Query("DELETE FROM [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "] WHERE " + request.Body.Parameters.KeyProperty + " = '" + request.Controls.Id + "'")
		if err != nil {
			response.IsSuccess = false
			request.Log("Error deleting object in MsSql : " + err.Error())
			response.GetErrorResponse("Error deleting one object in MsSql because no match was found!" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully deleted one object in MsSql"
			request.Log(response.Message)
		}
	}

	session.Close()
	return response
}

func (repository MssqlRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	request.Log("Starting SPECIAL!")
	queryType := request.Body.Special.Type

	switch queryType {
	case "getFields":
		request.Log("Starting GET-FIELDS sub routine!")
		fieldsInByte := executeMssqlGetFields(request)
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
		fieldsInByte := executeMssqlGetClasses(request)
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
		fieldsInByte := executeMssqlGetNamespaces(request)
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
		fieldsInByte := executeMssqlGetSelected(request)
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
		//default:
		//return search(request, request.Body.Special.Parameters)
	}

	return response
	//request.Log("Special not implemented in MsSql Db repository")
	//eturn getDefaultNotImplemented()
}

func (repository MssqlRepository) Test(request *messaging.ObjectRequest) {

}

//Sub Routines

func executeMssqlGetFields(request *messaging.ObjectRequest) (returnByte []byte) {
	session, isError, _ := getMssqlConnection(request)
	if isError == true {
		returnByte = nil
	} else {
		isError = false

		var returnMap map[string]interface{}
		returnMap = make(map[string]interface{})

		rows, err := session.Query("SELECT COLUMN_NAME FROM [" + getMssqlSQLnamespace(request) + "].INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = '" + request.Controls.Class + "'")

		if err != nil {
			request.Log("Error executing query in Mssql SQL")
		} else {
			request.Log("Successfully executed query in MsSql SQL")
			columns, _ := rows.Columns()
			count := len(columns)
			values := make([]interface{}, count)
			valuePtrs := make([]interface{}, count)

			index := 0
			for rows.Next() {

				var tempMap map[string]interface{}
				tempMap = make(map[string]interface{})

				for i, _ := range columns {
					valuePtrs[i] = &values[i]
				}

				rows.Scan(valuePtrs...)

				for i, col := range columns {

					var v interface{}

					val := values[i]

					b, ok := val.([]byte)

					if ok {
						v = string(b)
					} else {
						v = val
					}

					tempMap[col] = v

				}

				returnMap[strconv.Itoa(index)] = tempMap["COLUMN_NAME"]
				index++
			}

			var FieldArray []string
			FieldArray = make([]string, len(returnMap))

			for key, value := range returnMap {
				index, _ := strconv.Atoi(key)
				FieldArray[index] = value.(string)
			}

			byteValue, errMarshal := json.Marshal(FieldArray)
			if errMarshal != nil {
				request.Log("Error converting to byte array")
				byteValue = nil
			} else {
				request.Log("Successfully converted result to byte array")
			}

			returnByte = byteValue
		}

	}
	session.Close()
	return returnByte
}

func executeMssqlGetClasses(request *messaging.ObjectRequest) (returnByte []byte) {
	session, isError, _ := getMssqlConnection(request)
	if isError == true {
		returnByte = nil
	} else {
		isError = false

		var returnMap map[string]interface{}
		returnMap = make(map[string]interface{})

		rows, err := session.Query("SELECT TABLE_NAME FROM [" + getMssqlSQLnamespace(request) + "].INFORMATION_SCHEMA.TABLES;")

		if err != nil {
			request.Log("Error executing query in MsSql SQL")
		} else {
			request.Log("Successfully executed query in MsSql SQL")
			columns, _ := rows.Columns()
			count := len(columns)
			values := make([]interface{}, count)
			valuePtrs := make([]interface{}, count)

			index := 0
			for rows.Next() {

				var tempMap map[string]interface{}
				tempMap = make(map[string]interface{})

				for i, _ := range columns {
					valuePtrs[i] = &values[i]
				}

				rows.Scan(valuePtrs...)

				for i, col := range columns {

					var v interface{}

					val := values[i]

					b, ok := val.([]byte)

					if ok {
						v = string(b)
					} else {
						v = val
					}

					tempMap[col] = v

				}

				returnMap[strconv.Itoa(index)] = tempMap["TABLE_NAME"]
				index++
			}

			var FieldArray []string
			FieldArray = make([]string, len(returnMap))

			for key, value := range returnMap {
				index, _ := strconv.Atoi(key)
				FieldArray[index] = value.(string)
			}

			byteValue, errMarshal := json.Marshal(FieldArray)
			if errMarshal != nil {
				request.Log("Error converting to byte array")
				byteValue = nil
			} else {
				request.Log("Successfully converted result to byte array")
			}

			returnByte = byteValue
		}

	}

	session.Close()
	return returnByte
}

func executeMssqlGetNamespaces(request *messaging.ObjectRequest) (returnByte []byte) {
	session, isError, _ := getMssqlConnection(request)
	if isError == true {
		returnByte = nil
	} else {
		isError = false

		var returnMap map[string]interface{}
		returnMap = make(map[string]interface{})

		rows, err := session.Query("SELECT name FROM sys.databases;")

		if err != nil {
			request.Log("Error executing query in MsSql SQL")
		} else {
			request.Log("Successfully executed query in MsSql SQL")
			columns, _ := rows.Columns()
			count := len(columns)
			values := make([]interface{}, count)
			valuePtrs := make([]interface{}, count)

			index := 0
			for rows.Next() {

				var tempMap map[string]interface{}
				tempMap = make(map[string]interface{})

				for i, _ := range columns {
					valuePtrs[i] = &values[i]
				}

				rows.Scan(valuePtrs...)

				for i, col := range columns {

					var v interface{}

					val := values[i]

					b, ok := val.([]byte)

					if ok {
						v = string(b)
					} else {
						v = val
					}

					tempMap[col] = v

				}

				returnMap[strconv.Itoa(index)] = tempMap["name"]
				index++
			}

			var FieldArray []string
			FieldArray = make([]string, len(returnMap))

			for key, value := range returnMap {
				index, _ := strconv.Atoi(key)
				FieldArray[index] = value.(string)
			}

			byteValue, errMarshal := json.Marshal(FieldArray)
			if errMarshal != nil {
				request.Log("Error converting to byte array")
				byteValue = nil
			} else {
				request.Log("Successfully converted result to byte array")
			}

			returnByte = byteValue
		}

	}

	session.Close()
	return returnByte
}

func executeMssqlGetSelected(request *messaging.ObjectRequest) (returnByte []byte) {

	session, isError, _ := getMssqlConnection(request)
	if isError == true {
		request.Log("Error Connecting to MsSql")
	} else {
		var returnMap map[string]interface{}
		returnMap = make(map[string]interface{})

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
		request.Log("SELECT " + selectedItemsQuery + " FROM " + request.Controls.Class)
		rows, err := session.Query("SELECT " + selectedItemsQuery + " FROM [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "]")

		if err != nil {
			request.Log("Error Fetching data from MsSql")
		} else {
			request.Log("Successfully fetched data from MsSql")
			columns, _ := rows.Columns()
			count := len(columns)
			values := make([]interface{}, count)
			valuePtrs := make([]interface{}, count)

			index := 0
			for rows.Next() {

				var tempMap map[string]interface{}
				tempMap = make(map[string]interface{})

				for i, _ := range columns {
					valuePtrs[i] = &values[i]
				}

				rows.Scan(valuePtrs...)

				for i, col := range columns {

					var v interface{}

					val := values[i]

					b, ok := val.([]byte)

					if ok {
						v = string(b)
					} else {
						v = val
					}
					tempMap[col] = v
				}

				returnMap[strconv.Itoa(index)] = tempMap
				index++
			}

			byteValue, _ := json.Marshal(returnMap)
			returnByte = byteValue
		}

	}

	session.Close()
	return returnByte
}

func getMssqlSqlRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
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
		//request.Log("GUID Key generation requested!")
		returnID = uuid.NewV1().String()
	} else if isAutoIncrementId {
		//request.Log("Automatic Increment Key generation requested!")
		session, isError, _ := getMssqlConnection(request)
		if isError {
			returnID = ""
			request.Log("Connecting to MySQL Failed!")
		} else {
			//Read Table domainClassAttributes
			request.Log("Reading maxCount from DB")
			request.Log("SELECT maxCount FROM [" + getMssqlSQLnamespace(request) + "].[dbo].[domainClassAttributes] where class = '" + request.Controls.Class + "';")
			rows, err := session.Query("SELECT maxCount FROM [" + getMssqlSQLnamespace(request) + "].[dbo].[domainClassAttributes] where class = '" + request.Controls.Class + "';")

			if err != nil {
				//If err create new domainClassAttributes  table
				request.Log("No Class found.. Must be a new namespace")
				request.Log("create table [" + getMssqlSQLnamespace(request) + "].[dbo].[domainClassAttributes] ( class varchar(255) primary key, maxcount text, version text);")
				_, err = session.Query("create table [" + getMssqlSQLnamespace(request) + "].[dbo].[domainClassAttributes] ( class varchar(255) primary key, maxcount text, version text);")
				if err != nil {
					request.Log("Error : " + err.Error())
					returnID = ""
					return
				} else {
					//insert record with count 1 and return
					request.Log("INSERT INTO [" + getMssqlSQLnamespace(request) + "].[dbo].[domainClassAttributes] (class, maxcount,version) VALUES ('" + request.Controls.Class + "','1','" + uuid.NewV1().String() + "')")
					_, err := session.Query("INSERT INTO [" + getMssqlSQLnamespace(request) + "].[dbo].[domainClassAttributes] (class, maxcount,version) VALUES ('" + request.Controls.Class + "','1','" + uuid.NewV1().String() + "')")
					if err != nil {
						request.Log("Error : " + err.Error())
						returnID = ""
						return
					} else {
						returnID = "1"
						return
					}
				}
			} else {
				//read value
				var myMap map[string]interface{}
				myMap = make(map[string]interface{})

				columns, _ := rows.Columns()
				count := len(columns)
				values := make([]interface{}, count)
				valuePtrs := make([]interface{}, count)

				for rows.Next() {
					for i, _ := range columns {
						valuePtrs[i] = &values[i]
					}

					rows.Scan(valuePtrs...)

					for i, col := range columns {

						var v interface{}

						val := values[i]

						b, ok := val.([]byte)

						if ok {
							v = string(b)
						} else {
							v = val
						}

						myMap[col] = v
					}
				}

				if len(myMap) == 0 {
					request.Log("New Class! New record for this class will be inserted")
					request.Log("INSERT INTO [" + getMssqlSQLnamespace(request) + "].[dbo].[domainClassAttributes] (class,maxcount,version) values ('" + request.Controls.Class + "', '1', '" + uuid.NewV1().String() + "');")
					_, err = session.Query("INSERT INTO [" + getMssqlSQLnamespace(request) + "].[dbo].[domainClassAttributes] (class,maxcount,version) values ('" + request.Controls.Class + "', '1', '" + uuid.NewV1().String() + "');")
					if err != nil {
						request.Log("Error : " + err.Error())
						returnID = ""
						return
					} else {
						returnID = "1"
						return
					}
				} else {
					//inrement one and UPDATE
					request.Log("Record Available!")
					maxCount := 0
					maxCount, err = strconv.Atoi(myMap["maxCount"].(string))
					maxCount++
					returnID = strconv.Itoa(maxCount)
					request.Log("UPDATE [" + getMssqlSQLnamespace(request) + "].[dbo].[domainClassAttributes] SET maxcount='" + returnID + "' WHERE class = '" + request.Controls.Class + "' ;")
					_, err = session.Query("UPDATE [" + getMssqlSQLnamespace(request) + "].[dbo].[domainClassAttributes] SET maxcount='" + returnID + "' WHERE class = '" + request.Controls.Class + "' ;")
					if err != nil {
						request.Log("Error Updating index table : " + err.Error())
						returnID = ""
						return
					}
				}
			}
		}
		session.Close()
	} else {
		//	request.Log("Manual Key requested!")
		if obj == nil {
			returnID = request.Controls.Id
		} else {
			returnID = obj[request.Body.Parameters.KeyProperty].(string)
		}
	}

	return
}

func getMssqlDataType(item interface{}) (datatype string) {
	datatype = reflect.TypeOf(item).Name()
	if datatype == "bool" {
		datatype = "bit"
	} else if datatype == "float64" {
		datatype = "real"
	} else if datatype == "string" {
		datatype = "varchar(max)"
	} else if datatype == "" || datatype == "ControlHeaders" {
		datatype = "text"
	}
	return datatype
}
