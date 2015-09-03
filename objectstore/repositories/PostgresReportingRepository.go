package repositories

import (
	"database/sql"
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/twinj/uuid"
	"strconv"
	"strings"
)

type PostgresReportingRepository struct {
}

func (repository PostgresReportingRepository) GetRepositoryName() string {
	return "Postgres DB"
}

func getPostgresReportingConnection(request *messaging.ObjectRequest) (session *sql.DB, isError bool, errorMessage string) {
	isError = false
	username := request.Configuration.ServerConfiguration["POSTGRESREPORTING"]["Username"]
	password := request.Configuration.ServerConfiguration["POSTGRESREPORTING"]["Password"]
	dbUrl := request.Configuration.ServerConfiguration["POSTGRESREPORTING"]["Url"]
	dbPort := request.Configuration.ServerConfiguration["POSTGRESREPORTING"]["Port"]

	//session, err := sql.Open("postgres", "host="+dbUrl+" port="+dbPort+" user="+username+" password="+password+" dbname="+(getPostgresNamespace(request))+" sslmode=disable")
	session, err := sql.Open("postgres", "host="+dbUrl+" port="+dbPort+" user="+username+" password="+password+" dbname="+"test"+" sslmode=disable")

	if err != nil {
		isError = true
		request.Log("There is an error")
		errorMessage = err.Error()
		request.Log("Postgres connection initilizing failed!")
	}

	//Create schema if not available.
	request.Log("Checking if Database " + getPostgresReportingSQLnamespace(request) + " is available.")

	isDatabaseAvailbale := false

	rows, err := session.Query("SELECT datname FROM pg_database WHERE datistemplate = false;")

	if err != nil {
		request.Log("Error contacting PostGres Server")
	} else {
		request.Log("Successfully retrieved values for all objects in PostGres")

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
				request.Log("Check domain : " + getPostgresReportingSQLnamespace(request) + " : available schema : " + v.(string))
				if v.(string) == getPostgresReportingSQLnamespace(request) {
					//Database available
					isDatabaseAvailbale = true
					break
				}
			}
		}
	}

	if isDatabaseAvailbale {
		request.Log("Database already available. Nothing to do. Proceed!")
		session.Close()
		session, err = sql.Open("postgres", "host="+dbUrl+" port="+dbPort+" user="+username+" password="+password+" dbname="+(getPostgresReportingSQLnamespace(request))+" sslmode=disable")

	} else {
		_, err = session.Query("CREATE DATABASE " + getPostgresReportingSQLnamespace(request) + ";")
		if err != nil {
			request.Log("Creation of domain matched Schema failed")
		} else {
			request.Log("Creation of domain matched Schema Successful")
			session.Close()
			session, err = sql.Open("postgres", "host="+dbUrl+" port="+dbPort+" user="+username+" password="+password+" dbname="+(getPostgresReportingSQLnamespace(request))+" sslmode=disable")
			if err != nil {
				request.Log("Relogin Failed!")
			} else {
				request.Log("Relogin successful!")
			}

		}
	}
	request.Log("Reusing existing Postgres connection")
	return
}

func (repository PostgresReportingRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-ALL")
	response := RepositoryResponse{}
	session, isError, errorMessage := getPostgresReportingConnection(request)
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		isError = false
		skip := "0"
		if request.Extras["skip"] != nil {
			skip = request.Extras["skip"].(string)
		}

		take := "100000"
		if request.Extras["take"] != nil {
			take = request.Extras["take"].(string)
		}

		var returnMap []map[string]interface{}

		rows, err := session.Query("SELECT * FROM " + request.Controls.Class + " limit " + take + " offset " + skip)

		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting values for all objects in Postgres" + err.Error())
			response.Message = "Table Not Found in Database : " + getPostgresReportingSQLnamespace(request)
		} else {
			response.IsSuccess = true
			response.Message = "Successfully retrieved values for all objects in Postgres"
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

			byteValue, errMarshal := json.Marshal(returnMap)
			if errMarshal != nil {
				response.IsSuccess = false
				response.GetErrorResponse("Error getting values for all objects in Postgres" + err.Error())
			} else {
				response.IsSuccess = true
				response.GetResponseWithBody(byteValue)
				response.Message = "Successfully retrieved values for all objects in mongo"
				request.Log(response.Message)
			}
		}
	}
	session.Close()
	return response
}

