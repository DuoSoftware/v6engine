package repositories

import (
	"database/sql"
	"duov6.com/DuoEtlService/fileprocessor"
	"duov6.com/DuoEtlService/logger"
	"duov6.com/DuoEtlService/messaging"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/twinj/uuid"
	"strconv"
	"strings"
)

type Postgres5Repository struct {
}

func (repo Postgres5Repository) GetETLName() string {
	return "POSTGRESv5"
}

func (repo Postgres5Repository) ExecuteETLService(request *messaging.ETLRequest) messaging.ETLResponse {
	var response messaging.ETLResponse

	classPaths := fileprocessor.GetClassPaths(request.Configuration.DataPath)
	fmt.Println(classPaths)
	for _, classpath := range classPaths {
		metaData := strings.Split(classpath, "/")
		class := metaData[(len(metaData) - 1)]
		fmt.Println("Class : " + class)
		namespace := metaData[(len(metaData) - 2)]
		fmt.Println("Namespace : " + namespace)

		//Get Connection
		request.Controls.Namespace = namespace
		request.Controls.Class = class
		session, isError, errMsg := getPostgresReportingConnection(request)
		if isError {
			response.IsSuccess = false
			response.Message = errMsg
			return response
		}

		//Execute Posts
		objectArrayPost := fileprocessor.Process(classpath, "POST")

		for _, object := range objectArrayPost {
			//fmt.Println("Object Array Posts : ")
			//fmt.Println(objectArrayPost)
			var Request *messaging.ETLRequest
			Request = &messaging.ETLRequest{}
			Request.Configuration = request.Configuration
			Request.Controls.Namespace = namespace
			Request.Controls.Class = class
			Request.Body = object
			if object.Object != nil {
				Request.Controls.Id = object.Object[Request.Body.Parameters.KeyProperty].(string)
				//	fmt.Println(Request.Controls.Id)
			}
			fillControlHeaders(Request)
			logger.Log("\n")
			logger.Log("\n")
			insertToPostgres(Request, session)
			logger.Log("\n")
			logger.Log("\n")
		}

		//Execute UPDATES
		objectArrayPut := fileprocessor.Process(classpath, "PUT")

		for _, object := range objectArrayPut {
			//	fmt.Println("Object Array PUTS : ")
			//fmt.Println(objectArrayPut)
			var Request *messaging.ETLRequest
			Request = &messaging.ETLRequest{}
			Request.Configuration = request.Configuration
			Request.Controls.Namespace = namespace
			Request.Controls.Class = class
			Request.Body = object
			if object.Object != nil {
				Request.Controls.Id = object.Object[Request.Body.Parameters.KeyProperty].(string)
				//	fmt.Println(Request.Controls.Id)
			}
			fillControlHeaders(Request)
			logger.Log("\n")
			logger.Log("\n")
			updateOnPostgres(Request, session)
			logger.Log("\n")
			logger.Log("\n")
		}

		//Execute DELETES
		objectArrayDelete := fileprocessor.Process(classpath, "DELETE")

		for _, object := range objectArrayDelete {
			//fmt.Println("Object Array DELETES : ")
			//fmt.Println(objectArrayDelete)
			var Request *messaging.ETLRequest
			Request = &messaging.ETLRequest{}
			Request.Configuration = request.Configuration
			Request.Controls.Namespace = namespace
			Request.Controls.Class = class
			Request.Body = object
			if object.Object != nil {
				Request.Controls.Id = object.Object[Request.Body.Parameters.KeyProperty].(string)
				//	fmt.Println(Request.Controls.Id)
			}
			fillControlHeaders(Request)
			logger.Log("\n")
			logger.Log("\n")
			deleteFromPostgres(Request, session)
			logger.Log("\n")
			logger.Log("\n")
		}

		session.Close()

	}

	for _, classpath := range classPaths {
		fileprocessor.ClearCompletedFiles(classpath)
	}
	//fmt.Println(getPostgresReportingSQLnamespace(request))
	response.IsSuccess = true
	response.Message = "Postgres Successfully Executed!"
	return response
}

