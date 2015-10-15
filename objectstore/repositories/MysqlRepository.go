package repositories

import (
	"database/sql"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/queryparser"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/twinj/uuid"
	"reflect"
	"strconv"
	"strings"
)

type MysqlRepository struct {
}

func (repository MysqlRepository) GetRepositoryName() string {
	return "MYSQL DB"
}

func getMySQLnamespace(request *messaging.ObjectRequest) string {
	namespace := strings.Replace(request.Controls.Namespace, ".", "", -1)
	return "_" + strings.ToLower(namespace)
}

func getMysqlConnection(request *messaging.ObjectRequest) (session *sql.DB, isError bool, errorMessage string) {
	//creating database out of namespace
	isError = false
	server := request.Configuration.ServerConfiguration["MYSQL"]["Url"]
	port := request.Configuration.ServerConfiguration["MYSQL"]["Port"]

	session, err := sql.Open("mysql", request.Configuration.ServerConfiguration["MYSQL"]["Username"]+":"+request.Configuration.ServerConfiguration["MYSQL"]["Password"]+"@tcp("+server+":"+port+")/")

	if err != nil {
		request.Log("Failed to create connection to MySql! : " + err.Error())
	} else {
		request.Log("Successfully created connection to MySql!")
	}

	//Create schema if not available.
	request.Log("Checking if Database " + getMySQLnamespace(request) + " is available.")

	isDatabaseAvailbale := false

	rows, err := session.Query("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '" + getMySQLnamespace(request) + "'")

	if err != nil {
		request.Log("Error contacting Mysql Server to fetch available databases")
	} else {
		request.Log("Successfully retrieved values for all objects in MySQL")

		columns, _ := rows.Columns()
		count := len(columns)
		values := make([]interface{}, count)
		valuePtrs := make([]interface{}, count)

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
				request.Log("Check domain : " + getMySQLnamespace(request) + " : available schema : " + v.(string))
				if v.(string) == getMySQLnamespace(request) {
					//Database available
					isDatabaseAvailbale = true
					break
				}
			}
		}
	}

	if isDatabaseAvailbale {
		request.Log("Database already available. Nothing to do. Proceed!")
	} else {
		_, err = session.Query("create schema " + getMySQLnamespace(request) + ";")
		if err != nil {
			request.Log("Creation of domain matched Schema failed")
		} else {
			request.Log("Creation of domain matched Schema Successful")
		}
	}

	request.Log("Reusing existing MySQL connection")
	return
}

func (repository MysqlRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-ALL")
	response := RepositoryResponse{}
	session, isError, errorMessage := getMysqlConnection(request)
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		isError = false
		//Process A : Get Count of DB

		skip := "0"
		if request.Extras["skip"] != nil {
			skip = request.Extras["skip"].(string)
		}

		take := "100000"
		if request.Extras["take"] != nil {
			take = request.Extras["take"].(string)
		}
		var returnMap []map[string]interface{}

		rows, err := session.Query("SELECT * FROM " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + " limit " + take + " offset " + skip)

		if err != nil {
			//response.IsSuccess = false
			//response.GetErrorResponse("Error getting values for all objects in MySQL" + err.Error())
			response.IsSuccess = true
			response.Message = "No objects found in Couchbase"
			var emptyMap map[string]interface{}
			emptyMap = make(map[string]interface{})
			byte, _ := json.Marshal(emptyMap)
			response.GetResponseWithBody(byte)
		} else {
			response.IsSuccess = true
			response.Message = "Successfully retrieved values for all objects in MySQL"
			request.Log(response.Message)

			columns, _ := rows.Columns()
			count := len(columns)
			values := make([]interface{}, count)
			valuePtrs := make([]interface{}, count)

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
						//v = string(b)
						v = getCastableValue(string(b))
					} else {
						v = val
					}
					
					tempMap[col] = v
				}
				returnMap = append(returnMap, tempMap)
			}

			if len(returnMap) == 0 {
				response.IsSuccess = true
				response.Message = "No objects found in MySQL"
				var emptyMap map[string]interface{}
				emptyMap = make(map[string]interface{})
				byte, _ := json.Marshal(emptyMap)
				response.GetResponseWithBody(byte)
			}

			if request.Controls.SendMetaData == "false" {

				for index, arrVal := range returnMap {
					for key, _ := range arrVal {
						if key == "osHeaders" {
							delete(returnMap[index], key)
						}
					}
				}
			}

			byteValue, errMarshal := json.Marshal(returnMap)
			if errMarshal != nil {
				response.IsSuccess = false
				response.GetErrorResponse("Error getting values for all objects in MySQL" + err.Error())
			} else {
				response.IsSuccess = true
				response.GetResponseWithBody(byteValue)
				response.Message = "Successfully retrieved values for all objects in MySQL"
				request.Log(response.Message)
			}
		}

	}

	return response
}