func (repository PostgresReportingRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Get Search not implemented in Postgres Db repository")
	return getDefaultNotImplemented()
}

func (repository PostgresReportingRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-QUERY")
	response := RepositoryResponse{}

	queryType := request.Body.Query.Type
	switch queryType {
	case "Query":
		if request.Body.Query.Parameters != "*" {
			fieldsInByte := executePostgresReportingQuery(request)
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
		request.Log(queryType + " not implemented in Postgres_SQL Db repository")
		return getDefaultNotImplemented()

	}

	return response
}

func executePostgresReportingQuery(request *messaging.ObjectRequest) (returnByte []byte) {
	session, isError, _ := getPostgresReportingConnection(request)
	if isError == true {
		returnByte = nil
	} else {
		isError = false
		//Process A : Get Count of DB
		var returnMap map[string]interface{}
		returnMap = make(map[string]interface{})

		rows, err := session.Query(request.Body.Query.Parameters)

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

func (repository PostgresReportingRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {

	request.Log("Starting GET-BY-KEY")
	response := RepositoryResponse{}
	session, isError, errorMessage := getPostgresReportingConnection(request)
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
			rows, err := session.Query("SELECT c.column_name FROM information_schema.table_constraints tc JOIN information_schema.constraint_column_usage AS ccu USING (constraint_schema, constraint_name)JOIN information_schema.columns AS c ON c.table_schema = tc.constraint_schema AND tc.table_name = c.table_name AND ccu.column_name = c.column_name where constraint_type = 'PRIMARY KEY' and tc.table_name = '" + request.Controls.Class + "';")

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

			fmt.Println("&&&&&&&&&&&&&&&&&&&&&&&&&&")
			fmt.Println(keyMap)

			fieldName = keyMap["column_name"].(string)
		}

		request.Log("KeyProperty : " + fieldName)
		request.Log("KeyValue : " + request.Controls.Id)
		rows, err := session.Query("SELECT * FROM " + request.Controls.Class + " where " + fieldName + " = '" + parameter + "';")

		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting values for all objects in PostgresSQL" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully retrieved values for all objects in PostgresSQL"
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

			byteValue, errMarshal := json.Marshal(myMap)
			if errMarshal != nil {
				response.IsSuccess = false
				response.GetErrorResponse("Error getting values for all objects in PostgresSQL" + err.Error())
			} else {
				response.IsSuccess = true
				response.GetResponseWithBody(byteValue)
				response.Message = "Successfully retrieved values for all objects in PostgresSQL"
				request.Log(response.Message)
			}
		}
	}
	session.Close()
	return response
}

func createPostgresReportingTable(request *messaging.ObjectRequest, session *sql.DB) (status bool) {
	status = false

	//get table list
	classBytes := executePostgresReportingGetClasses(request)
	var classList []string
	err := json.Unmarshal(classBytes, &classList)
	if err != nil {
		status = false
	} else {
		for _, className := range classList {
			if request.Controls.Class == className {
				fmt.Println("Table Already Available")
				status = true
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
					dataObject[strings.ToLower(key)] = value
				}
			}
		} else {
			for key, value := range request.Body.Objects[0] {
				if key == "__osHeaders" {
					dataObject["osheaders"] = value
				} else {
					dataObject[strings.ToLower(key)] = value
				}
			}
		}
		//read fields
		noOfElements := len(dataObject)
		var keyArray = make([]string, noOfElements)

		var startIndex int = 0

		for key, _ := range dataObject {
			keyArray[startIndex] = key
			startIndex = startIndex + 1

		}

		//Create Table

		var argKeyList2 string

		for i := 0; i < noOfElements; i++ {
			if i != noOfElements-1 {
				if keyArray[i] == strings.ToLower(request.Body.Parameters.KeyProperty) {
					argKeyList2 = argKeyList2 + keyArray[i] + " text PRIMARY KEY, "
				} else {
					argKeyList2 = argKeyList2 + keyArray[i] + " text, "
				}

			} else {
				if keyArray[i] == strings.ToLower(request.Body.Parameters.KeyProperty) {
					argKeyList2 = argKeyList2 + keyArray[i] + " text PRIMARY KEY"
				} else {
					argKeyList2 = argKeyList2 + keyArray[i] + " text"
				}

			}
		}

		request.Log("create table " + request.Controls.Class + "(" + argKeyList2 + ");")

		_, er := session.Query("create table " + request.Controls.Class + "(" + argKeyList2 + ");")
		if er != nil {
			status = false
			request.Log("Table Creation Failed : " + er.Error())
			return
		}

		status = true

	}

	return
}

