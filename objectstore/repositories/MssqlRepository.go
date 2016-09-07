package repositories

import (
	//"bytes"
	"database/sql"
	"duov6.com/objectstore/messaging"
	"encoding/binary"
	"encoding/json"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/twinj/uuid"
	"math"
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

func Float64frombytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
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

		take := "100"
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

					//b, ok := val.([]byte)

					//if ok {
					//	v = string(b)
					//} else {
					v = val
					//}

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
					if col == "osheaders" || col == "r_n_n" {
						//Do Nothing :D
					} else {

						var v interface{}

						val := values[i]

						//b, ok := val.([]byte)
						//if ok {
						//	v = string(b)
						//} else {
						//v = val
						//}
						v = repository.sqlToGolang(val)
						tempMap[col] = v
					}
				}

				returnMap = append(returnMap, tempMap)

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

func (repository MssqlRepository) sqlToGolang(input interface{}) (output interface{}) {
	if input == nil {
		output = input
	} else {
		dataType := reflect.TypeOf(input).String()

		switch dataType {
		case "string":
			output = strings.TrimSpace(input.(string))
			break
		case "[]byte":
			b, ok := input.([]byte)
			if ok {
				output = string(b)
			}
			break
		case "[]uint8":
			if f64, err2 := strconv.ParseFloat(string(input.([]byte)), 64); err2 == nil {
				output = f64
			} else if f32, err2 := strconv.ParseFloat(string(input.([]byte)), 32); err2 == nil {
				output = f32
			} else if i64, err2 := strconv.ParseInt(string(input.([]byte)), 10, 64); err2 == nil {
				output = i64
			} else if i32, err2 := strconv.ParseInt(string(input.([]byte)), 10, 32); err2 == nil {
				output = i32
			} else {
				output = string(input.([]byte))
			}
			break
		default:
			output = input
			break
		}
	}
	return
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
			fieldsInByte := repository.executeMssqlQuery(request)
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

func (repository MssqlRepository) executeMssqlQuery(request *messaging.ObjectRequest) (returnByte []byte) {
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

					//b, ok := val.([]byte)

					//if ok {
					//	v = string(b)
					//} else {
					v = repository.sqlToGolang(val)
					//}

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

						//b, ok := val.([]byte)

						//if ok {
						//	v = string(b)
						//} else {
						v = val
						//}

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

					//b, ok := val.([]byte)

					//if ok {
					//	v = string(b)
					//} else {
					v = repository.sqlToGolang(val)
					//}

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
		fieldsInByte := repository.executeMssqlGetSelected(request)
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

					//b, ok := val.([]byte)

					//if ok {
					//	v = string(b)
					//} else {
					v = val
					//}

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

					//b, ok := val.([]byte)

					//if ok {
					//	v = string(b)
					//} else {
					v = val
					//}

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

					//b, ok := val.([]byte)

					//if ok {
					//	v = string(b)
					//} else {
					v = val
					//}

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

func (repository MssqlRepository) executeMssqlGetSelected(request *messaging.ObjectRequest) (returnByte []byte) {

	session, isError, _ := getMssqlConnection(request)
	if isError == true {
		request.Log("Error Connecting to MsSql")
	} else {
		var data []interface{}

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

					//b, ok := val.([]byte)

					//if ok {
					//	v = string(b)
					//} else {
					v = repository.sqlToGolang(val)
					//}
					tempMap[col] = v
				}
				data = append(data, tempMap)
				index++
			}

			byteValue, _ := json.Marshal(data)
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

func (repository MssqlRepository) ClearCache(request *messaging.ObjectRequest) {
}

//Reformed code for MsSQL - Work in Progress
/*
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
	_ "github.com/denisenkom/go-mssqldb"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MssqlRepository struct {
}

//var availableDbs map[string]interface{}
//var availableTables map[string]interface{}
//var tableCache map[string]map[string]string
//var connection map[string]*sql.DB

var msSqlTableCache map[string]map[string]string
var msSqlTableCacheLock = sync.RWMutex{}

var msSqlAvailableTables map[string]interface{}
var msSqlAvailableTablesLock = sync.RWMutex{}

var msSqlAvailableDbs map[string]interface{}
var msSqlAvailableDbsLock = sync.RWMutex{}

var msSqlConnections map[string]*sql.DB
var msSqlConnectionLock = sync.RWMutex{}

func (repository MssqlRepository) GetRepositoryName() string {
	return "MsSQL"
}

// Start of GET and SET methods

func (repository MssqlRepository) GetMsSqlConnections(index string) (conn *sql.DB) {
	msSqlConnectionLock.RLock()
	defer msSqlConnectionLock.RUnlock()
	conn = msSqlConnections[index]
	return
}

func (repository MssqlRepository) SetMsSqlConnections(index string, conn *sql.DB) {
	msSqlConnectionLock.Lock()
	defer msSqlConnectionLock.Unlock()
	msSqlConnections[index] = conn
}

func (repository MssqlRepository) GetMsSqlAvailableDbs(index string) (value interface{}) {
	msSqlAvailableDbsLock.RLock()
	defer msSqlAvailableDbsLock.RUnlock()
	value = msSqlAvailableDbs[index]
	return
}

func (repository MssqlRepository) SetMsSqlAvailabaleDbs(index string, value interface{}) {
	msSqlAvailableDbsLock.Lock()
	defer msSqlAvailableDbsLock.Unlock()
	msSqlAvailableDbs[index] = value
}

func (repository MssqlRepository) GetMsSqlAvailableTables(index string) (value interface{}) {
	msSqlAvailableTablesLock.RLock()
	defer msSqlAvailableTablesLock.RUnlock()
	value = msSqlAvailableTables[index]
	return
}

func (repository MssqlRepository) SetMsSqlAvailabaleTables(index string, value interface{}) {
	msSqlAvailableTablesLock.Lock()
	defer msSqlAvailableTablesLock.Unlock()
	msSqlAvailableTables[index] = value
}

func (repository MssqlRepository) GetMsSqlTableCache(index string) (value map[string]string) {
	msSqlTableCacheLock.RLock()
	defer msSqlTableCacheLock.RUnlock()
	value = msSqlTableCache[index]
	return
}

func (repository MssqlRepository) SetMsSqlTableCache(index string, value map[string]string) {
	msSqlTableCacheLock.Lock()
	defer msSqlTableCacheLock.Unlock()
	msSqlTableCache[index] = value
}

// End of GET and SET methods

func (repository MssqlRepository) GetConnection(request *messaging.ObjectRequest) (conn *sql.DB, err error) {

	if msSqlConnections == nil {
		msSqlConnections = make(map[string]*sql.DB)
	}

	mysqlConf := request.Configuration.ServerConfiguration["MSSQL"]

	username := mysqlConf["Username"]
	password := mysqlConf["Password"]
	url := mysqlConf["Server"]
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

	if repository.GetMsSqlConnections(poolPattern) == nil {
		conn, err = repository.CreateConnection(username, password, url, port, IdleLimit, OpenLimit, TTL)
		if err != nil {
			request.Log("Error : " + err.Error())
			return
		}
		repository.SetMsSqlConnections(poolPattern, conn)
	} else {
		if repository.GetMsSqlConnections(poolPattern).Ping(); err != nil {
			_ = repository.GetMsSqlConnections(poolPattern).Close()
			repository.SetMsSqlConnections(poolPattern, nil)
			conn, err = repository.CreateConnection(username, password, url, port, IdleLimit, OpenLimit, TTL)
			if err != nil {
				request.Log("Error : " + err.Error())
				return
			}
			repository.SetMsSqlConnections(poolPattern, conn)
		} else {
			conn = repository.GetMsSqlConnections(poolPattern)
		}
	}
	return conn, err
}

func (repository MssqlRepository) CreateConnection(username, password, url, port string, IdleLimit, OpenLimit, TTL int) (conn *sql.DB, err error) {
	conn, err = sql.Open("mssql", "server="+url+";port="+port+";user id="+username+";password="+password+";encrypt=disable")
	conn.SetMaxIdleConns(IdleLimit)
	conn.SetMaxOpenConns(OpenLimit)
	conn.SetConnMaxLifetime(time.Duration(TTL) * time.Minute)
	return
}

func (repository MssqlRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
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
	rowss, err := session.Query("select top " + take + " * from (select *, ROW_NUMBER() over (order by " + fieldName + ") as r_n_n from [" + getMssqlSQLnamespace(request) + "].[dbo].[" + request.Controls.Class + "]) xx where r_n_n >=" + skip + ";")

	query := "SELECT * FROM [" + repository.GetDatabaseName(request.Controls.Namespace) + "].[dbo].[" + request.Controls.Class + "] "

	if isOrderByAsc {
		query += " order by " + orderbyfield + " asc "
	} else if isOrderByDesc {
		query += " order by " + orderbyfield + " desc "
	}

	query += " OFFSET " + skip + " ROWS "
	query += " FETCH NEXT " + take + " ROWS ONLY;"

	response := repository.queryCommonMany(query, request)
	return response
}

func (repository MssqlRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
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

		formattedQuery, err := queryparser.GetMsSQLQuery(request.Body.Query.Parameters, request.Controls.Namespace, request.Controls.Class, parameters)
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

func (repository MssqlRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	if security.ValidateSecurity(getNoSqlKey(request)) {
		response.GetResponseWithBody(getEmptyByteObject())
		request.Log("Error! Security Violation of request detected. Aborting request with error!")
		return response
	}

	query := "SELECT * FROM [" + repository.GetDatabaseName(request.Controls.Namespace) + "].[dbo].[" + request.Controls.Class + "] WHERE __os_id = '" + getNoSqlKey(request) + "';"
	response = repository.queryCommonOne(query, request)
	return response
}

func (repository MssqlRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
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

	domain := repository.GetDatabaseName(request.Controls.Namespace)

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
			query = "select * from [" + domain + "].[dbo].[" + request.Controls.Class + "] where " + fieldName + " LIKE '%" + fieldValue + "%'"
		} else if strings.HasPrefix(fieldValue, "*") {
			fieldValue = strings.TrimPrefix(fieldValue, "*")
			query = "select * from [" + domain + "].[dbo].[" + request.Controls.Class + "] where " + fieldName + " LIKE '%" + fieldValue + "'"
		} else if strings.HasSuffix(fieldValue, "*") {
			fieldValue = strings.TrimSuffix(fieldValue, "*")
			query = "select * from [" + domain + "].[dbo].[" + request.Controls.Class + "] where " + fieldName + " LIKE '" + fieldValue + "%'"
		} else {
			query = "select * from [" + domain + "].[dbo].[" + request.Controls.Class + "] where " + fieldName + "='" + fieldValue + "'"
		}
	} else {
		if request.Body.Query.Parameters == "" || request.Body.Query.Parameters == "*" {
			//Get All Query
			query = "select * from [" + domain + "].[dbo].[" + request.Controls.Class + "]"
		} else {
			//Full Text Search Query
			query = repository.GetFullTextSearchQuery(request)
			isFullTextSearch = true
		}
	}

	if !isFullTextSearch {
		if isOrderByAsc {
			query += " order by " + orderbyfield + " asc "
		} else if isOrderByDesc {
			query += " order by " + orderbyfield + " desc "
		}

		query += " OFFSET " + take + " ROWS "
		query += " FETCH NEXT " + skip + " ROWS ONLY;"
	}

	response = repository.queryCommonMany(query, request)
	return response
}

func (repository MssqlRepository) GetFullTextSearchQuery(request *messaging.ObjectRequest) (query string) {
	var fieldNames []string

	domain := repository.GetDatabaseName(request.Controls.Namespace)

	indexedFields := repository.GetFullTextIndexes(request)

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

		localTableCacheEntry := repository.GetMsSqlTableCache(domain + "." + request.Controls.Class)

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
		} else if localTableCacheEntry != nil {
			//Available in Table Cache
			for name, fieldType := range localTableCacheEntry {
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
		query = "SELECT * FROM " + repository.GetDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class + " WHERE Concat("

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

func (repository MssqlRepository) GetFullTextIndexes(request *messaging.ObjectRequest) (fieldnames []string) {

	conn, err := repository.GetConnection(request)
	if err != nil {
		request.Log("Error : " + err.Error())
		return
	}

	domain := repository.GetDatabaseName(request.Controls.Namespace)
	getIndexesQuery := "show index from " + domain + "." + request.Controls.Class + " where Index_type = 'FULLTEXT'"

	indexResult, err := repository.ExecuteQueryMany(request, conn, getIndexesQuery, "")
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

func (repository MssqlRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {

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

func (repository MssqlRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {

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

func (repository MssqlRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {

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

func (repository MssqlRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {

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

func (repository MssqlRepository) ReRun(request *messaging.ObjectRequest, conn *sql.DB, obj map[string]interface{}) RepositoryResponse {
	var response RepositoryResponse

	repository.CheckSchema(request, conn, request.Controls.Namespace, request.Controls.Class, obj)
	response = repository.queryStore(request)
	if !response.IsSuccess {
		if CheckRedisAvailability(request) {
			cache.FlushCache(request)
		} else {
			msSqlTableCache = make(map[string]map[string]string)
			msSqlAvailableDbs = make(map[string]interface{})
			msSqlAvailableTables = make(map[string]interface{})
		}
		repository.CheckSchema(request, conn, request.Controls.Namespace, request.Controls.Class, obj)
		response = repository.queryStore(request)
	}

	return response
}

func (repository MssqlRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {

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

func (repository MssqlRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {

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

func (repository MssqlRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {

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
		//fieldNameList := make([]string, 0)
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

		// query = "EXPLAIN " + domain + "." + request.Controls.Class + ";"
		// var resultSet2 []map[string]interface{}
		// repoResponse := repository.queryCommonMany(query, request)
		// err := json.Unmarshal(repoResponse.Body, &resultSet2)
		// if err != nil {
		// 	isError = true
		// } else {
		// 	if len(resultSet2) > 0 {
		// 		for x := 0; x < len(resultSet2); x++ {
		// 			fieldNameList = append(fieldNameList, resultSet2[x]["Field"].(string))
		// 		}
		// 	}
		// }

		if isError {
			response.IsSuccess = false
		} else {
			response.IsSuccess = true
			returnMap := make(map[string]interface{})
			returnMap["RecordCount"] = recordCount
			//returnMap["FieldList"] = fieldNameList
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
					_ = cache.DeleteKey(request, ("CloudSqlTableCache." + domain + "." + request.Controls.Class), cache.MetaData)
					_ = cache.DeleteKey(request, ("CloudSqlAvailableTables." + domain + "." + request.Controls.Class), cache.MetaData)
					_ = cache.DeletePattern(request, (domain + "." + request.Controls.Class + "*"), cache.Data)
				} else {
					delete(msSqlAvailableTables, (domain + "." + request.Controls.Class))
					delete(msSqlTableCache, (domain + "." + request.Controls.Class))
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
					delete(msSqlAvailableDbs, domain)
					//Delete all associated Classes from it's TableCache and availableTables
					for key, _ := range msSqlAvailableTables {
						if strings.Contains(key, domain) {
							delete(msSqlAvailableTables, key)
							delete(msSqlTableCache, key)
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
			msSqlTableCache = make(map[string]map[string]string)
			msSqlAvailableDbs = make(map[string]interface{})
			msSqlAvailableTables = make(map[string]interface{})
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

		conn, err := repository.GetConnection(request)
		if err != nil {
			request.Log("Error : " + err.Error())
			response.IsSuccess = false
			response.Message = "Connection Error! : " + err.Error()
			return response
		}

		err = repository.CheckAvailabilityDb(request, conn, domain)
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

		conn, err := repository.GetConnection(request)
		if err != nil {
			response.IsSuccess = false
			response.Message = err.Error()
			return response
		}

		switch FTH_command {
		case "reset":
			getIndexesQuery := "show index from " + domain + "." + request.Controls.Class + " where Index_type = 'FULLTEXT'"

			indexResult, err := repository.ExecuteQueryMany(request, conn, getIndexesQuery, "")
			if err != nil {
				request.Log("Error : " + err.Error())
			} else {
				executedList := ""
				for _, obj := range indexResult {
					keyName := obj["Key_name"].(string)
					if !strings.Contains(executedList, keyName) {
						alterQuery := "ALTER TABLE " + domain + "." + request.Controls.Class + " DROP INDEX " + keyName
						_, _ = repository.ExecuteNonQuery(conn, alterQuery, request)
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
			err, _ = repository.ExecuteNonQuery(conn, alterQuery, request)
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

		conn, err := repository.GetConnection(request)
		if err != nil {
			response.IsSuccess = false
			response.Message = err.Error()
			return response
		}

		switch UIC_command {
		case "reset":
			getIndexesQuery := "show index from " + domain + "." + request.Controls.Class + " where Index_type = 'BTREE' AND Key_name != 'PRIMARY';"

			indexResult, err := repository.ExecuteQueryMany(request, conn, getIndexesQuery, "")
			if err != nil {
				request.Log("Error : " + err.Error())
			} else {
				executedList := ""
				for _, obj := range indexResult {
					keyName := obj["Key_name"].(string)
					if !strings.Contains(executedList, keyName) {
						alterQuery := "ALTER TABLE " + domain + "." + request.Controls.Class + " DROP INDEX " + keyName
						_, _ = repository.ExecuteNonQuery(conn, alterQuery, request)
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
				err, _ = repository.ExecuteNonQuery(conn, alterQuery, request)
				if err != nil {
					//1170 - Non defined key length for indexable field
					if strings.Contains(err.Error(), "BLOB/TEXT") {
						modifyQuery := "ALTER TABLE " + domain + "." + request.Controls.Class + " MODIFY COLUMN " + singleName + " varchar(255);"
						err, _ = repository.ExecuteNonQuery(conn, modifyQuery, request)
						if err != nil {
							request.Log("Error : " + err.Error())
						} else {
							err, _ = repository.ExecuteNonQuery(conn, alterQuery, request)
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

func (repository MssqlRepository) Test(request *messaging.ObjectRequest) {
}

func (repository MssqlRepository) ClearCache(request *messaging.ObjectRequest) {
	if CheckRedisAvailability(request) {
		cache.FlushCache(request)
	} else {
		msSqlTableCache = make(map[string]map[string]string)
		msSqlAvailableDbs = make(map[string]interface{})
		msSqlAvailableTables = make(map[string]interface{})
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////SQL GENERATORS/////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////
func (repository MssqlRepository) queryCommon(query string, request *messaging.ObjectRequest, isOne bool) RepositoryResponse {
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

func (repository MssqlRepository) queryCommonMany(query string, request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryCommon(query, request, false)

}

func (repository MssqlRepository) queryCommonOne(query string, request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryCommon(query, request, true)
}

func (repository MssqlRepository) queryStore(request *messaging.ObjectRequest) RepositoryResponse {
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

func (repository MssqlRepository) getByKey(conn *sql.DB, namespace string, class string, id string, request *messaging.ObjectRequest) (obj map[string]interface{}) {

	isCacheable := false
	if request != nil {
		if CheckRedisAvailability(request) {
			isCacheable = true
		}
	}

	if isCacheable {
		result := cache.GetByKey(request, cache.Data)
		if result == nil {
			query := "SELECT * FROM " + repository.GetDatabaseName(namespace) + "." + class + " WHERE __os_id = '" + id + "';"
			obj, _ = repository.ExecuteQueryOne(request, conn, query, nil)
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
		query := "SELECT * FROM " + repository.GetDatabaseName(namespace) + "." + class + " WHERE __os_id = '" + id + "';"
		obj, _ = repository.ExecuteQueryOne(request, conn, query, nil)
	}

	return
}

func (repository MssqlRepository) GetMultipleStoreScripts(conn *sql.DB, request *messaging.ObjectRequest) (query []map[string]interface{}, err error) {
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

func (repository MssqlRepository) GetMultipleInsertQuery(request *messaging.ObjectRequest, namespace, class string, records []map[string]interface{}, conn *sql.DB) (queryData map[string]interface{}) {
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

func (repository MssqlRepository) GetSingleObjectInsertQuery(request *messaging.ObjectRequest, namespace, class string, obj map[string]interface{}, conn *sql.DB) (query string) {
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

	query += "(__os_id" + keyList + ") VALUES "
	query += ("(\"" + id + "\"" + valueList + ")")
	return
}

func (repository MssqlRepository) GetSingleObjectUpdateQuery(request *messaging.ObjectRequest, namespace, class string, obj map[string]interface{}, conn *sql.DB) (query string) {

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
	query = ("UPDATE " + repository.GetDatabaseName(namespace) + "." + class + " SET " + updateValues + " WHERE __os_id=\"" + getNoSqlKeyById(request, obj) + "\";")
	return
}

func (repository MssqlRepository) GetDeleteScript(namespace string, class string, id string) string {
	return "DELETE FROM " + repository.GetDatabaseName(namespace) + "." + class + " WHERE __os_id = \"" + id + "\""
}

func (repository MssqlRepository) GetCreateScript(namespace string, class string, obj map[string]interface{}) string {

	domain := repository.GetDatabaseName(namespace)

	query := "CREATE TABLE IF NOT EXISTS " + domain + "." + class + "(__os_id varchar(255) primary key"

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

	query += ")"
	return query
}

func (repository MssqlRepository) CheckAvailabilityDb(request *messaging.ObjectRequest, conn *sql.DB, dbName string) (err error) {
	if msSqlAvailableDbs == nil {
		msSqlAvailableDbs = make(map[string]interface{})
	}

	if CheckRedisAvailability(request) {
		if cache.ExistsKeyValue(request, ("CloudSqlAvailableDbs." + dbName), cache.MetaData) {
			return
		}
	} else {
		if repository.GetMsSqlAvailableDbs(dbName) != nil {
			return
		}
	}

	dbQuery := "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '" + dbName + "'"
	dbResult, err := repository.ExecuteQueryOne(request, conn, dbQuery, nil)

	if err == nil {
		if dbResult["SCHEMA_NAME"] == nil {
			repository.ExecuteNonQuery(conn, "CREATE DATABASE IF NOT EXISTS "+dbName, request)
			repository.ExecuteNonQuery(conn, "create table "+dbName+".domainClassAttributes ( __os_id VARCHAR(255) primary key, class text, maxCount text, version text);", request)
		}

		if CheckRedisAvailability(request) {
			err = cache.StoreKeyValue(request, ("CloudSqlAvailableDbs." + dbName), "true", cache.MetaData)
		} else {
			if repository.GetMsSqlAvailableDbs(dbName) == nil {
				repository.SetMsSqlAvailabaleDbs(dbName, true)
			}
		}
	} else {
		request.Log("Error : " + err.Error())
	}

	return
}

func (repository MssqlRepository) CheckAvailabilityTable(request *messaging.ObjectRequest, conn *sql.DB, dbName string, namespace string, class string, obj map[string]interface{}) (err error) {

	if msSqlAvailableTables == nil {
		msSqlAvailableTables = make(map[string]interface{})
	}

	isTableCreatedNow := false

	if CheckRedisAvailability(request) {
		if !cache.ExistsKeyValue(request, ("CloudSqlAvailableTables." + dbName + "." + class), cache.MetaData) {
			var tableResult map[string]interface{}
			tableResult, err = repository.ExecuteQueryOne(request, conn, "SHOW TABLES FROM "+dbName+" LIKE \""+class+"\"", nil)
			if err == nil {
				//if tableResult["Tables_in_"+dbName] == nil {
				if len(tableResult) == 0 {
					script := repository.GetCreateScript(namespace, class, obj)
					err, _ = repository.ExecuteNonQuery(conn, script, request)
					if err != nil {
						return
					} else {
						isTableCreatedNow = true
						recordForIDService := "INSERT INTO " + dbName + ".domainClassAttributes (__os_id, class, maxCount,version) VALUES ('" + getDomainClassAttributesKey(request) + "','" + request.Controls.Class + "','0','" + common.GetGUID() + "')"
						_, _ = repository.ExecuteNonQuery(conn, recordForIDService, request)
						keygenerator.CreateNewKeyGenBundle(request)
					}
				}
				if CheckRedisAvailability(request) {
					err = cache.StoreKeyValue(request, ("CloudSqlAvailableTables." + dbName + "." + class), "true", cache.MetaData)
				} else {
					keyword := dbName + "." + class
					availableTableValue := repository.GetMsSqlAvailableTables(keyword)
					if availableTableValue == nil || availableTableValue.(bool) == false {
						repository.SetMsSqlAvailabaleTables(keyword, true)
					}
				}

			} else {
				return
			}
		}
	} else {
		keyword := dbName + "." + class
		availableTableValue := repository.GetMsSqlAvailableTables(keyword)
		if availableTableValue == nil {
			var tableResult map[string]interface{}
			tableResult, err = repository.ExecuteQueryOne(request, conn, "SHOW TABLES FROM "+dbName+" LIKE \""+class+"\"", nil)
			if err == nil {
				if tableResult["Tables_in_"+dbName] == nil {
					script := repository.GetCreateScript(namespace, class, obj)
					err, _ = repository.ExecuteNonQuery(conn, script, request)
					if err != nil {
						return
					} else {
						isTableCreatedNow = true
						recordForIDService := "INSERT INTO " + dbName + ".domainClassAttributes (__os_id, class, maxCount,version) VALUES ('" + getDomainClassAttributesKey(request) + "','" + request.Controls.Class + "','0','" + common.GetGUID() + "')"
						_, _ = repository.ExecuteNonQuery(conn, recordForIDService, request)
					}
				}
				if availableTableValue == nil || availableTableValue.(bool) == false {
					repository.SetMsSqlAvailabaleTables(keyword, true)
				}

			} else {
				return
			}
		}
	}

	err = repository.BuildTableCache(request, conn, dbName, class)

	if !isTableCreatedNow {
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
			cacheItem = repository.GetMsSqlTableCache(dbName + "." + class)
		}

		isFirst := true
		for k, v := range obj {
			if !strings.EqualFold(k, "OriginalIndex") || !strings.EqualFold(k, "__osHeaders") {
				_, ok := cacheItem[k]
				if !ok {
					if isFirst {
						isFirst = false
					} else {
					}

					alterQuery := "ALTER TABLE [" + dbName + "].[dbo].[" + class + "] ADD " + k + " " + repository.GolangToSql(v)
					err, _ = repository.ExecuteNonQuery(conn, alterQuery, request)
					if err != nil {
						request.Log("Error : " + err.Error())
					}

					repository.AddColumnToTableCache(request, dbName, class, k, repository.GolangToSql(v))
					cacheItem[k] = repository.GolangToSql(v)
				}
			}
		}

	}

	return
}

func (repository MssqlRepository) AddColumnToTableCache(request *messaging.ObjectRequest, dbName string, class string, field string, datatype string) {
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
		dataMap = repository.GetMsSqlTableCache(dbName + "." + class)
		dataMap[field] = datatype
		repository.SetMsSqlTableCache(dbName+"."+class, dataMap)
	}
}

func (repository MssqlRepository) BuildTableCache(request *messaging.ObjectRequest, conn *sql.DB, dbName string, class string) (err error) {
	if msSqlTableCache == nil {
		msSqlTableCache = make(map[string]map[string]string)
	}

	if !CheckRedisAvailability(request) {
		var ok bool
		tableCacheLocalEntry := repository.GetMsSqlTableCache(dbName + "." + class)
		if tableCacheLocalEntry != nil {
			ok = true
		}

		if !ok {
			var exResult []map[string]interface{}
			exResult, err = repository.ExecuteQueryMany(request, conn, "EXPLAIN "+dbName+"."+class, nil)
			if err == nil {
				newMap := make(map[string]string)

				for _, cRow := range exResult {
					newMap[cRow["Field"].(string)] = cRow["Type"].(string)
				}

				if repository.GetMsSqlTableCache(dbName+"."+class) == nil {
					repository.SetMsSqlTableCache(dbName+"."+class, newMap)
				}
			}
		} else {
			if len(tableCacheLocalEntry) == 0 {
				var exResult []map[string]interface{}
				exResult, err = repository.ExecuteQueryMany(request, conn, "EXPLAIN "+dbName+"."+class, nil)
				if err == nil {
					newMap := make(map[string]string)
					for _, cRow := range exResult {
						newMap[cRow["Field"].(string)] = cRow["Type"].(string)
					}

					repository.SetMsSqlTableCache(dbName+"."+class, newMap)
				}
			}
		}
	} else {
		tableCachePattern := ("CloudSqlTableCache." + dbName + "." + request.Controls.Class)
		IsTableCacheKeys := cache.ExistsKeyValue(request, tableCachePattern, cache.MetaData)
		if !IsTableCacheKeys {
			var exResult []map[string]interface{}
			exResult, err := repository.ExecuteQueryMany(request, conn, "EXPLAIN "+dbName+"."+class, nil)
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
				exResult, err := repository.ExecuteQueryMany(request, conn, "EXPLAIN "+dbName+"."+class, nil)
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

func (repository MssqlRepository) CheckSchema(request *messaging.ObjectRequest, conn *sql.DB, namespace string, class string, obj map[string]interface{}) {
	dbName := repository.GetDatabaseName(namespace)
	err := repository.CheckAvailabilityDb(request, conn, dbName)

	if err == nil {
		err := repository.CheckAvailabilityTable(request, conn, dbName, namespace, class, obj)

		if err != nil {
			request.Log("Error : " + err.Error())
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////Helper functions/////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////

func (repository MssqlRepository) GetDatabaseName(namespace string) string {
	return "_" + strings.ToLower(strings.Replace(namespace, ".", "", -1))
}

func (repository MssqlRepository) GetJson(m interface{}) string {
	bytes, _ := json.Marshal(m)
	return string(bytes[:len(bytes)])
}

func (repository MssqlRepository) GetSqlFieldValue(value interface{}) string {
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
		// else if (strings.Contains(sval, "'")){
		//   		    sEnc := base64.StdEncoding.EncodeToString([]byte(sval))
		//       		strValue = "'^" + sEnc + "'";
		//   		}
		break
	default:
		strValue = "'" + repository.GetJson(v) + "'"
		break

	}

	return strValue
}

func (repository MssqlRepository) GolangToSql(value interface{}) string {

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

func (repository MssqlRepository) SqlToGolang(b []byte, t string) interface{} {

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
			outData = repository.GetInterfaceValue(string(decData))
		} else {
			outData = repository.GetInterfaceValue(tmp)
		}
	}

	return outData
}

func (repository MssqlRepository) GetInterfaceValue(tmp string) (outData interface{}) {
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

func (repository MssqlRepository) RowsToMap(request *messaging.ObjectRequest, rows *sql.Rows, tableName interface{}) (tableMap []map[string]interface{}, err error) {

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
			cacheItem = repository.GetMsSqlTableCache(tName)
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
						v = repository.SqlToGolang(b, t)
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

func (repository MssqlRepository) ExecuteQueryMany(request *messaging.ObjectRequest, conn *sql.DB, query string, tableName interface{}) (result []map[string]interface{}, err error) {
	rows, err := conn.Query(query)

	if err == nil {
		result, err = repository.RowsToMap(request, rows, tableName)
	} else {
		if strings.HasPrefix(err.Error(), "Error 1146") {
			err = nil
			result = make([]map[string]interface{}, 0)
		}
	}

	return
}

func (repository MssqlRepository) ExecuteQueryOne(request *messaging.ObjectRequest, conn *sql.DB, query string, tableName interface{}) (result map[string]interface{}, err error) {
	rows, err := conn.Query(query)

	if err == nil {
		var resultSet []map[string]interface{}
		resultSet, err = repository.RowsToMap(request, rows, tableName)
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

func (repository MssqlRepository) ExecuteNonQuery(conn *sql.DB, query string, request *messaging.ObjectRequest) (err error, message string) {
	request.Log("Debug Query : " + query)
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

func (repository MssqlRepository) GetRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
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

func (repository MssqlRepository) CloseConnection(conn *sql.DB) {
	// err := conn.Close()
	// if err != nil {
	// 	request.Log(err.Error())
	// } else {
	// 	request.Log("Connection Closed!")
	// }
}
*/