func getCastableValue(value string) (returnVal interface{}){
	
	if value == "\u0001" || value == "\u0000"{
		if value == "\u0001"{
			returnVal = true
		}else if value == "\u0000"{
			returnVal = false
		}
	}else if iVal, err := strconv.Atoi(value); err == nil {
		returnVal = iVal
	}else{
		returnVal = value
	}

	return
}

func (repository MysqlRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Get Search not implemented in MySQL Db repository")
	return getDefaultNotImplemented()
}

func (repository MysqlRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-QUERY")
	response := RepositoryResponse{}

	queryType := request.Body.Query.Type
	switch queryType {
	case "Query":
		if request.Body.Query.Parameters != "*" {
			fieldsInByte := executeMySQLQuery(request)
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
		request.Log(queryType + " not implemented in MySQL Db repository")
		return getDefaultNotImplemented()

	}

	return response
}

func executeMySQLQuery(request *messaging.ObjectRequest) (returnByte []byte) {

	//if checkIfTenantIsAllowed(request.Body.Query.Parameters, request.Controls.Namespace) {

	//request.Log("This Tenent is ALLOWED to perform this Query!")
	session, isError, _ := getMysqlConnection(request)
	if isError == true {
		returnByte = nil
	} else {
		isError = false
		//Process A : Get Count of DB
		request.Log("User Input Query : " + request.Body.Query.Parameters)
		formattedQuery := queryparser.GetFormattedQuery(request.Body.Query.Parameters)
		request.Log("Formatted MySQL Query : " + formattedQuery)

		//var returnMap map[string]interface{}
		//returnMap = make(map[string]interface{})
		var returnMap []map[string]interface{}

		rows, err := session.Query(formattedQuery)

		if err != nil {
			request.Log("Error executing query in MySQL")
		} else {
			request.Log("Successfully executed query in MySQL")
			columns, _ := rows.Columns()
			count := len(columns)
			values := make([]interface{}, count)
			valuePtrs := make([]interface{}, count)

			//index := 0
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
				returnMap = append(returnMap, tempMap)
				//returnMap[strconv.Itoa(index)] = tempMap
				//index++
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
	//} else {
	//returnByte = ([]byte("This Tenent is NOT ALLOWED to perform submitted Query!"))
	//return ([]byte("This Tenent is NOT ALLOWED to perform submitted Query!"))
	//}

	return returnByte
}

func (repository MysqlRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-BY-KEY")
	response := RepositoryResponse{}
	session, isError, errorMessage := getMysqlConnection(request)
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		isError = false
		request.Log("Id key : " + request.Controls.Id)

		var myMap map[string]interface{}
		myMap = make(map[string]interface{})

		var keyMap map[string]interface{}
		keyMap = make(map[string]interface{})

		skip := "0"
		if request.Extras["skip"] != nil {
			skip = request.Extras["skip"].(string)
		}

		take := "100000"
		if request.Extras["take"] != nil {
			take = request.Extras["take"].(string)
		}

		fieldName := ""
		parameter := request.Controls.Id
		if request.Extras["fieldName"] != nil {
			fieldName = request.Extras["fieldName"].(string)
			parameter = request.Controls.Id
		} else {
			request.Log("Getting Primary Key")
			rows, err := session.Query("SELECT DISTINCT COLUMN_NAME FROM INFORMATION_SCHEMA.key_column_usage where TABLE_SCHEMA='" + getMySQLnamespace(request) + "' AND TABLE_NAME='" + strings.ToLower(request.Controls.Class) + "';")

			if err != nil {
				//response.IsSuccess = false
				//response.Message = "Error reading primary key"
				//request.Log("Error reading primary key")
				//return response
				fmt.Println(err.Error())
				response.IsSuccess = true
				response.Message = "No objects found in MySQL"
				var emptyMap map[string]interface{}
				emptyMap = make(map[string]interface{})
				byte, _ := json.Marshal(emptyMap)
				response.GetResponseWithBody(byte)

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
						//v = string(b)
						v = getCastableValue(string(b))
					} else {
						v = val
					}

						keyMap[col] = v

					}
				}
			}

			if len(keyMap) == 0 {
				response.IsSuccess = true
				response.Message = "No objects found in MySQL"
				var emptyMap map[string]interface{}
				emptyMap = make(map[string]interface{})
				byte, _ := json.Marshal(emptyMap)
				response.GetResponseWithBody(byte)
				return response
			}
			if keyMap["COLUMN_NAME"] == nil{
				response.IsSuccess = true
				response.Message = "No objects found in MySQL"
				var emptyMap map[string]interface{}
				emptyMap = make(map[string]interface{})
				byte, _ := json.Marshal(emptyMap)
				response.GetResponseWithBody(byte)
				return response
			}
			fieldName = keyMap["COLUMN_NAME"].(string)
		}

		request.Log("KeyProperty : " + fieldName)
		request.Log("KeyValue : " + request.Controls.Id)
		rows, err := session.Query("SELECT * FROM " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + " where " + fieldName + " = '" + parameter + "'" + " limit " + take + " offset " + skip)

		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting values for all objects in MySQL" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully retrieved values for all objects in MySQL"
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
					if index == "osHeaders" {
						delete(myMap, index)
					}
				}
			}

			byteValue, errMarshal := json.Marshal(myMap)
			if errMarshal != nil {
				response.IsSuccess = false
				response.GetErrorResponse("Error getting values for all objects in MySQL" + err.Error())
			} else {
				response.IsSuccess = true
				response.GetResponseWithBody(byteValue)
				response.Message = "Successfully retrieved values for all objects in MySQL"
				request.Log(response.Message)
			}
		}
	}

	return response
}