func insertToPostgres(request *messaging.ETLRequest, session *sql.DB) messaging.ETLResponse {
	var response messaging.ETLResponse
	if request.Body.Object == nil {
		response = insertMultiple(request, session)
	} else {
		response = insertSingle(request, session)
	}
	return response
}

func deleteFromPostgres(request *messaging.ETLRequest, session *sql.DB) messaging.ETLResponse {
	var response messaging.ETLResponse
	if request.Body.Object == nil {
		response = deleteMultiple(request, session)
	} else {
		response = deleteSingle(request, session)
	}
	return response
}

func updateOnPostgres(request *messaging.ETLRequest, session *sql.DB) messaging.ETLResponse {
	var response messaging.ETLResponse
	if request.Body.Object == nil {
		response = updateMultiple(request, session)
	} else {
		response = updateSingle(request, session)
	}
	return response
}

func getPostgresReportingConnection(request *messaging.ETLRequest) (session *sql.DB, isError bool, errorMessage string) {
	isError = false
	username := request.Configuration.EtlConfig["POSTGRESv5"]["Username"]
	password := request.Configuration.EtlConfig["POSTGRESv5"]["Password"]
	dbUrl := request.Configuration.EtlConfig["POSTGRESv5"]["Url"]
	dbPort := request.Configuration.EtlConfig["POSTGRESv5"]["Port"]

	session, err := sql.Open("postgres", "host="+dbUrl+" port="+dbPort+" user="+username+" password="+password+" dbname="+"postgres"+" sslmode=disable")

	if err != nil {
		isError = true
		logger.Log("There is an error")
		errorMessage = err.Error()
		logger.Log("Postgres connection initilizing failed!")
	}

	//Create schema if not available.
	logger.Log("Checking if Database " + getPostgresReportingSQLnamespace(request) + " is available.")

	isDatabaseAvailbale := false

	rows, err := session.Query("SELECT datname FROM pg_database WHERE datistemplate = false;")

	if err != nil {
		logger.Log(err.Error())
		logger.Log("Error contacting PostGres Server")
	} else {
		logger.Log("Successfully retrieved values for all objects in PostGres")

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
				logger.Log("Check domain : " + getPostgresReportingSQLnamespace(request) + " : available schema : " + v.(string))
				if v.(string) == getPostgresReportingSQLnamespace(request) {
					//Database available
					isDatabaseAvailbale = true
					break
				}
			}
		}
	}

	if isDatabaseAvailbale {
		logger.Log("Database already available. Nothing to do. Proceed!")
		session.Close()
		session, err = sql.Open("postgres", "host="+dbUrl+" port="+dbPort+" user="+username+" password="+password+" dbname="+(getPostgresReportingSQLnamespace(request))+" sslmode=disable")

	} else {
		_, err = session.Query("CREATE DATABASE " + getPostgresReportingSQLnamespace(request) + ";")
		if err != nil {
			logger.Log("Creation of domain matched Schema failed")
		} else {
			logger.Log("Creation of domain matched Schema Successful")
			session.Close()
			session, err = sql.Open("postgres", "host="+dbUrl+" port="+dbPort+" user="+username+" password="+password+" dbname="+(getPostgresReportingSQLnamespace(request))+" sslmode=disable")
			if err != nil {
				logger.Log("Relogin Failed!")
			} else {
				logger.Log("Relogin successful!")
			}

		}
	}
	logger.Log("Reusing existing Postgres connection")
	return
}

