package repositories

import (
	"duov6.com/common"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/queryparser"
	"duov6.com/term"
	"encoding/base64"
	"encoding/json"
	"errors"
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

func (repository HiveRepository) getDatabaseName(request *messaging.ObjectRequest) string {
	return strings.Replace(strings.ToLower(request.Controls.Namespace), ".", "", -1)
}

func (repository HiveRepository) getConnection(request *messaging.ObjectRequest) (conn *hive.HiveConnection, isError bool, errorMessage string) {
	isError = false
	hive.MakePool(request.Configuration.ServerConfiguration["HIVE"]["Host"] + ":" + request.Configuration.ServerConfiguration["HIVE"]["Port"])
	conn, err := hive.GetHiveConn()
	if err != nil {
		isError = true
		errorMessage = err.Error()
		term.Write("HIVE connection initilizing failed! : "+errorMessage, 1)
	} else {
		isError = false
		term.Write("HIVE connection initilizing Successful!", 2)
	}
	return
}

func (repository HiveRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Get-All!", 2)
	query := "SELECT * FROM " + repository.getDatabaseName(request) + "." + strings.ToLower(request.Controls.Class)
	return repository.queryCommonMany(query, request)
}

func (repository HiveRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Get-Search!", 2)
	response := RepositoryResponse{}
	query := ""
	if strings.Contains(request.Body.Query.Parameters, ":") {
		tokens := strings.Split(request.Body.Query.Parameters, ":")
		fieldName := tokens[0]
		fieldValue := tokens[1]
		fieldName = strings.TrimSpace(fieldName)
		fieldValue = strings.TrimSpace(fieldValue)
		query = "select * from " + repository.getDatabaseName(request) + "." + strings.ToLower(request.Controls.Class) + " where " + fieldName + "='" + fieldValue + "';"
	} else {
		query = "select * from " + repository.getDatabaseName(request) + "." + strings.ToLower(request.Controls.Class) + ";"
	}
	request.Log(query)
	response = repository.queryCommonMany(query, request)
	return response
}

func (repository HiveRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-QUERY")
	response := RepositoryResponse{}
	queryType := request.Body.Query.Type

	switch queryType {
	case "Query":
		if request.Body.Query.Parameters != "*" {
			request.Log("USER INPUT QUERY : " + request.Body.Query.Parameters)
			formattedQuery := queryparser.GetFormattedQuery(request.Body.Query.Parameters)
			request.Log("HIVE QUERY : " + formattedQuery)
			return repository.queryCommonMany(formattedQuery, request)
		} else {
			return repository.GetAll(request)
		}
	default:
		request.Log(queryType + " not implemented in Hive Db repository")
		return getDefaultNotImplemented()

	}

	return response
}

func (repository HiveRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Get-By-Key!", 2)
	query := "SELECT * FROM " + repository.getDatabaseName(request) + "." + strings.ToLower(request.Controls.Class) + " WHERE osid = '" + getNoSqlKey(request) + "'"
	return repository.queryCommonOne(query, request)
}

func (repository HiveRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Insert-Multiple!", 2)
	var idData map[string]interface{}
	idData = make(map[string]interface{})

	for index, obj := range request.Body.Objects {
		id := repository.getRecordID(request, obj)
		idData[strconv.Itoa(index)] = id
		request.Body.Objects[index][request.Body.Parameters.KeyProperty] = id
	}

	var DataMap []map[string]interface{}
	DataMap = make([]map[string]interface{}, 1)
	var idMap map[string]interface{}
	idMap = make(map[string]interface{})
	idMap["ID"] = idData
	DataMap[0] = idMap

	response := repository.queryStore(request)
	response.Data = DataMap
	return response
}

func (repository HiveRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Insert-Single!", 2)
	id := repository.getRecordID(request, request.Body.Object)
	request.Controls.Id = id
	request.Body.Object[request.Body.Parameters.KeyProperty] = id

	//Add IDs to return Data
	var Data []map[string]interface{}
	Data = make([]map[string]interface{}, 1)
	var idData map[string]interface{}
	idData = make(map[string]interface{})
	idData["ID"] = id
	Data[0] = idData

	response := repository.queryStore(request)
	response.Data = Data
	return response
}

func (repository HiveRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Update-Multiple!", 2)
	return repository.queryStore(request)
}

func (repository HiveRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Update-Single!", 2)
	return repository.queryStore(request)
}

func (repository HiveRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Delete-Multiple!", 2)
	response := RepositoryResponse{}
	conn, _, _ := repository.getConnection(request)

	isError := false
	for _, obj := range request.Body.Objects {
		query := repository.getDeleteScript(request, strings.ToLower(request.Controls.Class), getNoSqlKeyById(request, obj))
		err := repository.executeNonQuery(conn, query)
		if err != nil {
			isError = true
		}
	}
	if isError {
		response.IsSuccess = false
		response.Message = "Error deleting all objects. Please double check data!"
	} else {
		response.IsSuccess = true
		response.Message = "Successfully Deleted all objects from Hive repository!"
	}

	return response
}

func (repository HiveRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Delete-Single!", 2)
	response := RepositoryResponse{}
	conn, _, _ := repository.getConnection(request)

	query := repository.getDeleteScript(request, strings.ToLower(request.Controls.Class), getNoSqlKey(request))
	err := repository.executeNonQuery(conn, query)
	if err != nil {
		response.IsSuccess = false
		response.Message = "Failed Deleting from Hive repository : " + err.Error()
	} else {
		response.IsSuccess = true
		response.Message = "Successfully Deleted from Hive repository!"
	}

	return response
}

func (repository HiveRepository) getDeleteScript(request *messaging.ObjectRequest, class string, id string) string {
	return "DELETE FROM " + repository.getDatabaseName(request) + "." + class + " WHERE osid = '" + id + "'"
}

func (repository HiveRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	request.Log("Starting SPECIAL!")
	queryType := request.Body.Special.Type

	conn, isError, errorMessage := repository.getConnection(request)
	if isError {
		response.IsSuccess = isError
		response.Message = errorMessage
		response.GetErrorResponse(errorMessage)
		return response
	}

	switch queryType {
	case "getFields":
		request.Log("Starting GET-FIELDS sub routine!")
		fieldsInByte := repository.executeGetFields(request, conn)
		if fieldsInByte != nil {
			response.IsSuccess = true
			response.Message = "Successfully Retrieved Fileds on Class : " + strings.ToLower(request.Controls.Class)
			response.GetResponseWithBody(fieldsInByte)
		} else {
			response.IsSuccess = false
			response.Message = "Aborted! Unsuccessful Retrieving Fileds on Class : " + strings.ToLower(request.Controls.Class)
			errorMessage := response.Message
			response.GetErrorResponse(errorMessage)
		}
	case "getDataTypes":
		request.Log("Starting GET-Data-Types sub routine!")
		fieldsInByte := repository.executeGetDataTypes(request, conn)
		if fieldsInByte != nil {
			response.IsSuccess = true
			response.Message = "Successfully Retrieved Data Types on Class : " + strings.ToLower(request.Controls.Class)
			response.GetResponseWithBody(fieldsInByte)
		} else {
			response.IsSuccess = false
			response.Message = "Aborted! Unsuccessful Retrieving Data Types on Class : " + strings.ToLower(request.Controls.Class)
			errorMessage := response.Message
			response.GetErrorResponse(errorMessage)
		}
	case "getClasses":
		request.Log("Starting GET-CLASSES sub routine")
		fieldsInByte := repository.executeGetClasses(request, conn)
		if fieldsInByte != nil {
			response.IsSuccess = true
			response.Message = "Successfully Retrieved Fileds on Class : " + strings.ToLower(request.Controls.Class)
			response.GetResponseWithBody(fieldsInByte)
		} else {
			response.IsSuccess = false
			response.Message = "Aborted! Unsuccessful Retrieving Fileds on Class : " + strings.ToLower(request.Controls.Class)
			errorMessage := response.Message
			response.GetErrorResponse(errorMessage)
		}
	case "getNamespaces":
		request.Log("Starting GET-NAMESPACES sub routine")
		fieldsInByte := repository.executeGetNamespaces(request, conn)
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
		fieldsInByte := repository.executeGetSelected(request, conn)
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
		return repository.GetAll(request)

	}

	return response
}

func (repository HiveRepository) Test(request *messaging.ObjectRequest) {

}

//Sub Routines

func (repository HiveRepository) executeGetFields(request *messaging.ObjectRequest, conn *hive.HiveConnection) (returnByte []byte) {

	namespace := repository.getDatabaseName(request)
	class := strings.ToLower(request.Controls.Class)

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

	return
}
func (repository HiveRepository) executeGetDataTypes(request *messaging.ObjectRequest, conn *hive.HiveConnection) (returnByte []byte) {

	namespace := repository.getDatabaseName(request)
	class := strings.ToLower(request.Controls.Class)

	er, err := conn.Client.Execute("describe " + namespace + "." + class)
	if er == nil && err == nil {

		//Get Schema
		schema, _, _ := conn.Client.GetSchema()

		var allMaps []map[string]interface{}

		for {
			row, _, _ := conn.Client.FetchOne()
			if row == "" {
				break
			} else {
				var myMap map[string]interface{}
				myMap = make(map[string]interface{})

				keyValue := ""
				valueValue := ""

				temp := strings.Split(row, "\t")
				index := 1
				for key, value := range temp {

					value = strings.TrimSpace(value)
					if !repository.isEven(index) {
						if value != "" && value != " " && schema.FieldSchemas[key].Name == "col_name" {
							keyValue = value
						}
					} else {
						if value != "" && value != " " && schema.FieldSchemas[key].Name == "data_type" {
							valueValue = value
						}
						myMap[keyValue] = valueValue
						allMaps = append(allMaps, myMap)
					}
					index++
				}
			}
		}

		byteValue, errMarshal := json.Marshal(allMaps)

		if errMarshal != nil {
			byteValue = nil
		}

		returnByte = byteValue
	} else {
		returnByte = nil
	}

	return
}

func (repository HiveRepository) isEven(value int) (response bool) {
	if value%2 == 0 {
		response = true
	} else {
		response = false
	}
	return
}

func (repository HiveRepository) executeGetClasses(request *messaging.ObjectRequest, conn *hive.HiveConnection) (returnByte []byte) {

	namespace := repository.getDatabaseName(request)

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

	return
}

func (repository HiveRepository) executeGetNamespaces(request *messaging.ObjectRequest, conn *hive.HiveConnection) (returnByte []byte) {

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

	return
}

func (repository HiveRepository) executeGetSelected(request *messaging.ObjectRequest, conn *hive.HiveConnection) (returnByte []byte) {

	tableName := repository.getDatabaseName(request) + "." + strings.ToLower(request.Controls.Class)

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

	return
}

func (repository HiveRepository) getRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
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

func (repository HiveRepository) verifySchema(conn *hive.HiveConnection, request *messaging.ObjectRequest) {
	repository.verifyDatabase(conn, repository.getDatabaseName(request))
	if request.Body.Object != nil {
		repository.verifyTable(request, conn, request.Body.Object)
	} else {
		repository.verifyTable(request, conn, request.Body.Objects[0])
	}
}

func (repository HiveRepository) verifyDatabase(conn *hive.HiveConnection, database string) {
	query := "create database if not exists " + database
	repository.executeNonQuery(conn, query)
}

func (repository HiveRepository) verifyTable(request *messaging.ObjectRequest, conn *hive.HiveConnection, obj map[string]interface{}) {
	table := repository.getDatabaseName(request) + "." + strings.ToLower(request.Controls.Class)
	//get table list
	isTableAvailable := false
	var tableList []string
	tableBytes := repository.executeGetClasses(request, conn)
	_ = json.Unmarshal(tableBytes, &tableList)

	fmt.Print("Table List : ")
	fmt.Println(tableList)

	for _, tableName := range tableList {
		if strings.ToLower(request.Controls.Class) == strings.ToLower(tableName) {
			isTableAvailable = true
			break
		}
	}
	//if available get fields list
	if !isTableAvailable {
		query := "create table IF NOT EXISTS " + table + " (osid string"

		for k, v := range obj {
			if k != "OriginalIndex" {
				if k != "__osHeaders" {
					query += ("," + k + " " + repository.golangToSql(v))
				} else {
					query += ("," + "osheaders" + " " + repository.golangToSql(v))
				}
			}
		}

		query += ") clustered by (osid) into " + strconv.Itoa((len(obj) + 1)) + " buckets stored as orc TBLPROPERTIES ('transactional'='true')"
		repository.executeNonQuery(conn, query)
	} else {
		//get fields list...
		var fieldList []string
		fieldBytes := repository.executeGetFields(request, conn)
		_ = json.Unmarshal(fieldBytes, &fieldList)

		//check for new fields and alter the table
		fieldString := ""
		for _, fieldFromDb := range fieldList {
			if fieldFromDb != "__osHeaders" {
				fieldString += ("|" + strings.ToLower(fieldFromDb))
			} else {
				fieldString += ("|" + "osheaders")
			}
		}

		for fieldFromObj, value := range obj {
			customFieldName := ""
			if fieldFromObj == "__osHeaders" {
				customFieldName = "osheaders"
			} else if fieldFromObj == "__os__id" {
				customFieldName = "osid"
			} else {
				customFieldName = fieldFromObj
			}

			if !strings.Contains(fieldString, strings.ToLower(customFieldName)) {
				alterQuery := "ALTER TABLE " + table + " ADD COLUMNS (" + customFieldName + " " + repository.golangToSql(value) + ");"
				repository.executeNonQuery(conn, alterQuery)
			}
		}
	}

}

func (repository HiveRepository) executeNonQuery(conn *hive.HiveConnection, query string) (err error) {
	fmt.Println(query)
	common.PublishLog("ObjectStoreLog.log", query)
	_, err = conn.Client.Execute(query)
	if err != nil {
		term.Write(err.Error(), 1)
	}
	return
}

func (repository HiveRepository) golangToSql(value interface{}) string {
	var strValue string

	//fmt.Println(reflect.TypeOf(value))
	switch value.(type) {
	case string:
		strValue = "string"
	case bool:
		strValue = "boolean"
		break
	case uint:
	case int:
	//case uintptr:
	case uint8:
	case uint16:
	case uint32:
	case uint64:
	case int8:
	case int16:
	case int32:
	case int64:
		strValue = "int"
		break
	case float32:
	case float64:
		strValue = "double"
		break
	default:
		strValue = "binary"
		break

	}

	return strValue
}

func (repository HiveRepository) getSqlFieldValue(value interface{}) string {
	var strValue string
	switch v := value.(type) {
	case bool:
		if value.(bool) == true {
			strValue = "true"
		} else {
			strValue = "false"
		}
		break
	case string:
		sval := fmt.Sprint(value)
		if strings.ContainsAny(sval, "\"'\n\r\t") {
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

func (repository HiveRepository) getInterfaceValue(tmp string) (outData interface{}) {
	var m interface{}
	if string(tmp[0]) == "{" || string(tmp[0]) == "[" {
		err := json.Unmarshal([]byte(tmp), &m)
		if err == nil {
			outData = m
		} else {
			term.Write(err.Error(), 1)
			outData = tmp
		}
	} else {
		outData = tmp
	}
	return
}

func (repository HiveRepository) queryStore(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	conn, _, _ := repository.getConnection(request)

	script, err := repository.getStoreScript(conn, request)

	queryArray := strings.Split(script, "###")

	if len(queryArray) > 1 && err == nil {
		//Multiple Updates
		status := make([]bool, len(queryArray)-1)
		for index := 0; index < (len(queryArray) - 1); index++ {
			err := repository.executeNonQuery(conn, queryArray[index])
			if err == nil {
				status[index] = true
			} else {
				status[index] = false
			}
		}
		for _, stat := range status {
			if stat == false {
				response.IsSuccess = false
				response.Message = "Error Updating All Objects in Hive. Check Data!"
				return response
			}
		}

		response.IsSuccess = true
		response.Message = "Successfully stored object(s) in Hive"

	} else {
		if err == nil {
			err := repository.executeNonQuery(conn, script)
			if err == nil {
				response.IsSuccess = true
				response.Message = "Successfully stored object(s) in Hive"
			} else {
				response.IsSuccess = false
				response.Message = "Error storing data in Hive : " + err.Error()
			}
		} else {
			response.IsSuccess = false
			response.Message = "Error generating Hive query : " + err.Error()
		}
	}
	return response
}

func (repository HiveRepository) getStoreScript(conn *hive.HiveConnection, request *messaging.ObjectRequest) (query string, err error) {
	namespace := request.Controls.Namespace
	class := strings.ToLower(request.Controls.Class)
	var schemaObj map[string]interface{}
	var allObjects []map[string]interface{}
	if request.Body.Object != nil {
		schemaObj = request.Body.Object
		allObjects = make([]map[string]interface{}, 1)
		allObjects[0] = schemaObj
	} else {
		if request.Body.Objects != nil {
			if len(request.Body.Objects) != 0 {
				schemaObj = request.Body.Objects[0]
				allObjects = request.Body.Objects
			} else {
				err = errors.New("No objects available to store")
				return
			}
		} else {
			err = errors.New("No objects available to store")
			return
		}

	}

	repository.verifySchema(conn, request)

	query = ""

	isFirstRow := true
	var keyArray []string

	for _, obj := range allObjects {

		currentObject := repository.getByKey(conn, namespace, class, getNoSqlKeyById(request, obj), request)

		if currentObject == nil {
			if isFirstRow {
				query += ("insert into table " + repository.getDatabaseName(request) + "." + class)
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

			//get fields list
			var fieldOrder []string
			fieldBytes := repository.executeGetFields(request, conn)
			_ = json.Unmarshal(fieldBytes, &fieldOrder)
			fmt.Print("Field Order : ")
			fmt.Println(fieldOrder)

			for index, val := range fieldOrder {
				if val == "osheaders" {
					fieldOrder[index] = "__osheaders"
				}
			}

			if isFirstRow {
				for k, _ := range obj {
					keyList += ("," + k)
					keyArray = append(keyArray, k)
				}
			}
			fmt.Println(keyArray)
			for k := 0; k < len(fieldOrder); k++ {
				for keyvalue, _ := range obj {
					if strings.ToLower(keyvalue) == fieldOrder[k] {
						v := obj[keyvalue]
						valueList += ("," + repository.getSqlFieldValue(v))
					}
				}
			}

			fmt.Println(keyList)
			if isFirstRow {
				query += " VALUES "
			} else {
				query += ","
			}

			query += ("('" + id + "'" + valueList + ")")

		} else {
			updateValues := ""
			isFirst := true
			for k, v := range obj {
				if isFirst {
					isFirst = false
				} else {
					updateValues += ","
				}
				customUpdateKey := ""
				if k == "__osHeaders" {
					customUpdateKey = "osheaders"
				} else {
					customUpdateKey = k
				}
				updateValues += (strings.ToLower(customUpdateKey) + "=" + repository.getSqlFieldValue(v))
			}
			query += ("UPDATE " + repository.getDatabaseName(request) + "." + class + " SET " + updateValues + " WHERE osid='" + getNoSqlKeyById(request, obj) + "'###")
		}

		if isFirstRow {
			isFirstRow = false
		}
	}

	return
}

func (repository HiveRepository) getByKey(conn *hive.HiveConnection, namespace string, class string, id string, request *messaging.ObjectRequest) (obj map[string]interface{}) {
	query := "SELECT * FROM " + repository.getDatabaseName(request) + "." + class + " WHERE osid = '" + id + "'"
	obj, _ = repository.executeQueryOne(conn, query, nil, request)
	return
}

func (repository HiveRepository) getJson(m interface{}) string {
	bytes, _ := json.Marshal(m)
	return string(bytes[:len(bytes)])
}

/*func (repository HiveRepository) getHiveRequest(request *messaging.ObjectRequest) (Request *messaging.ObjectRequest) {
	Request.Controls = request.Controls
	Request.Configuration = request.Configuration
	Request.Extras = request.Extras
	Request.IsLogEnabled = request.IsLogEnabled
	Request.MessageStack = request.MessageStack
	Request.Body.Parameters = request.Body.Parameters
	Request.Body.Parameters.KeyProperty = strings.ToLower(Request.Body.Parameters.KeyProperty)
	Request.Body.Query = request.Body.Query
	Request.Body.Special = request.Body.Special

	var obj map[string]interface{}
	obj = make(map[string]interface{})
	for key, value := range request.Body.Object {
		if key != "__osHeaders" {
			obj[key] = value
		} else {
			obj["osheaders"] = value
		}
	}
	Request.Body.Object = obj

	var objs []map[string]interface{}
	objs = make([]map[string]interface{}, len(request.Body.Objects))

	for index, object := range request.Body.Objects {
		var singleObject map[string]interface{}
		singleObject = make(map[string]interface{})
		for key, value := range object {
			if key != "__osHeaders" {
				singleObject[key] = value
			} else {
				singleObject["osheaders"] = value
			}
		}
		objs[index] = singleObject
	}

	Request.Body.Objects = objs
	return
}*/

func (repository HiveRepository) queryCommonMany(query string, request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryCommon(query, request, false)

}

func (repository HiveRepository) queryCommonOne(query string, request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryCommon(query, request, true)
}

func (repository HiveRepository) queryCommon(query string, request *messaging.ObjectRequest, isOne bool) RepositoryResponse {
	response := RepositoryResponse{}

	conn, _, _ := repository.getConnection(request)
	var err error
	dbName := repository.getDatabaseName(request)

	var obj interface{}
	tableName := dbName + "." + strings.ToLower(request.Controls.Class)
	if isOne {
		obj, err = repository.executeQueryOne(conn, query, tableName, request)
	} else {
		obj, err = repository.executeQueryMany(conn, query, tableName, request)
	}

	if err == nil {
		bytes, _ := json.Marshal(obj)
		if len(bytes) == 4 {
			var empty map[string]interface{}
			empty = make(map[string]interface{})
			response.GetSuccessResByObject(empty)
		} else {
			response.GetResponseWithBody(bytes)
		}
	} else {
		var empty map[string]interface{}
		empty = make(map[string]interface{})
		response.GetSuccessResByObject(empty)
	}
	return response
}

func (repository HiveRepository) executeQueryMany(conn *hive.HiveConnection, query string, tableName interface{}, request *messaging.ObjectRequest) (result []map[string]interface{}, err error) {
	result, err = repository.rowsToMap(conn, tableName, query, request)
	return
}

func (repository HiveRepository) executeQueryOne(conn *hive.HiveConnection, query string, tableName interface{}, request *messaging.ObjectRequest) (result map[string]interface{}, err error) {
	queryResults, err := repository.rowsToMap(conn, tableName, query, request)
	if len(queryResults) > 0 {
		result = queryResults[0]
	}
	return
}

var fieldNamesAndTypes []map[string]interface{}

func (repository HiveRepository) rowsToMap(conn *hive.HiveConnection, tableName interface{}, query string, request *messaging.ObjectRequest) (tableMap []map[string]interface{}, err error) {
	//Get Fields and Types
	fieldNamesAndTypes = nil
	typeBytes := repository.executeGetDataTypes(request, conn)
	_ = json.Unmarshal(typeBytes, &fieldNamesAndTypes)

	fmt.Println("Field Names and Types : ")
	fmt.Println(fieldNamesAndTypes)

	_, err = conn.Client.Execute(query)
	ignoreOsid := strings.ToLower(request.Controls.Class) + ".osid"
	ignoreOsheaders := strings.ToLower(request.Controls.Class) + ".osheaders"
	if err == nil {
		//Get Schema
		schema, _, _ := conn.Client.GetSchema()

		for {
			row, _, _ := conn.Client.FetchOne()
			if row == "" {
				break
			} else {

				var myMap map[string]interface{}
				myMap = make(map[string]interface{})

				temp := strings.Split(row, "\t")

				for key, _ := range temp {
					if (schema.FieldSchemas[key].Name) != ignoreOsid && (schema.FieldSchemas[key].Name) != ignoreOsheaders {
						myMap[(schema.FieldSchemas[key].Name)] = repository.sqlToGolang((schema.FieldSchemas[key].Name), temp[key])
					}
				}
				tableMap = append(tableMap, myMap)
			}
		}
	}

	return
}

func (repository HiveRepository) sqlToGolang(key string, value string) interface{} {

	var object interface{}

	//remove class prefix form key
	keyTokens := strings.Split(key, ".")
	key = keyTokens[(len(keyTokens) - 1)]

	//get datatype for key
	datatype := ""
	for _, nameType := range fieldNamesAndTypes {
		if nameType[key] != nil {
			datatype = nameType[key].(string)
			break
		}
	}

	switch datatype {
	case "string":
		object = value
		break
	case "boolean":
		if value == "true" {
			object = true
		} else {
			object = false
		}
		break
	case "int":
		intValue, err := strconv.Atoi(value)
		if err != nil {
			object = value
		} else {
			object = intValue
		}
		break
	case "double":
		value = strings.TrimSpace(value)
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			fmt.Println(err.Error())
			object = value
		} else {
			object = floatValue
		}
		break
	case "binary":
		var temp interface{}
		err := json.Unmarshal([]byte(value), &temp)
		if err != nil {
			fmt.Println(err.Error())
			object = value
		} else {
			object = temp
		}
		break
	default:
		object = value
		break

	}
	return object

}

//............... OLD CODE - PLEASE DON'T REMOVE-----------------

/*
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
		return repository.GetAll(request)

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
*/