func (repository MysqlRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-MULTIPLE")
	response := RepositoryResponse{}
	session, isError, errorMessage := getMysqlConnection(request)

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
					tempMapObject["osHeaders"] = value
				} else {
					tempMapObject[key] = value
				}
			}

			DataObjects[i] = tempMapObject
		}
		//check for table in MsSql
		if createMySQLTable(request, session) {
			request.Log("Table Verified Successfully!")
		} else {
			response.IsSuccess = false
			return response
		}

		indexNames := getMySQLFieldOrder(request)
		fmt.Println("Index Names : ")
		fmt.Println(indexNames)
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
				keyValue := getMySqlRecordID(request, DataObjects[i])
				DataObjects[i][request.Body.Parameters.KeyProperty] = keyValue
				idData[strconv.Itoa(i)] = keyValue
				if keyValue == "" {
					response.IsSuccess = false
					response.Message = "Failed inserting multiple object in MYSQL"
					request.Log(response.Message)
					request.Log("Inavalid ID request")
					return response
				}

				var keyArray = make([]string, noOfElements)
				var valueArray = make([]string, noOfElements)

				for index := 0; index < len(indexNames); index++ {

					if indexNames[index] != "osHeaders" {

						if _, ok := DataObjects[i][indexNames[index]].(string); ok {
							keyArray[index] = indexNames[index]
							valueArray[index] = DataObjects[i][indexNames[index]].(string)
						} else {
							//	fmt.Println("Non string value detected, Will be strigified!")
							keyArray[index] = indexNames[index]
							valueArray[index] = getStringByObject(DataObjects[i][indexNames[index]])
						}
					} else {
						// __osHeaders Catched!
						keyArray[index] = "osHeaders"
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
			//request.Log("INSERT INTO " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + " (" + argKeyList + ") VALUES " + argValueList + ";")
			fmt.Println("INSERT INTO " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + " (" + argKeyList + ") VALUES " + argValueList + ";")
			
			_, err := session.Query("INSERT INTO " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + " (" + argKeyList + ") VALUES " + argValueList + ";")
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
				keyValue := getMySqlRecordID(request, DataObjects[i])
				DataObjects[i][request.Body.Parameters.KeyProperty] = keyValue
				idData[strconv.Itoa(i)] = keyValue
				if keyValue == "" {
					response.IsSuccess = false
					response.Message = "Failed inserting multiple object in MYSQL"
					request.Log(response.Message)
					request.Log("Inavalid ID request")
					return response
				}

				var keyArray = make([]string, noOfElements)
				var valueArray = make([]string, noOfElements)

				for index := 0; index < len(indexNames); index++ {
					if indexNames[index] != "osHeaders" {

						if _, ok := DataObjects[i][indexNames[index]].(string); ok {
							keyArray[index] = indexNames[index]
							valueArray[index] = DataObjects[i][indexNames[index]].(string)
						} else {
							//	fmt.Println("Non string value detected, Will be strigified!")
							keyArray[index] = indexNames[index]
							valueArray[index] = getStringByObject(DataObjects[i][indexNames[index]])
						}
					} else {
						// __osHeaders Catched!
						keyArray[index] = "osHeaders"
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
			//request.Log("INSERT INTO " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + " (" + argKeyList + ") VALUES " + argValueList + ";")
			_, err := session.Query("INSERT INTO " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + " (" + argKeyList + ") VALUES " + argValueList + ";")
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
			response.Message = "Successfully inserted many objects in to MYSQL"
			request.Log(response.Message)
		} else {
			response.IsSuccess = false
			response.Message = "Error inserting many objects in to MYSQL"
			request.Log(response.Message)
			response.GetErrorResponse("Error inserting many objects in to MYSQL")
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
	response.IsSuccess = true
	response.Message = "Successfully inserted Many object in MySQL"
	session.Close()
	return response
}

func (repository MysqlRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-SINGLE")
	response := RepositoryResponse{}
	keyValue := getMySqlRecordID(request, nil)

	session, isError, errorMessage := getMysqlConnection(request)
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else if keyValue != "" {

		//copy array
		var DataObject map[string]interface{}
		DataObject = make(map[string]interface{})

		for key, value := range request.Body.Object {
			if key == "__osHeaders" {
				DataObject["osHeaders"] = value
			} else {
				DataObject[key] = value
			}
		}

		noOfElements := len(DataObject)
		DataObject[request.Body.Parameters.KeyProperty] = keyValue

		if createMySQLTable(request, session) {
			request.Log("Table Verified Successfully!")
		} else {
			response.IsSuccess = false
			return response
		}

		indexNames := getMySQLFieldOrder(request)

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
		var valueArray = make([]interface{}, noOfElements)

		// Process A :start identifying individual data in array and convert to string
		for index := 0; index < len(indexNames); index++ {
			if indexNames[index] != "osHeaders" {

				if _, ok := DataObject[indexNames[index]].(string); ok {
					keyArray[index] = indexNames[index]
					valueArray[index] = DataObject[indexNames[index]].(string)
				} else {
					//fmt.Println("Non string value detected, Will be strigified!")
					keyArray[index] = indexNames[index]
					//valueArray[index] = getStringByObject(DataObject[indexNames[index]])
					valueArray[index] = DataObject[indexNames[index]]
				}
			} else {
				// __osHeaders Catched!
				keyArray[index] = "osHeaders"
				valueArray[index] = ConvertOsheaders(DataObject[indexNames[index]].(messaging.ControlHeaders))
			}

		}
		/*
					recordFieldList = make([]string, len(request.Body.Object))
					recordFieldType = make([]string, len(request.Body.Object))
					index := 0
					for key, value := range request.Body.Object {
						if key == "__osHeaders" {
							recordFieldList[index] = "osHeaders"
							recordFieldType[index] = "text"
						} else {
							recordFieldList[index] = key
							recordFieldType[index] = getMySqlDataType(value)
						}
						index++
					}
		*/
		fmt.Println("**********************************")
		//Build the query string
		for i := 0; i < noOfElements; i++ {
			dataType := getMySqlDataType(valueArray[i])
			fmt.Println(dataType)
			if i != noOfElements-1 {
				if dataType == "BIT" || dataType == "real"{
					argValueList = argValueList + getStringByObject(valueArray[i]) + ", "
				}else{
					argValueList = argValueList + "'" + getStringByObject(valueArray[i]) + "'" + ", "
				}
			} else {
				if dataType == "BIT" || dataType == "real"{
					argValueList = argValueList + getStringByObject(valueArray[i])
				}else{
					argValueList = argValueList + "'" + getStringByObject(valueArray[i]) + "'"
				}
			}
		}
		fmt.Println("**********************************")
		//DEBUG USE : Display Query information
		//fmt.Println("Table Name : " + request.Controls.Class)
		//fmt.Println("Key list : " + argKeyList)
		//fmt.Println("Value list : " + argValueList)
		//fmt.Println("INSERT INTO " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + " (" + argKeyList + ") VALUES (" + argValueList + ");")
		_, err := session.Query("INSERT INTO " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + " (" + argKeyList + ") VALUES (" + argValueList + ");")
		fmt.Println("INSERT INTO " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + " (" + argKeyList + ") VALUES (" + argValueList + ");")
		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error inserting one object in MySQL" + err.Error())
			response.Message = err.Error()
			fmt.Println(err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted one object in MySQL"
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
	response.IsSuccess = true
	response.Message = "Successfully inserted one object in MySQL"
	return response
}

func getMySQLFieldOrder(request *messaging.ObjectRequest) []string {
	var returnArray []string
	//read fields
	byteValue := executeMySqlGetFields(request)

	err := json.Unmarshal(byteValue, &returnArray)
	if err != nil {
		request.Log("Converstion of Json Failed!")
		returnArray = make([]string, 1)
		returnArray[0] = "nil"
		return returnArray
	}

	return returnArray
}

func createMySQLTable(request *messaging.ObjectRequest, session *sql.DB) (status bool) {
	status = false

	//get table list
	classBytes := executeMySqlGetClasses(request)
	var classList []string
	err := json.Unmarshal(classBytes, &classList)
	if err != nil {
		status = false
	} else {
		for _, className := range classList {
			if strings.ToLower(request.Controls.Class) == className {
				fmt.Println("Table Already Available")
				status = true
				//Get all fields
				classBytes := executeMySqlGetFields(request)
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
							recordFieldList[index] = "osHeaders"
							recordFieldType[index] = "text"
						} else {
							recordFieldList[index] = key
							recordFieldType[index] = getMySqlDataType(value)
						}
						index++
					}
				} else {
					recordFieldList = make([]string, len(request.Body.Object))
					recordFieldType = make([]string, len(request.Body.Object))
					index := 0
					for key, value := range request.Body.Object {
						if key == "__osHeaders" {
							recordFieldList[index] = "osHeaders"
							recordFieldType[index] = "text"
						} else {
							recordFieldList[index] = key
							recordFieldType[index] = getMySqlDataType(value)
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
					_, er := session.Query("ALTER TABLE " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + " ADD COLUMN " + newFields[key] + " " + newTypes[key] + ";")
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
					dataObject["osHeaders"] = value
				} else {
					dataObject[key] = value
				}
			}
		} else {
			for key, value := range request.Body.Objects[0] {
				if key == "__osHeaders" {
					dataObject["osHeaders"] = value
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
			dataTypeArray[startIndex] = getMySqlDataType(value)
			startIndex = startIndex + 1

		}

		//Create Table

		var argKeyList2 string

		for i := 0; i < noOfElements; i++ {
			if i != noOfElements-1 {
				if keyArray[i] == request.Body.Parameters.KeyProperty {
					argKeyList2 = argKeyList2 + keyArray[i] + " VARCHAR(255) PRIMARY KEY, "
				} else {
					argKeyList2 = argKeyList2 + keyArray[i] + " " + dataTypeArray[i] + ", "
				}

			} else {
				if keyArray[i] == request.Body.Parameters.KeyProperty {
					argKeyList2 = argKeyList2 + keyArray[i] + " VARCHAR(255) PRIMARY KEY"
				} else {
					argKeyList2 = argKeyList2 + keyArray[i] + " " + dataTypeArray[i]
				}

			}
		}

		request.Log("create table " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + "(" + argKeyList2 + ");")
		
		_, er := session.Query("create table " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + "(" + argKeyList2 + ");")
		if er != nil {
			status = false
			request.Log("Table Creation Failed : " + er.Error())
			return
		}

		status = true

	}

	return
}

func (repository MysqlRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-MULTIPLE")
	response := RepositoryResponse{}
	session, isError, errorMessage := getMysqlConnection(request)
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		if createMySQLTable(request, session) {
			request.Log("Table Verified Successfully!")
		} else {
			response.IsSuccess = false
			return response
		}

		for i := 0; i < len(request.Body.Objects); i++ {
			noOfElements := len(request.Body.Objects[i]) - 1
			var keyUpdate = make([]string, noOfElements)
			var valueUpdate = make([]string, noOfElements)

			var startIndex = 0
			for key, value := range request.Body.Objects[i] {
				if key != request.Body.Parameters.KeyProperty {
					if key != "__osHeaders" {
						//if str, ok := value.(string); ok {
						if _, ok := value.(string); ok {

							//Implement all MAP related logic here. All correct data are being caught in here
							keyUpdate[startIndex] = key
							valueUpdate[startIndex] = value.(string)
							startIndex = startIndex + 1

						} else {
							//request.Log("Non string value detected, Will be strigified!")
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
			//	fmt.Println("Table Name : " + request.Controls.Class)
			//	fmt.Println("Value list : " + argValueList)
			//request.Log("UPDATE " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + " SET " + argValueList + " WHERE " + request.Body.Parameters.KeyProperty + " =" + "'" + request.Body.Objects[i][request.Body.Parameters.KeyProperty].(string) + "'")
			_, err := session.Query("UPDATE " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + " SET " + argValueList + " WHERE " + request.Body.Parameters.KeyProperty + " =" + "'" + request.Body.Objects[i][request.Body.Parameters.KeyProperty].(string) + "'")

			if err != nil {
				response.IsSuccess = false
				request.Log("Error updating object in MySQL  : " + getNoSqlKey(request) + ", " + err.Error())
				response.GetErrorResponse("Error updating one object in MySQL because no match was found!" + err.Error())
			} else {
				response.IsSuccess = true
				response.Message = "Successfully updating one object in MySQL "
				request.Log(response.Message)
			}
		}

	}
	return response
}

func (repository MysqlRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-SINGLE")
	response := RepositoryResponse{}
	session, isError, errorMessage := getMysqlConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		if createMySQLTable(request, session) {
			request.Log("Table Verified Successfully!")
		} else {
			response.IsSuccess = false
			return response
		}

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
						//request.Log("Non string value detected, Will be strigified!")
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

		_, err := session.Query("UPDATE " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + " SET " + argValueList + " WHERE " + request.Body.Parameters.KeyProperty + " =" + "'" + request.Controls.Id + "'")

		if err != nil {
			response.IsSuccess = false
			request.Log("Error updating object in MySQL  : " + getNoSqlKey(request) + ", " + err.Error())
			response.GetErrorResponse("Error updating one object in MySQL because no match was found!" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully updating one object in MySQL "
			request.Log(response.Message)
		}

	}
	return response
}

func (repository MysqlRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-MULTIPLE")
	response := RepositoryResponse{}
	session, isError, errorMessage := getMysqlConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		for _, obj := range request.Body.Objects {
			_, err := session.Query("DELETE FROM " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class) + " WHERE " + request.Body.Parameters.KeyProperty + " = '" + obj[request.Body.Parameters.KeyProperty].(string) + "'")
			if err != nil {
				response.IsSuccess = false
				request.Log("Error deleting object in MySQL : " + err.Error())
				response.GetErrorResponse("Error deleting one object in MySQL because no match was found!" + err.Error())
			} else {
				response.IsSuccess = true
				response.Message = "Successfully deleted one object in MySQL"
				request.Log(response.Message)
			}
		}
	}

	return response
}

func (repository MysqlRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-SINGLE")
	response := RepositoryResponse{}
	session, isError, errorMessage := getMysqlConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		_, err := session.Query("DELETE FROM " + getMySQLnamespace(request) + "." + request.Controls.Class + " WHERE " + request.Body.Parameters.KeyProperty + " = '" + request.Controls.Id + "'")
		if err != nil {
			response.IsSuccess = false
			request.Log("Error deleting object in MySQL : " + err.Error())
			response.GetErrorResponse("Error deleting one object in MySQL because no match was found!" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully deleted one object in MySQL"
			request.Log(response.Message)
		}
	}

	return response
}

func (repository MysqlRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	request.Log("Starting SPECIAL!")
	queryType := request.Body.Special.Type

	switch queryType {
	case "getFields":
		request.Log("Starting GET-FIELDS sub routine!")
		fieldsInByte := executeMySqlGetFields(request)
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
		fieldsInByte := executeMySqlGetClasses(request)
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
		fieldsInByte := executeMySqlGetNamespaces(request)
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
		request.Log("Starting GET-SELECED sub routine")
		fieldsInByte := executeMySqlGetSelected(request)
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

func (repository MysqlRepository) Test(request *messaging.ObjectRequest) {

}

//Sub Routines

func executeMySqlGetFields(request *messaging.ObjectRequest) (returnByte []byte) {

	namespace := getMySQLnamespace(request)
	class := strings.ToLower(request.Controls.Class)

	session, isError, _ := getMysqlConnection(request)
	if isError == true {
		returnByte = nil
	} else {
		isError = false

		var returnMap map[string]interface{}
		returnMap = make(map[string]interface{})

		rows, err := session.Query("describe " + namespace + "." + class)

		if err != nil {
			request.Log("Error executing query in MySQL")
		} else {
			request.Log("Successfully executed query in MySQL")
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

				returnMap[strconv.Itoa(index)] = tempMap["Field"]
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

	return returnByte
}

func executeMySqlGetClasses(request *messaging.ObjectRequest) (returnByte []byte) {

	namespace := getMySQLnamespace(request)

	session, isError, _ := getMysqlConnection(request)
	if isError == true {
		returnByte = nil
	} else {
		isError = false

		var returnMap map[string]interface{}
		returnMap = make(map[string]interface{})

		rows, err := session.Query("SELECT DISTINCT TABLE_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='" + namespace + "';")

		if err != nil {
			request.Log("Error executing query in MySQL")
		} else {
			request.Log("Successfully executed query in MySQL")
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

			var classArray []string
			classArray = make([]string, len(returnMap))

			for key, value := range returnMap {
				index, _ := strconv.Atoi(key)
				classArray[index] = value.(string)
			}

			byteValue, errMarshal := json.Marshal(classArray)
			if errMarshal != nil {
				request.Log("Error converting to byte array")
				byteValue = nil
			} else {
				request.Log("Successfully converted result to byte array")
			}

			returnByte = byteValue
		}

	}

	return
}

func executeMySqlGetNamespaces(request *messaging.ObjectRequest) (returnByte []byte) {
	session, isError, _ := getMysqlConnection(request)
	if isError == true {
		returnByte = nil
	} else {
		isError = false

		var returnMap map[string]interface{}
		returnMap = make(map[string]interface{})

		rows, err := session.Query("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME != 'information_schema' AND SCHEMA_NAME !='mysql' AND SCHEMA_NAME !='performance_schema';")

		if err != nil {
			request.Log("Error executing query in MySQL")
		} else {
			request.Log("Successfully executed query in MySQL")
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

				returnMap[strconv.Itoa(index)] = tempMap["SCHEMA_NAME"]
				index++
			}

			var schemaArray []string
			schemaArray = make([]string, len(returnMap))

			for key, value := range returnMap {
				index, _ := strconv.Atoi(key)
				schemaArray[index] = value.(string)
			}

			byteValue, errMarshal := json.Marshal(schemaArray)
			if errMarshal != nil {
				request.Log("Error converting to byte array")
				byteValue = nil
			} else {
				request.Log("Successfully converted result to byte array")
			}

			returnByte = byteValue
		}

	}

	return
}

func executeMySqlGetSelected(request *messaging.ObjectRequest) (returnByte []byte) {

	session, isError, _ := getMysqlConnection(request)
	if isError == true {
		request.Log("Error Connecting to MySQL")
	} else {
		var returnMap []map[string]interface{}
		//returnMap = make(map[string]interface{})

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

		rows, err := session.Query("SELECT " + selectedItemsQuery + " FROM " + getMySQLnamespace(request) + "." + request.Controls.Class)
		request.Log("SELECT " + selectedItemsQuery + " FROM " + getMySQLnamespace(request) + "." + strings.ToLower(request.Controls.Class))
		if err != nil {
			request.Log("Error Fetching data from MySQL")
		} else {
			request.Log("Successfully fetched data from MySQL")
			columns, _ := rows.Columns()
			count := len(columns)
			values := make([]interface{}, count)
			valuePtrs := make([]interface{}, count)

			//index := 0
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

				returnMap = append(returnMap, tempMap)
				//returnMap[strconv.Itoa(index)] = tempMap
				//index++
			}

			byteValue, _ := json.Marshal(returnMap)
			returnByte = byteValue
		}

	}

	return returnByte
}

func getMySqlRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
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
		session, isError, _ := getMysqlConnection(request)
		if isError {
			returnID = ""
			request.Log("Connecting to MySQL Failed!")
		} else {
			//Read Table domainClassAttributes
			request.Log("Reading maxCount from DB")
			rows, err := session.Query("SELECT maxCount FROM " + getMySQLnamespace(request) + ".domainClassAttributes where class = '" + strings.ToLower(request.Controls.Class) + "';")

			if err != nil {
				//If err create new domainClassAttributes  table
				request.Log("No Class found.. Must be a new namespace")
				_, err = session.Query("create table " + getMySQLnamespace(request) + ".domainClassAttributes ( class text primary key, maxCount text, version text);")
				if err != nil {
					returnID = ""
					return
				} else {
					//insert record with count 1 and return
					_, err := session.Query("INSERT INTO " + getMySQLnamespace(request) + ".domainClassAttributes (class, maxCount,version) VALUES ('" + strings.ToLower(request.Controls.Class) + "','1','" + uuid.NewV1().String() + "')")
					if err != nil {
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
					_, err = session.Query("INSERT INTO " + getMySQLnamespace(request) + ".domainClassAttributes (class,maxCount,version) values ('" + strings.ToLower(request.Controls.Class) + "', '1', '" + uuid.NewV1().String() + "');")
					if err != nil {
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
					_, err = session.Query("UPDATE " + getMySQLnamespace(request) + ".domainClassAttributes SET maxCount='" + returnID + "' WHERE class = '" + strings.ToLower(request.Controls.Class) + "' ;")
					if err != nil {
						request.Log("Error Updating index table : " + err.Error())
						returnID = ""
						return
					}
				}
			}
		}
	} else {
		//request.Log("Manual Key requested!")
		if obj == nil {
			returnID = request.Controls.Id
		} else {
			returnID = obj[request.Body.Parameters.KeyProperty].(string)
		}
	}

	return
}

func getMySqlDataType(item interface{}) (datatype string) {
	fmt.Println(item)
	datatype = reflect.TypeOf(item).Name()
	if datatype == "bool" {
		datatype = "BIT"
	} else if datatype == "float64" {
		datatype = "real"
	} else if datatype == "" || datatype == "string" || datatype == "ControlHeaders" {
		datatype = "text"
	}
	return datatype
}