func insertMultiple(request *messaging.ETLRequest, session *sql.DB) messaging.ETLResponse {
	logger.Log("Starting INSERT-MULTIPLE")
	var response messaging.ETLResponse

	//update table with Namespace field

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
		DataObjects[i]["namespace"] = request.Controls.Namespace
	}

	//create dumb Request object

	var dumbRequest messaging.ETLRequest
	dumbRequest.Controls = request.Controls
	dumbRequest.Body = request.Body
	dumbRequest.Configuration = request.Configuration
	dumbRequest.Body.Objects = DataObjects

	pointerDumbRequest := &dumbRequest

	//check for table in postgres

	if createPostgresReportingTable(pointerDumbRequest, session) {
		logger.Log("Table Verified Successfully!")
	} else {
		response.IsSuccess = false
		return response
	}

	indexNames := getPostgresReportingFieldOrder(request, session)
	fmt.Print("Index Names : ")
	fmt.Println(indexNames)

	var argKeyList string
	var argValueList string

	//create keyvalue list

	// for i := 0; i < len(indexNames); i++ {
	// 	if i != len(indexNames)-1 {
	// 		argKeyList = argKeyList + indexNames[i] + ", "
	// 	} else {
	// 		argKeyList = argKeyList + indexNames[i]
	// 	}
	// }
	//fmt.Println(len(indexNames))
	for i := 0; i < len(DataObjects); i++ {
		noOfElements := len(DataObjects[i])
		//fmt.Println("---------------")
		////fmt.Println(DataObjects[i])
		//fmt.Println("---------------")
		//fmt.Println(len(DataObjects[i]))
		//fmt.Println(noOfElements)
		keyValue := getPostgresReportingSqlRecordID(request, DataObjects[i], session)
		DataObjects[i][strings.ToLower(request.Body.Parameters.KeyProperty)] = keyValue
		if keyValue == "" {
			response.IsSuccess = false
			response.Message = "Failed inserting multiple object in Cassandra"
			logger.Log(response.Message)
			logger.Log("Inavalid ID request")
			return response
		}

		var keyArray = make([]string, noOfElements)
		var valueArray = make([]string, noOfElements)

		arrlength := 0
		if len(DataObjects[i]) < len(indexNames) {
			arrlength = len(DataObjects[i])
			var reducedIndexes []string
			reducedIndexes = make([]string, len(DataObjects[i]))
			index := 0
			/*for key, _ := range DataObject {
				for x := 0; x < len(indexNames); x++ {
					if strings.ToLower(key) == indexNames[x] {
						reducedIndexes[index] = indexNames[x]
						index++
						break
					}
				}
			}*/

			for x := 0; x < len(indexNames); x++ {
				for key, _ := range DataObjects[i] {
					if strings.ToLower(key) == indexNames[x] {
						reducedIndexes[index] = indexNames[x]
						index++
						break
					}
				}
			}

			fmt.Print("Reduced Indexes : ")
			fmt.Println(reducedIndexes)
			indexNames = reducedIndexes
			fmt.Print(indexNames)
			arrlength = len(indexNames)
			for i := 0; i < len(indexNames); i++ {
				if i != len(indexNames)-1 {
					argKeyList = argKeyList + indexNames[i] + ", "
				} else {
					argKeyList = argKeyList + indexNames[i]
				}
			}

		} else {
			arrlength = len(indexNames)
			for i := 0; i < len(indexNames); i++ {
				if i != len(indexNames)-1 {
					argKeyList = argKeyList + indexNames[i] + ", "
				} else {
					argKeyList = argKeyList + indexNames[i]
				}
			}
		}

		for index := 0; index < arrlength; index++ {
			//fmt.Println(indexNames[index])
			if indexNames[index] != "osheaders" {
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
	fmt.Println("INSERT INTO " + request.Controls.Class + " (" + argKeyList + ") VALUES " + argValueList + ";")
	_, err := session.Query("INSERT INTO " + request.Controls.Class + " (" + argKeyList + ") VALUES " + argValueList + ";")
	if err != nil {
		response.IsSuccess = false
		response.Message = "Error inserting Many object in Postgres" + err.Error()
	} else {
		response.IsSuccess = true
		response.Message = "Successfully inserted Many object in Postgres"
		logger.Log(response.Message)
	}
	//session.Close()
	return response
}

func insertSingle(request *messaging.ETLRequest, session *sql.DB) messaging.ETLResponse {
	logger.Log("Starting INSERT-SINGLE")
	var response messaging.ETLResponse

	keyValue := getPostgresReportingSqlRecordID(request, nil, session)

	if keyValue != "" {
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

		DataObject["namespace"] = request.Controls.Namespace

		noOfElements := len(DataObject)
		DataObject[strings.ToLower(request.Body.Parameters.KeyProperty)] = keyValue

		//create dumb Request object

		var dumbRequest messaging.ETLRequest
		dumbRequest.Controls = request.Controls
		dumbRequest.Body = request.Body
		dumbRequest.Configuration = request.Configuration
		dumbRequest.Body.Object = DataObject

		pointerDumbRequest := &dumbRequest

		if createPostgresReportingTable(pointerDumbRequest, session) {
			logger.Log("Table Verified Successfully!")
		} else {
			response.IsSuccess = false
			return response
		}

		indexNames := getPostgresReportingFieldOrder(request, session)
		fmt.Print("Index Names : ")
		fmt.Println(indexNames)
		var argKeyList string
		var argValueList string

		//create keyvalue list

		/*for i := 0; i < len(indexNames); i++ {
			if i != len(indexNames)-1 {
				argKeyList = argKeyList + indexNames[i] + ", "
			} else {
				argKeyList = argKeyList + indexNames[i]
			}
		}*/

		var keyArray = make([]string, noOfElements)
		var valueArray = make([]string, noOfElements)
		fmt.Println("----------------")
		fmt.Println(len(indexNames))
		fmt.Println(len(DataObject))
		fmt.Println("----------------")

		arrlength := 0
		if len(DataObject) < len(indexNames) {
			arrlength = len(DataObject)
			var reducedIndexes []string
			reducedIndexes = make([]string, len(DataObject))
			index := 0
			/*for key, _ := range DataObject {
				for x := 0; x < len(indexNames); x++ {
					if strings.ToLower(key) == indexNames[x] {
						reducedIndexes[index] = indexNames[x]
						index++
						break
					}
				}
			}*/

			for x := 0; x < len(indexNames); x++ {
				for key, _ := range DataObject {
					if strings.ToLower(key) == indexNames[x] {
						reducedIndexes[index] = indexNames[x]
						index++
						break
					}
				}
			}

			fmt.Print("Reduced Indexes : ")
			fmt.Println(reducedIndexes)
			indexNames = reducedIndexes
			fmt.Print(indexNames)
			arrlength = len(indexNames)
			for i := 0; i < len(indexNames); i++ {
				if i != len(indexNames)-1 {
					argKeyList = argKeyList + indexNames[i] + ", "
				} else {
					argKeyList = argKeyList + indexNames[i]
				}
			}

		} else {
			arrlength = len(indexNames)
			for i := 0; i < len(indexNames); i++ {
				if i != len(indexNames)-1 {
					argKeyList = argKeyList + indexNames[i] + ", "
				} else {
					argKeyList = argKeyList + indexNames[i]
				}
			}
		}
		// Process A :start identifying individual data in array and convert to string
		for index := 0; index < arrlength; index++ {
			if indexNames[index] != "osheaders" {

				if _, ok := DataObject[indexNames[index]].(string); ok {
					//fmt.Println(indexNames[index])
					//fmt.Println(valueArray[index])
					keyArray[index] = indexNames[index]
					valueArray[index] = DataObject[indexNames[index]].(string)
				} else {
					//fmt.Println(indexNames[index])
					//fmt.Println(valueArray[index])
					//fmt.Println("Non string value detected, Will be strigified!")
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
		fmt.Println("INSERT INTO " + request.Controls.Class + " (" + argKeyList + ") VALUES (" + argValueList + ")")
		_, err := session.Query("INSERT INTO " + request.Controls.Class + " (" + argKeyList + ") VALUES (" + argValueList + ")")
		if err != nil {
			response.IsSuccess = false
			response.Message = "Error inserting one object in Postgres" + err.Error()
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted one object in Postgres"
			logger.Log(response.Message)
		}
	}
	//session.Close()
	return response
}

func deleteMultiple(request *messaging.ETLRequest, session *sql.DB) messaging.ETLResponse {
	logger.Log("Starting DELETE-MULTIPLE")
	var response messaging.ETLResponse

	for _, obj := range request.Body.Objects {
		_, err := session.Query("DELETE FROM " + strings.ToLower(request.Controls.Class) + " WHERE " + strings.ToLower(request.Body.Parameters.KeyProperty) + " = '" + obj[request.Body.Parameters.KeyProperty].(string) + "'")
		if err != nil {
			response.IsSuccess = false
			logger.Log("Error deleting object in Postgres : " + err.Error())
			response.Message = "Error deleting one object in Postgres because no match was found!" + err.Error()
		} else {
			response.IsSuccess = true
			response.Message = "Successfully deleted one object in Postgres"
			logger.Log(response.Message)
		}
	}

	//session.Close()
	return response
}

func deleteSingle(request *messaging.ETLRequest, session *sql.DB) messaging.ETLResponse {
	logger.Log("Starting DELETE-SINGLE")
	var response messaging.ETLResponse

	_, err := session.Query("DELETE FROM " + strings.ToLower(request.Controls.Class) + " WHERE " + strings.ToLower(request.Body.Parameters.KeyProperty) + " = '" + request.Controls.Id + "'")
	if err != nil {
		response.IsSuccess = false
		logger.Log("Error deleting object in Postgres : " + err.Error())
		response.Message = "Error deleting one object in Postgres because no match was found!" + err.Error()
	} else {
		response.IsSuccess = true
		response.Message = "Successfully deleted one object in Postgres"
		logger.Log(response.Message)
	}

	//	session.Close()
	return response
}

func updateMultiple(request *messaging.ETLRequest, session *sql.DB) messaging.ETLResponse {
	logger.Log("Starting UPDATE-MULTIPLE")
	var response messaging.ETLResponse

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
		DataObjects[i]["namespace"] = request.Controls.Namespace
	}

	//create dumb Request object

	var dumbRequest messaging.ETLRequest
	dumbRequest.Controls = request.Controls
	dumbRequest.Body = request.Body
	dumbRequest.Configuration = request.Configuration
	dumbRequest.Body.Objects = DataObjects

	pointerDumbRequest := &dumbRequest

	//check for table in postgres

	if createPostgresReportingTable(pointerDumbRequest, session) {
		logger.Log("Table Verified Successfully!")
	} else {
		response.IsSuccess = false
		return response
	}

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
					logger.Log("Non string value detected, Will be strigified!")
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
		_, err := session.Query("UPDATE " + request.Controls.Class + " SET " + argValueList + " WHERE " + strings.ToLower(request.Body.Parameters.KeyProperty) + " =" + "'" + request.Body.Objects[i][request.Body.Parameters.KeyProperty].(string) + "'")
		//err := collection.Update(bson.M{key: value}, bson.M{"$set": request.Body.Object})
		if err != nil {
			response.IsSuccess = false
			logger.Log("Error updating object in Postgres  : " + request.Controls.Id + ", " + err.Error())
			response.Message = "Error updating one object in Postgres because no match was found!" + err.Error()
		} else {
			response.IsSuccess = true
			response.Message = "Successfully updating one object in Postgres "
			logger.Log(response.Message)
		}
	}

	//session.Close()
	return response
}

func updateSingle(request *messaging.ETLRequest, session *sql.DB) messaging.ETLResponse {
	logger.Log("Starting UPDATE-SINGLE")
	var response messaging.ETLResponse

	var DataObject map[string]interface{}
	DataObject = make(map[string]interface{})

	for key, value := range request.Body.Object {
		if key == "__osHeaders" {
			DataObject["osheaders"] = value
		} else {
			DataObject[strings.ToLower(key)] = value
		}
	}

	DataObject["namespace"] = request.Controls.Namespace

	//create dumb Request object

	var dumbRequest messaging.ETLRequest
	dumbRequest.Controls = request.Controls
	dumbRequest.Body = request.Body
	dumbRequest.Configuration = request.Configuration

	dumbRequest.Body.Object = DataObject

	pointerDumbRequest := &dumbRequest

	if createPostgresReportingTable(pointerDumbRequest, session) {
		logger.Log("Table Verified Successfully!")
	} else {
		response.IsSuccess = false
		return response
	}

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
				logger.Log("Non string value detected, Will be strigified!")
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

	_, err := session.Query("UPDATE " + strings.ToLower(request.Controls.Class) + " SET " + argValueList + " WHERE " + strings.ToLower(request.Body.Parameters.KeyProperty) + " =" + "'" + request.Controls.Id + "'")
	//err := collection.Update(bson.M{key: value}, bson.M{"$set": request.Body.Object})
	if err != nil {
		response.IsSuccess = false
		logger.Log("Error updating object in Postgres  : " + request.Controls.Id + ", " + err.Error())
		response.Message = "Error updating one object in Postgres because no match was found!" + err.Error()
	} else {
		response.IsSuccess = true
		response.Message = "Successfully updating one object in Postgres "
		logger.Log(response.Message)
	}

	//session.Close()
	return response
}

//Helper Functions

func createPostgresReportingTable(request *messaging.ETLRequest, session *sql.DB) (status bool) {
	status = false

	//get table list
	classBytes := executePostgresReportingGetClasses(session)
	var classList []string
	err := json.Unmarshal(classBytes, &classList)
	if err != nil {
		status = false
	} else {
		fmt.Print("Available Table List : ")
		fmt.Println(classList)
		fmt.Print("Needed Class : ")
		fmt.Println(request.Controls.Class)
		for _, className := range classList {
			if strings.ToLower(request.Controls.Class) == className {
				fmt.Println("Table Already Available")
				status = true
				//Get all fields
				classBytes := executePostgresReportingGetFields(session, strings.ToLower(request.Controls.Class))
				var tableFieldList []string
				_ = json.Unmarshal(classBytes, &tableFieldList)
				fmt.Print("Fields from DB : ")
				fmt.Println(tableFieldList)
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
							if value != nil {
								recordFieldList[index] = strings.ToLower(key)
								recordFieldType[index] = getDataType(value)
							} else {
								fmt.Println("Nil value found at : " + className + " @ " + key)
								recordFieldList[index] = strings.ToLower(key)
								recordFieldType[index] = "text"
							}
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
							if value != nil {
								recordFieldList[index] = strings.ToLower(key)
								recordFieldType[index] = getDataType(value)
							} else {
								fmt.Println("Nil value found at : " + className + " @ " + key)
								recordFieldList[index] = strings.ToLower(key)
								recordFieldType[index] = "text"
							}
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

				fmt.Print("New Fields :")
				fmt.Println(newFields)

				//ALTER TABLES

				for key, _ := range newFields {
					_, er := session.Query("ALTER TABLE " + strings.ToLower(request.Controls.Class) + " ADD COLUMN " + newFields[key] + " " + newTypes[key] + ";")
					if er != nil {
						status = false
						logger.Log("Table Alter Failed : " + er.Error())
						return
					} else {
						status = true
						logger.Log("Table Alter Success!")
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
		noOfElements := len(dataObject) + 1
		var keyArray = make([]string, noOfElements)
		var dataTypeArray = make([]string, noOfElements)

		var startIndex int = 0

		for key, value := range dataObject {
			//fmt.Print(key + " : ")
			//fmt.Println(value)sss
			if value != nil {
				keyArray[startIndex] = key
				dataTypeArray[startIndex] = getDataType(value)
			} else {
				fmt.Println("Nil value found at : " + key)
				keyArray[startIndex] = key
				dataTypeArray[startIndex] = "text"
			}
			startIndex = startIndex + 1

		}

		keyArray[noOfElements-1] = "incrementkey"
		dataTypeArray[noOfElements-1] = "SERIAL"

		//Create Table

		var argKeyList2 string

		for i := 0; i < noOfElements; i++ {
			if i != noOfElements-1 {
				if keyArray[i] == strings.ToLower(request.Body.Parameters.KeyProperty) {
					argKeyList2 = argKeyList2 + keyArray[i] + " text PRIMARY KEY, "
				} else {
					argKeyList2 = argKeyList2 + keyArray[i] + " " + dataTypeArray[i] + ", "
				}

			} else {
				if keyArray[i] == strings.ToLower(request.Body.Parameters.KeyProperty) {
					argKeyList2 = argKeyList2 + keyArray[i] + " text PRIMARY KEY"
				} else {
					argKeyList2 = argKeyList2 + keyArray[i] + " " + dataTypeArray[i]
				}

			}
		}

		logger.Log("create table " + strings.ToLower(request.Controls.Class) + "(" + argKeyList2 + ");")

		_, er := session.Query("create table " + strings.ToLower(request.Controls.Class) + "(" + argKeyList2 + ");")
		if er != nil {
			status = false
			logger.Log("Table Creation Failed : " + er.Error())
			return
		}

		status = true

	}

	return
}

func executePostgresReportingGetClasses(session *sql.DB) (returnByte []byte) {

	var returnMap map[string]interface{}
	returnMap = make(map[string]interface{})

	rows, err := session.Query(" SELECT table_name FROM information_schema.tables WHERE table_schema='public';")

	if err != nil {
		logger.Log("Error executing query in Postgres SQL")
	} else {
		logger.Log("Successfully executed query in Postgres SQL")
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
			logger.Log("Error converting to byte array")
			byteValue = nil
		} else {
			logger.Log("Successfully converted result to byte array")
		}

		returnByte = byteValue
	}

	//session.Close()
	return returnByte
}

func executePostgresReportingGetFields(session *sql.DB, class string) (returnByte []byte) {

	var returnMap map[string]interface{}
	returnMap = make(map[string]interface{})

	rows, err := session.Query("select column_name from information_schema.columns where table_name='" + class + "';")

	if err != nil {
		logger.Log("Error executing query in Postgres SQL")
	} else {
		logger.Log("Successfully executed query in Postgres SQL")
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
			logger.Log("Error converting to byte array")
			byteValue = nil
		} else {
			logger.Log("Successfully converted result to byte array")
		}

		returnByte = byteValue
	}

	//session.Close()
	return returnByte
}

// func getPostgresReportingSQLnamespace(request *messaging.ETLRequest) string {
// 	return strings.ToLower(request.Configuration.EtlConfig["POSTGRESv5"]["DatabaseName"])
// }

func getPostgresReportingSQLnamespace(request *messaging.ETLRequest) string {
	//return ("_" + strings.Replace(request.Controls.Namespace, ".", "", -1))
	return "epayreportingv6"
}

func getPostgresReportingFieldOrder(request *messaging.ETLRequest, session *sql.DB) []string {
	var returnArray []string
	class := strings.ToLower(request.Controls.Class)
	//read fields
	fmt.Println("Reading Fields From : " + class)
	byteValue := executePostgresReportingGetFields(session, class)

	err := json.Unmarshal(byteValue, &returnArray)
	//	fmt.Println(returnArray)
	var fieldArray []string
	fieldArray = make([]string, len(returnArray)-1)
	index := 0
	for _, value := range returnArray {
		if value != "incrementkey" {
			fieldArray[index] = value
			index++
		}
	}
	if err != nil {
		logger.Log("Converstion of Json Failed!")
		fieldArray = make([]string, 1)
		fieldArray[0] = "nil"
		return fieldArray
	}

	return fieldArray
}

func getPostgresReportingSqlRecordID(request *messaging.ETLRequest, obj map[string]interface{}, session *sql.DB) (returnID string) {
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
		logger.Log("GUID Key generation requested!")
		returnID = uuid.NewV1().String()
	} else if isAutoIncrementId {
		logger.Log("Automatic Increment Key generation requested!")

		//Read Table domainClassAttributes
		logger.Log("Reading maxCount from DB")
		rows, err := session.Query("SELECT maxCount FROM domainClassAttributes where class = '" + request.Controls.Class + "';")

		if err != nil {
			//If err create new domainClassAttributes  table
			logger.Log("No Class found.. Must be a new namespace")
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
				logger.Log("New Class! New record for this class will be inserted")
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
				logger.Log("Record Available!")
				maxCount := 0
				maxCount, err = strconv.Atoi(myMap["maxcount"].(string))
				maxCount++
				returnID = strconv.Itoa(maxCount)
				_, err = session.Query("UPDATE domainClassAttributes SET maxcount='" + returnID + "' WHERE class = '" + request.Controls.Class + "' ;")
				if err != nil {
					logger.Log("Error Updating index table : " + err.Error())
					returnID = ""
					return
				}
			}
		}

		//session.Close()
	} else {
		logger.Log("Manual Key requested!")
		if obj == nil {
			returnID = request.Controls.Id
		} else {
			returnID = obj[strings.ToLower(request.Body.Parameters.KeyProperty)].(string)
		}
	}

	return
}