func getPostgresReportingFieldOrder(request *messaging.ObjectRequest) []string {
	var returnArray []string
	//read fields
	byteValue := executePostgresReportingGetFields(request)

	err := json.Unmarshal(byteValue, &returnArray)
	if err != nil {
		request.Log("Converstion of Json Failed!")
		returnArray = make([]string, 1)
		returnArray[0] = "nil"
		return returnArray
	}

	return returnArray
}

func (repository PostgresReportingRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-MULTIPLE")
	response := RepositoryResponse{}
	session, isError, errorMessage := getPostgresReportingConnection(request)

	var idData map[string]interface{}
	idData = make(map[string]interface{})

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		//update table with Namespace field

		for i := 0; i < len(request.Body.Objects); i++ {
			request.Body.Objects[i]["Namespace"] = request.Controls.Namespace
		}

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
		//check for table in postgres

		if createPostgresReportingTable(request, session) {
			request.Log("Table Verified Successfully!")
		} else {
			response.IsSuccess = false
			return response
		}

		indexNames := getPostgresReportingFieldOrder(request)

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

		for i := 0; i < len(DataObjects); i++ {
			noOfElements := len(DataObjects[i])
			keyValue := getPostgresReportingSqlRecordID(request, DataObjects[i])
			DataObjects[i][strings.ToLower(request.Body.Parameters.KeyProperty)] = keyValue
			idData[strconv.Itoa(i)] = keyValue
			if keyValue == "" {
				response.IsSuccess = false
				response.Message = "Failed inserting multiple object in Cassandra"
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
			if i != len(DataObjects)-1 {
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
		_, err := session.Query("INSERT INTO " + request.Controls.Class + " (" + argKeyList + ") VALUES " + argValueList + ";")
		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error inserting one object in Postgres" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted one object in Postgres"
			request.Log(response.Message)
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

func (repository PostgresReportingRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-SINGLE")
	response := RepositoryResponse{}

	keyValue := getPostgresReportingSqlRecordID(request, nil)
	session, isError, errorMessage := getPostgresReportingConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else if keyValue != "" {
		request.Body.Object["Namespace"] = request.Controls.Namespace

		//change field names to Lower Case
		var DataObject map[string]interface{}
		DataObject = make(map[string]interface{})

		for key, value := range request.Body.Object {
			if key == "__osHeaders" {
				DataObject["osheaders"] = value
			} else {
				DataObject[strings.ToLower(key)] = value
			}
		}

		noOfElements := len(DataObject)
		DataObject[strings.ToLower(request.Body.Parameters.KeyProperty)] = keyValue

		if createPostgresReportingTable(request, session) {
			request.Log("Table Verified Successfully!")
		} else {
			response.IsSuccess = false
			return response
		}

		indexNames := getPostgresReportingFieldOrder(request)

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
		_, err := session.Query("INSERT INTO " + request.Controls.Class + " (" + argKeyList + ") VALUES (" + argValueList + ")")
		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error inserting one object in Postgres" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted one object in Postgres"
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

func (repository PostgresReportingRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-MULTIPLE")
	response := RepositoryResponse{}
	session, isError, errorMessage := getPostgresReportingConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		for i := 0; i < len(request.Body.Objects); i++ {
			noOfElements := len(request.Body.Objects[i]) /*2*/
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

			_, err := session.Query("UPDATE " + request.Controls.Class + " SET " + argValueList + " WHERE " + strings.ToLower(request.Body.Parameters.KeyProperty) + " =" + "'" + request.Body.Objects[i][strings.ToLower(request.Body.Parameters.KeyProperty)].(string) + "'")
			//err := collection.Update(bson.M{key: value}, bson.M{"$set": request.Body.Object})
			if err != nil {
				response.IsSuccess = false
				request.Log("Error updating object in Postgres  : " + getNoSqlKey(request) + ", " + err.Error())
				response.GetErrorResponse("Error updating one object in Postgres because no match was found!" + err.Error())
			} else {
				response.IsSuccess = true
				response.Message = "Successfully updating one object in Postgres "
				request.Log(response.Message)
			}
		}

	}

	session.Close()
	return response
}

func (repository PostgresReportingRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-SINGLE")
	response := RepositoryResponse{}
	session, isError, errorMessage := getPostgresReportingConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		noOfElements := len(request.Body.Object) /*2*/
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

		_, err := session.Query("UPDATE " + request.Controls.Class + " SET " + argValueList + " WHERE " + strings.ToLower(request.Body.Parameters.KeyProperty) + " =" + "'" + request.Controls.Id + "'")
		//err := collection.Update(bson.M{key: value}, bson.M{"$set": request.Body.Object})
		if err != nil {
			response.IsSuccess = false
			request.Log("Error updating object in Postgres  : " + getNoSqlKey(request) + ", " + err.Error())
			response.GetErrorResponse("Error updating one object in Postgres because no match was found!" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully updating one object in Postgres "
			request.Log(response.Message)
		}

	}

	session.Close()
	return response
}

func (repository PostgresReportingRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-MULTIPLE")
	response := RepositoryResponse{}
	session, isError, errorMessage := getPostgresReportingConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		for _, obj := range request.Body.Objects {
			_, err := session.Query("DELETE FROM " + request.Controls.Class + " WHERE " + strings.ToLower(request.Body.Parameters.KeyProperty) + " = '" + obj[strings.ToLower(request.Body.Parameters.KeyProperty)].(string) + "'")
			if err != nil {
				response.IsSuccess = false
				request.Log("Error deleting object in Postgres : " + err.Error())
				response.GetErrorResponse("Error deleting one object in Postgres because no match was found!" + err.Error())
			} else {
				response.IsSuccess = true
				response.Message = "Successfully deleted one object in Postgres"
				request.Log(response.Message)
			}
		}
	}

	session.Close()
	return response
}

func (repository PostgresReportingRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-SINGLE")
	response := RepositoryResponse{}
	session, isError, errorMessage := getPostgresReportingConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		_, err := session.Query("DELETE FROM " + request.Controls.Class + " WHERE " + strings.ToLower(request.Body.Parameters.KeyProperty) + " = '" + request.Controls.Id + "'")
		if err != nil {
			response.IsSuccess = false
			request.Log("Error deleting object in Postgres : " + err.Error())
			response.GetErrorResponse("Error deleting one object in Postgres because no match was found!" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully deleted one object in Postgres"
			request.Log(response.Message)
		}
	}

	session.Close()
	return response
}

func (repository PostgresReportingRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	request.Log("Starting SPECIAL!")
	queryType := request.Body.Special.Type

	switch queryType {
	case "getFields":
		request.Log("Starting GET-FIELDS sub routine!")
		fieldsInByte := executePostgresReportingGetFields(request)
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
		fieldsInByte := executePostgresReportingGetClasses(request)
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
		fieldsInByte := executePostgresReportingGetNamespaces(request)
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
		fieldsInByte := executePostgresReportingGetSelected(request)
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

func (repository PostgresReportingRepository) Test(request *messaging.ObjectRequest) {

}

//Sub Routines

func executePostgresReportingGetFields(request *messaging.ObjectRequest) (returnByte []byte) {

	class := request.Controls.Class
	session, isError, _ := getPostgresReportingConnection(request)
	if isError == true {
		returnByte = nil
	} else {
		isError = false

		var returnMap map[string]interface{}
		returnMap = make(map[string]interface{})

		rows, err := session.Query("select column_name from information_schema.columns where table_name='" + class + "';")

		if err != nil {
			request.Log("Error executing query in Postgres SQL")
		} else {
			request.Log("Successfully executed query in Postgres SQL")
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

				returnMap[strconv.Itoa(index)] = tempMap["column_name"]
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

func executePostgresReportingGetClasses(request *messaging.ObjectRequest) (returnByte []byte) {
	session, isError, _ := getPostgresReportingConnection(request)
	if isError == true {
		returnByte = nil
	} else {
		isError = false

		var returnMap map[string]interface{}
		returnMap = make(map[string]interface{})

		rows, err := session.Query(" SELECT table_name FROM information_schema.tables WHERE table_schema='public';")

		if err != nil {
			request.Log("Error executing query in Postgres SQL")
		} else {
			request.Log("Successfully executed query in Postgres SQL")
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

				returnMap[strconv.Itoa(index)] = tempMap["table_name"]
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

func executePostgresReportingGetNamespaces(request *messaging.ObjectRequest) (returnByte []byte) {
	session, isError, _ := getPostgresReportingConnection(request)
	if isError == true {
		returnByte = nil
	} else {
		isError = false

		var returnMap map[string]interface{}
		returnMap = make(map[string]interface{})

		rows, err := session.Query("SELECT datname FROM pg_database WHERE datistemplate = false;")

		if err != nil {
			request.Log("Error executing query in Postgres SQL")
		} else {
			request.Log("Successfully executed query in Postgres SQL")
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

				returnMap[strconv.Itoa(index)] = tempMap["datname"]
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

func executePostgresReportingGetSelected(request *messaging.ObjectRequest) (returnByte []byte) {

	session, isError, _ := getPostgresReportingConnection(request)
	if isError == true {
		request.Log("Error Connecting to Postgres")
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
		rows, err := session.Query("SELECT " + selectedItemsQuery + " FROM " + request.Controls.Class)

		if err != nil {
			request.Log("Error Fetching data from Postgres")
		} else {
			request.Log("Successfully fetched data from Postgres")
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

func getPostgresReportingSqlRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
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
		if (obj[strings.ToLower(request.Body.Parameters.KeyProperty)].(string) == "-999") || (request.Body.Parameters.AutoIncrement == true) {
			isAutoIncrementId = true
		}

		if (obj[strings.ToLower(request.Body.Parameters.KeyProperty)].(string) == "-888") || (request.Body.Parameters.GUIDKey == true) {
			isGUIDKey = true
		}

	}

	if isGUIDKey {
		//request.Log("GUID Key generation requested!")
		returnID = uuid.NewV1().String()
	} else if isAutoIncrementId {
		request.Log("Automatic Increment Key generation requested!")
		session, isError, _ := getPostgresReportingConnection(request)
		if isError {
			returnID = ""
			request.Log("Connecting to MySQL Failed!")
		} else {
			//Read Table domainClassAttributes
			request.Log("Reading maxCount from DB")
			rows, err := session.Query("SELECT maxCount FROM domainClassAttributes where class = '" + request.Controls.Class + "';")

			if err != nil {
				//If err create new domainClassAttributes  table
				request.Log("No Class found.. Must be a new namespace")
				_, err = session.Query("create table domainClassAttributes ( class text primary key, maxcount text, version text);")
				if err != nil {
					returnID = ""
					return
				} else {
					//insert record with count 1 and return
					_, err := session.Query("INSERT INTO domainClassAttributes (class, maxcount,version) VALUES ('" + request.Controls.Class + "','1','" + uuid.NewV1().String() + "')")
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
					_, err = session.Query("INSERT INTO domainClassAttributes (class,maxcount,version) values ('" + request.Controls.Class + "', '1', '" + uuid.NewV1().String() + "');")
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
					maxCount, err = strconv.Atoi(myMap["maxcount"].(string))
					maxCount++
					returnID = strconv.Itoa(maxCount)
					_, err = session.Query("UPDATE domainClassAttributes SET maxcount='" + returnID + "' WHERE class = '" + request.Controls.Class + "' ;")
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
		request.Log("Manual Key requested!")
		if obj == nil {
			returnID = request.Controls.Id
		} else {
			returnID = obj[strings.ToLower(request.Body.Parameters.KeyProperty)].(string)
		}
	}

	return
}

func getPostgresReportingSQLnamespace(request *messaging.ObjectRequest) string {
	return strings.ToLower(request.Configuration.ServerConfiguration["POSTGRESREPORTING"]["DatabaseName"])
}
