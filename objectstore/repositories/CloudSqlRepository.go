package repositories

import (
	"database/sql"
	"duov6.com/common"
	"duov6.com/objectstore/messaging"
	"duov6.com/queryparser"
	"duov6.com/term"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
	"time"
)

type CloudSqlRepository struct {
}

func (repository CloudSqlRepository) GetRepositoryName() string {
	return "CloudSQL"
}

func (repository CloudSqlRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Get-All!", term.Debug)
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
	term.Write("Executing Get-Query!", 2)
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

		formattedQuery, err := queryparser.GetCloudSQLQuery(request.Body.Query.Parameters, request.Controls.Namespace, request.Controls.Class, parameters)
		if err != nil {
			term.Write(err.Error(), term.Error)
			response.IsSuccess = false
			response.Message = err.Error()
			return response
		}

		term.Write("Query : "+formattedQuery, term.Debug)
		query := formattedQuery
		response = repository.queryCommonMany(query, request)
	} else {
		response = repository.GetAll(request)
	}
	return response
}

func (repository CloudSqlRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Get-By-Key!", 2)
	query := "SELECT * FROM " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class + " WHERE __os_id = '" + getNoSqlKey(request) + "';"
	//query := "SELECT * FROM " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class + " WHERE __os_id = \"" + getNoSqlKey(request) + "\""
	return repository.queryCommonOne(query, request)
}

func (repository CloudSqlRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Get-Search!", 2)
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

	response := RepositoryResponse{}
	query := ""
	if strings.Contains(request.Body.Query.Parameters, ":") {
		tokens := strings.Split(request.Body.Query.Parameters, ":")
		fieldName := tokens[0]
		fieldValue := tokens[1]

		if len(tokens) > 2 {
			fieldValue = ""
			for x := 1; x < len(tokens); x++ {
				fieldValue += tokens[x] + " "
			}
		}

		fieldName = strings.TrimSpace(fieldName)
		fieldValue = strings.TrimSpace(fieldValue)

		query = "select * from " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class + " where " + fieldName + "='" + fieldValue + "'"
	} else {
		if request.Body.Query.Parameters == "" || request.Body.Query.Parameters == "*" {
			//Get All Query
			query = "select * from " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class
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

	if tableCache[repository.getDatabaseName(request.Controls.Namespace)+"."+request.Controls.Class] != nil {
		//Available in Table Cache
		for name, fieldType := range tableCache[repository.getDatabaseName(request.Controls.Namespace)+"."+request.Controls.Class] {
			if name != "__osHeaders" && strings.EqualFold(fieldType, "TEXT") {
				fieldNames = append(fieldNames, name)
			}
		}
	} else {
		//Get From Db
		query := "SELECT COLUMN_NAME, DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = '" + repository.getDatabaseName(request.Controls.Namespace) + "' AND TABLE_NAME = '" + request.Controls.Class + "';"
		repoResponse := repository.queryCommonMany(query, request)
		var mapArray []map[string]interface{}
		err := json.Unmarshal(repoResponse.Body, &mapArray)
		if err != nil {
			term.Write(err.Error(), term.Error)
		} else {
			for _, value := range mapArray {
				if value["COLUMN_NAME"].(string) != "__osHeaders" && strings.EqualFold(value["DATA_TYPE"].(string), "TEXT") {
					fieldNames = append(fieldNames, value["COLUMN_NAME"].(string))
				}
			}
		}
	}

	/*query = "SELECT * FROM " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class + " WHERE Concat("

	//Make Argument Array
	fullTextArguments := ""
	for _, field := range fieldNames {
		fullTextArguments += "IFNULL(" + field + ",''), '',"
	}

	fullTextArguments = fullTextArguments[:(len(fullTextArguments) - 5)]

	queryParam := request.Body.Query.Parameters
	queryParam = strings.TrimPrefix(queryParam, "*")
	queryParam = strings.TrimSuffix(queryParam, "*")
	query += fullTextArguments + ") LIKE '%" + queryParam + "%' "*/

	//Indexed Queries
	queryParam := request.Body.Query.Parameters
	queryParam = strings.TrimPrefix(queryParam, "*")
	queryParam = strings.TrimSuffix(queryParam, "*")

	query = "SELECT * FROM " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class + " WHERE MATCH ("

	fullTextArguments := ""
	for _, field := range fieldNames {
		fullTextArguments += field + ","
	}
	fullTextArguments = strings.TrimSuffix(fullTextArguments, ",")
	query += fullTextArguments + ") AGAINST ('" + queryParam
	query += "' IN BOOLEAN MODE);"

	return
}

func (repository CloudSqlRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
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

func (repository CloudSqlRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Insert-Single!", 2)
	id := repository.getRecordID(request, request.Body.Object)
	request.Controls.Id = id
	request.Body.Object[request.Body.Parameters.KeyProperty] = id

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

func (repository CloudSqlRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Update-Multiple!", 2)
	return repository.queryStore(request)
}

func (repository CloudSqlRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Update-Single!", 2)
	return repository.queryStore(request)
}

func (repository CloudSqlRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Delete-Multiple!", 2)
	response := RepositoryResponse{}
	conn, err := repository.getConnection(request)
	if err == nil {
		isError := false
		for _, obj := range request.Body.Objects {
			query := repository.getDeleteScript(request.Controls.Namespace, request.Controls.Class, getNoSqlKeyById(request, obj))
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
	term.Write("Executing Delete-Single!", 2)
	response := RepositoryResponse{}
	conn, err := repository.getConnection(request)
	if err == nil {
		query := repository.getDeleteScript(request.Controls.Namespace, request.Controls.Class, getNoSqlKey(request))
		err := repository.executeNonQuery(conn, query)
		if err != nil {
			response.IsSuccess = false
			response.Message = "Failed Deleting from CloudSQL repository : " + err.Error()
		} else {
			response.IsSuccess = true
			response.Message = "Successfully Deleted from CloudSQL repository!"
		}
	} else {
		response.IsSuccess = false
		response.Message = "Failed Deleting from CloudSQL repository : " + err.Error()
	}
	repository.closeConnection(conn)
	return response
}

func (repository CloudSqlRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Special!", 2)
	response := RepositoryResponse{}
	queryType := request.Body.Special.Type

	switch queryType {
	case "getFields":
		request.Log("Starting GET-FIELDS sub routine!")
		query := "SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = '" + repository.getDatabaseName(request.Controls.Namespace) + "' AND TABLE_NAME = '" + request.Controls.Class + "';"
		repoResponse := repository.queryCommonMany(query, request)
		var mapArray []map[string]interface{}
		err := json.Unmarshal(repoResponse.Body, &mapArray)
		if err != nil {
			term.Write(err.Error(), 1)
			repoResponse.Body = nil
			return repoResponse
		} else {
			valueArray := make([]string, len(mapArray))
			for index, value := range mapArray {
				valueArray[index] = value["COLUMN_NAME"].(string)
			}
			repoResponse.Body, _ = json.Marshal(valueArray)
			return repoResponse
		}
	case "getClasses":
		request.Log("Starting GET-CLASSES sub routine")
		query := "SELECT DISTINCT TABLE_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='" + repository.getDatabaseName(request.Controls.Namespace) + "';"
		repoResponse := repository.queryCommonMany(query, request)
		var mapArray []map[string]interface{}
		err := json.Unmarshal(repoResponse.Body, &mapArray)
		if err != nil {
			term.Write(err.Error(), 1)
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
	case "getNamespaces":
		request.Log("Starting GET-NAMESPACES sub routine")
		query := "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME != 'information_schema' AND SCHEMA_NAME !='mysql' AND SCHEMA_NAME !='performance_schema';"
		return repository.queryCommonMany(query, request)
	case "getSelected":
		fieldNames := strings.Split(strings.TrimSpace(request.Body.Special.Parameters), " ")
		query := "select " + fieldNames[0]
		for x := 1; x < len(fieldNames); x++ {
			query += "," + fieldNames[x]
		}
		query += " from " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class
		return repository.queryCommonMany(query, request)
	case "DropClass":
		request.Log("Starting Delete-Class sub routine")
		conn, err := repository.getConnection(request)
		if err == nil {
			query := "DROP TABLE " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class
			err := repository.executeNonQuery(conn, query)
			if err != nil {
				response.IsSuccess = false
				response.Message = "Error Dropping Table in CloudSQL Repository : " + err.Error()
			} else {
				//Delete Class from availableTables and tablecache
				delete(availableTables, (repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class))
				delete(tableCache, (repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class))
				response.IsSuccess = true
				response.Message = "Successfully Dropped Table : " + request.Controls.Class
			}
		} else {
			response.IsSuccess = false
			response.Message = "Connection Failed to CloudSQL Server"
		}
		repository.closeConnection(conn)
	case "DropNamespace":
		request.Log("Starting Delete-Database sub routine")
		conn, err := repository.getConnection(request)
		if err == nil {
			query := "DROP SCHEMA " + repository.getDatabaseName(request.Controls.Namespace)
			err := repository.executeNonQuery(conn, query)
			if err != nil {
				response.IsSuccess = false
				response.Message = "Error Dropping Table in CloudSQL Repository : " + err.Error()
			} else {
				//Delete Namespace from availableDbs
				delete(availableDbs, repository.getDatabaseName(request.Controls.Namespace))
				//Delete all associated Classes from it's TableCache and availableTables
				for key, _ := range availableTables {
					if strings.Contains(key, repository.getDatabaseName(request.Controls.Namespace)) {
						delete(availableTables, key)
						delete(tableCache, key)
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
	default:
		return repository.GetAll(request)

	}

	return response
}

func (repository CloudSqlRepository) Test(request *messaging.ObjectRequest) {

}

////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////SQL GENERATORS/////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////
func (repository CloudSqlRepository) queryCommon(query string, request *messaging.ObjectRequest, isOne bool) RepositoryResponse {
	response := RepositoryResponse{}

	conn, err := repository.getConnection(request)
	if err == nil {
		var err error
		dbName := repository.getDatabaseName(request.Controls.Namespace)
		err = repository.buildTableCache(conn, dbName, request.Controls.Class)
		if err != nil {

		}

		var obj interface{}
		tableName := dbName + "." + request.Controls.Class
		if isOne {
			obj, err = repository.executeQueryOne(conn, query, tableName)
		} else {
			obj, err = repository.executeQueryMany(conn, query, tableName)
		}

		if err == nil {
			var bytes []byte
			if isOne {
				bytes, _ = json.Marshal(obj.(map[string]interface{}))
			} else {
				bytes, _ = json.Marshal(obj.([]map[string]interface{}))
			}

			term.Write("--------- Object Value ----------", term.Debug)
			if len(bytes) > 1000 {
				term.Write("Data Found but Too Long to STDOUT!", term.Debug)
			} else {
				term.Write(obj, term.Debug)
			}
			term.Write("---------------------------------", term.Debug)

			//bytes, _ := json.Marshal(obj)
			if checkEmptyByteArray(bytes) {
				response.GetResponseWithBody(getEmptyByteObject())
			} else {
				response.GetResponseWithBody(bytes)
				//response.GetSuccessResByObject(obj)
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

var updateQueryCloudSql []string

func (repository CloudSqlRepository) queryStore(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	conn, _ := repository.getConnection(request)

	isOkay := true

	//execute insert queries
	scripts, err := repository.getStoreScript(conn, request)
	for x := 0; x < len(scripts); x++ {
		script := scripts[x]
		if err == nil {
			if script != "" {
				err := repository.executeNonQuery(conn, script)
				if err != nil {
					isOkay = false
				}
			}
		} else {
			isOkay = false
		}
	}

	//execute update queries
	if updateQueryCloudSql != nil && len(updateQueryCloudSql) > 0 {
		for x := 0; x < len(updateQueryCloudSql); x++ {
			updateQuery := updateQueryCloudSql[x]
			err := repository.executeNonQuery(conn, updateQuery)
			if err != nil {
				term.Write("Error! "+err.Error(), term.Error)
				isOkay = false
			}
		}
	}
	//clear update array
	updateQueryCloudSql = updateQueryCloudSql[:0]

	if isOkay {
		response.IsSuccess = true
		response.Message = "Successfully stored object(s) in CloudSQL"
		request.Log(response.Message)
	} else {
		response.IsSuccess = false
		response.Message = "Error storing/updaing all object(s) in CloudSQL."
		request.Log(response.Message)
	}

	repository.closeConnection(conn)
	return response
}

// func (repository CloudSqlRepository) queryStore(request *messaging.ObjectRequest) RepositoryResponse {
// 	response := RepositoryResponse{}

// 	conn, _ := repository.getConnection(request)

// 	scripts, err := repository.getStoreScript(conn, request)
// 	for x := 0; x < len(scripts); x++ {
// 		script := scripts[x]
// 		queryArray := strings.Split(script, "###")

// 		if strings.Contains(script, "###") {
// 			//Multiple Updates
// 			for index := 0; index < (len(queryArray) - 1); index++ {
// 				err := repository.executeNonQuery(conn, queryArray[index])
// 				if err != nil {
// 					response.IsSuccess = false
// 					response.Message = "Error Inserting All Objects in CloudSQL. Check Data!"
// 					return response
// 				}
// 			}

// 		} else {
// 			if err == nil {
// 				err := repository.executeNonQuery(conn, script)
// 				if err != nil {
// 					response.IsSuccess = false
// 					response.Message = "Error Inserting All Objects in CloudSQL. Check Data!"
// 					return response
// 				}
// 			} else {
// 				response.IsSuccess = false
// 				response.Message = "Error generating CloudSQL query : " + err.Error()
// 				return response
// 			}
// 		}
// 	}

// 	response.IsSuccess = true
// 	response.Message = "Successfully stored object(s) in CloudSQL"

// 	repository.closeConnection(conn)
// 	return response
// }

func (repository CloudSqlRepository) getByKey(conn *sql.DB, namespace string, class string, id string) (obj map[string]interface{}) {
	query := "SELECT * FROM " + repository.getDatabaseName(namespace) + "." + class + " WHERE __os_id = '" + id + "';"
	//query := "SELECT * FROM " + repository.getDatabaseName(namespace) + "." + class + " WHERE __os_id = \"" + id + "\""
	obj, _ = repository.executeQueryOne(conn, query, nil)
	term.Write("------------  GetByKey Value ---------------", term.Debug)
	bytes, _ := json.Marshal(obj)
	if len(bytes) > 1000 {
		term.Write("Data Found but Too Long to STDOUT!", term.Debug)
	} else {
		term.Write(obj, term.Debug)
	}
	term.Write("--------------------------------------------", term.Debug)
	return
}

func (repository CloudSqlRepository) getStoreScript(conn *sql.DB, request *messaging.ObjectRequest) (query []string, err error) {
	namespace := request.Controls.Namespace
	class := request.Controls.Class
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

	repository.checkSchema(conn, namespace, class, schemaObj)

	if request.Body.Object != nil {
		arr := make([]map[string]interface{}, 1)
		arr[0] = request.Body.Object
		queryOutput := repository.getSingleQuery(request, namespace, class, arr, conn)
		query = append(query, queryOutput)
	} else {

		noOfElementsPerSet := 1000
		noOfSets := (len(request.Body.Objects) / noOfElementsPerSet)
		remainderFromSets := 0
		remainderFromSets = (len(request.Body.Objects) - (noOfSets * noOfElementsPerSet))

		startIndex := 0
		stopIndex := noOfElementsPerSet

		for x := 0; x < noOfSets; x++ {
			queryOutput := repository.getSingleQuery(request, namespace, class, request.Body.Objects[startIndex:stopIndex], conn)
			query = append(query, queryOutput)
			startIndex += noOfElementsPerSet
			stopIndex += noOfElementsPerSet
		}

		if remainderFromSets > 0 {
			start := len(request.Body.Objects) - remainderFromSets
			queryOutput := repository.getSingleQuery(request, namespace, class, request.Body.Objects[start:len(request.Body.Objects)], conn)
			query = append(query, queryOutput)
		}

	}
	return
}

func (repository CloudSqlRepository) getSingleQueryNew(request *messaging.ObjectRequest, namespace, class string, records []map[string]interface{}, conn *sql.DB) (query string) {
	var updateArray []map[string]interface{}
	var insertArray []map[string]interface{}

	for _, obj := range records {
		currentObject := repository.getByKey(conn, namespace, class, getNoSqlKeyById(request, obj))
		if currentObject == nil || len(currentObject) == 0 {
			insertArray = append(insertArray, obj)
		} else {
			updateArray = append(updateArray, obj)
		}
	}

	//create update scripts
	for _, obj := range updateArray {
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
		Updatequery := ("UPDATE " + repository.getDatabaseName(namespace) + "." + class + " SET " + updateValues + " WHERE __os_id=\"" + getNoSqlKeyById(request, obj) + "\";")
		updateQueryCloudSql = append(updateQueryCloudSql, Updatequery)
	}

	//create insert scripts
	isFirstRow := true
	var keyArray []string
	for _, obj := range insertArray {
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
		//term.Write(keyArray)
		for _, k := range keyArray {
			v := obj[k]
			valueList += ("," + repository.getSqlFieldValue(v))
		}

		if isFirstRow {
			query += "(__os_id" + keyList + ") VALUES "
		} else {
			query += ","
		}

		//query += ("(\"" + getNoSqlKeyById(request, obj) + "\"" + valueList + ")")
		query += ("(\"" + id + "\"" + valueList + ")")

		if isFirstRow {
			isFirstRow = false
		}
	}

	return
}

func (repository CloudSqlRepository) getSingleQuery(request *messaging.ObjectRequest, namespace, class string, records []map[string]interface{}, conn *sql.DB) (query string) {

	return repository.getSingleQueryNew(request, namespace, class, records, conn)

	// runtime := 0       //times of loop run
	// lastOperation := 1 //0=update, 1= insert
	// isFirstRow := true
	// var keyArray []string
	// for _, obj := range records {
	// 	currentObject := repository.getByKey(conn, namespace, class, getNoSqlKeyById(request, obj))

	// 	if currentObject == nil || len(currentObject) == 0 {
	// 		lastOperation = 1
	// 		runtime += 1
	// 		if isFirstRow {
	// 			query += ("INSERT INTO " + repository.getDatabaseName(namespace) + "." + class)
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
	// 		//term.Write(keyArray)
	// 		for _, k := range keyArray {
	// 			v := obj[k]
	// 			valueList += ("," + repository.getSqlFieldValue(v))
	// 		}

	// 		if isFirstRow {
	// 			query += "(__os_id" + keyList + ") VALUES "
	// 		} else {
	// 			query += ","
	// 		}

	// 		//query += ("(\"" + getNoSqlKeyById(request, obj) + "\"" + valueList + ")")
	// 		query += ("(\"" + id + "\"" + valueList + ")")

	// 	} else {
	// 		runtime += 1
	// 		lastOperation = 0
	// 		updateValues := ""
	// 		isFirst := true
	// 		for k, v := range obj {
	// 			if isFirst {
	// 				isFirst = false
	// 			} else {
	// 				updateValues += ","
	// 			}

	// 			updateValues += (k + "=" + repository.getSqlFieldValue(v))
	// 		}
	// 		Updatequery := ("UPDATE " + repository.getDatabaseName(namespace) + "." + class + " SET " + updateValues + " WHERE __os_id=\"" + getNoSqlKeyById(request, obj) + "\";")
	// 		updateQueryCloudSql = append(updateQueryCloudSql, Updatequery)
	// 		//query += ("UPDATE " + repository.getDatabaseName(namespace) + "." + class + " SET " + updateValues + " WHERE __os_id=\"" + getNoSqlKeyById(request, obj) + "\";###")
	// 		//_ = repository.executeNonQuery(conn, Updatequery)
	// 	}

	// 	if isFirstRow {
	// 		isFirstRow = false
	// 	}

	// 	if runtime == 1 && lastOperation == 0 {
	// 		isFirstRow = true
	// 	} else {
	// 		isFirstRow = false
	// 	}

	// 	// if isFirstRow {
	// 	// 	isFirstRow = false
	// 	// }
	// }
	// return
}

func (repository CloudSqlRepository) getDeleteScript(namespace string, class string, id string) string {
	return "DELETE FROM " + repository.getDatabaseName(namespace) + "." + class + " WHERE __os_id = \"" + id + "\""
}

func (repository CloudSqlRepository) getCreateScript(namespace string, class string, obj map[string]interface{}) string {
	query := "CREATE TABLE IF NOT EXISTS " + repository.getDatabaseName(namespace) + "." + class + "(__os_id varchar(255) primary key"
	//query := "CREATE TABLE IF NOT EXISTS " + repository.getDatabaseName(namespace) + "." + class + "(__os_id text"

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

	if len(textFields) > 0 {
		query += ", FULLTEXT("
		fieldList := ""

		for _, field := range textFields {
			fieldList += field + ","
		}

		fieldList = strings.TrimSuffix(fieldList, ",")
		query += fieldList
		query += ")"
	}

	query += ")"

	//term.Write(query)
	return query
}

var availableDbs map[string]interface{}
var availableTables map[string]interface{}
var tableCache map[string]map[string]string

func (repository CloudSqlRepository) checkAvailabilityDb(conn *sql.DB, dbName string) (err error) {
	if availableDbs == nil {
		availableDbs = make(map[string]interface{})
	}

	if availableDbs[dbName] != nil {
		return
	}

	dbQuery := "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '" + dbName + "'"
	dbResult, err := repository.executeQueryOne(conn, dbQuery, nil)

	if err == nil {
		if dbResult["SCHEMA_NAME"] == nil {
			repository.executeNonQuery(conn, "CREATE DATABASE IF NOT EXISTS "+dbName)
		}
		if availableDbs[dbName] == nil {
			availableDbs[dbName] = true
		}
	} else {
		term.Write(err.Error(), 1)
	}

	return
}

func (repository CloudSqlRepository) checkAvailabilityTable(conn *sql.DB, dbName string, namespace string, class string, obj map[string]interface{}) (err error) {

	if availableTables == nil {
		availableTables = make(map[string]interface{})
	}

	if availableTables[dbName+"."+class] == nil {
		var tableResult map[string]interface{}
		tableResult, err = repository.executeQueryOne(conn, "SHOW TABLES FROM "+dbName+" LIKE \""+class+"\"", nil)
		if err == nil {
			if tableResult["Tables_in_"+dbName] == nil {
				script := repository.getCreateScript(namespace, class, obj)
				err = repository.executeNonQuery(conn, script)

				if err != nil {
					return
				}
			}
			if availableTables[dbName+"."+class] == nil || availableTables[dbName+"."+class] == false {
				availableTables[dbName+"."+class] = true
			}

		} else {
			return
		}
	}

	err = repository.buildTableCache(conn, dbName, class)

	alterColumns := ""
	cacheItem := tableCache[dbName+"."+class]

	term.Write("------------ CURRENT TABLE FIELDS AND TYPES -----------------", term.Debug)
	for singleFieldName, singleFieldType := range cacheItem {
		term.Write(singleFieldName+" : "+singleFieldType, term.Debug)
	}
	term.Write("-------------------------------------------------------------", term.Debug)

	isFirst := true
	for k, v := range obj {
		if k != "OriginalIndex" {
			_, ok := cacheItem[k]
			if !ok {
				if isFirst {
					isFirst = false
				} else {
					alterColumns += ", "
				}

				alterColumns += ("ADD COLUMN " + k + " " + repository.golangToSql(v))
				repository.addColumnToTableCache(dbName, class, k, repository.golangToSql(v))
			}
		}
	}

	if len(alterColumns) != 0 {
		alterQuery := "ALTER TABLE " + dbName + "." + class + " " + alterColumns
		err = repository.executeNonQuery(conn, alterQuery)
		//update Fulltext fields
		fullTextQuery := "ALTER TABLE " + dbName + "." + class + " ADD FULLTEXT("
		tableTypes := tableCache[dbName+"."+class]

		fullTextFields := ""

		for field, fieldtype := range tableTypes {
			if strings.EqualFold(fieldtype, "TEXT") {
				fullTextFields += field + ","
			}
		}

		fullTextFields = strings.TrimSuffix(fullTextFields, ",")
		fullTextQuery += fullTextFields
		fullTextQuery += ");"
		err = repository.executeNonQuery(conn, fullTextQuery)

	}

	return
}

func (repository CloudSqlRepository) addColumnToTableCache(dbName string, class string, field string, datatype string) {
	dataMap := make(map[string]string)
	dataMap = tableCache[dbName+"."+class]
	dataMap[field] = datatype
	tableCache[dbName+"."+class] = dataMap
}

func (repository CloudSqlRepository) buildTableCache(conn *sql.DB, dbName string, class string) (err error) {
	if tableCache == nil {
		tableCache = make(map[string]map[string]string)
	}

	_, ok := tableCache[dbName+"."+class]

	if !ok {
		var exResult []map[string]interface{}
		exResult, err = repository.executeQueryMany(conn, "EXPLAIN "+dbName+"."+class, nil)
		if err == nil {
			newMap := make(map[string]string)

			for _, cRow := range exResult {
				newMap[cRow["Field"].(string)] = cRow["Type"].(string)
			}
			if tableCache[dbName+"."+class] == nil {
				tableCache[dbName+"."+class] = newMap
			}
		}
	}

	return
}

func (repository CloudSqlRepository) checkSchema(conn *sql.DB, namespace string, class string, obj map[string]interface{}) {
	dbName := repository.getDatabaseName(namespace)
	err := repository.checkAvailabilityDb(conn, dbName)

	if err == nil {
		err := repository.checkAvailabilityTable(conn, dbName, namespace, class, obj)

		if err != nil {
			term.Write(err.Error(), 1)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////Helper functions/////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////

var connection map[string]*sql.DB

func (repository CloudSqlRepository) getConnection(request *messaging.ObjectRequest) (conn *sql.DB, err error) {

	if connection == nil {
		connection = make(map[string]*sql.DB)
	}

	//var stats sql.DBStats
	//_ = time.Now()
	//term.Write("-----------------------------------------------", term.Information)
	//connParams := request.Configuration.ServerConfiguration["MYSQL"]
	//term.Write("Using Server : "+connParams["Url"], term.Information)
	//term.Write("Connection : ", term.Information)

	if connection[request.Controls.Namespace] == nil {
		//	term.Write("Nil Connection! Creating MySQL Connection!", term.Information)
		var c *sql.DB
		mysqlConf := request.Configuration.ServerConfiguration["MYSQL"]
		c, err = sql.Open("mysql", mysqlConf["Username"]+":"+mysqlConf["Password"]+"@tcp("+mysqlConf["Url"]+":"+mysqlConf["Port"]+")/")
		c.SetMaxIdleConns(10)
		c.SetMaxOpenConns(0)
		c.SetConnMaxLifetime(time.Duration(120) * time.Second)
		conn = c
		connection[request.Controls.Namespace] = c
		//stats = conn.Stats()
	} else {
		if connection[request.Controls.Namespace].Ping(); err != nil {
			_ = connection[request.Controls.Namespace].Close()
			connection[request.Controls.Namespace] = nil
			//	term.Write("Cached Connection Timed Out! Creating new Connection!", term.Information)
			var c *sql.DB
			mysqlConf := request.Configuration.ServerConfiguration["MYSQL"]
			c, err = sql.Open("mysql", mysqlConf["Username"]+":"+mysqlConf["Password"]+"@tcp("+mysqlConf["Url"]+":"+mysqlConf["Port"]+")/")
			c.SetMaxIdleConns(10)
			c.SetMaxOpenConns(0)
			c.SetConnMaxLifetime(time.Duration(120) * time.Second)
			conn = c
			connection[request.Controls.Namespace] = c
			//stats = conn.Stats()
		} else {
			//term.Write("Using Cached Connection!", term.Information)
			conn = connection[request.Controls.Namespace]
			//stats = conn.Stats()
		}
	}

	//fmt.Println("Open Connections : " + strconv.Itoa(stats.OpenConnections))
	return conn, err
}

/*
func (repository CloudSqlRepository) getConnection(request *messaging.ObjectRequest) (conn *sql.DB, err error) {

	// if connection == nil {
	// 	connection = make(map[string]*sql.DB)
	// }

	// term.Write()
	// term.Write("-------------------------------")
	// connParams := request.Configuration.ServerConfiguration["MYSQL"]
	// term.Write("Using Server : " + connParams["Url"])
	// term.Write("Connection : ")

	// if connection[request.Controls.Namespace] == nil {
	term.Write("Nil Connection! Creating MySQL Connection!", term.Debug)
	var c *sql.DB
	mysqlConf := request.Configuration.ServerConfiguration["MYSQL"]
	c, err = sql.Open("mysql", mysqlConf["Username"]+":"+mysqlConf["Password"]+"@tcp("+mysqlConf["Url"]+":"+mysqlConf["Port"]+")/")
	//c.SetMaxIdleConns(100)
	//c.SetMaxOpenConns(0)
	//c.SetConnMaxLifetime(time.Duration(600) * time.Second)
	_ = time.Now()
	conn = c
	//connection[request.Controls.Namespace] = c
	// } else {
	// 	if connection[request.Controls.Namespace].Ping(); err != nil {
	// 		term.Write("Cached Connection Timed Out! Creating new Connection!")
	// 		var c *sql.DB
	// 		mysqlConf := request.Configuration.ServerConfiguration["MYSQL"]
	// 		c, err = sql.Open("mysql", mysqlConf["Username"]+":"+mysqlConf["Password"]+"@tcp("+mysqlConf["Url"]+":"+mysqlConf["Port"]+")/")
	// 		//c.SetMaxIdleConns(100)
	// 		//c.SetMaxOpenConns(0)
	// 		//c.SetConnMaxLifetime(time.Duration(600) * time.Second)
	// 		conn = c
	// 		connection[request.Controls.Namespace] = c
	// 	} else {
	// 		term.Write("Using Cached Connection!")
	// 		conn = connection[request.Controls.Namespace]
	// 	}
	// }

	// if conn == nil {
	// 	term.Write("NIL CONNECTION! SOMETHING TERRIBLY GONE WRONG!")
	// } else {
	// 	term.Write(conn)
	// }

	// term.Write("-------------------------------")
	// term.Write()
	if err != nil {
		term.Write("Connection Error! Check DB : "+err.Error(), term.Debug)
	} else {
		term.Write("Connection Created Successfully!", term.Debug)
	}
	return conn, err
}
*/

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

func (repository CloudSqlRepository) golangToSql(value interface{}) string {

	var strValue string

	//term.Write(reflect.TypeOf(value))
	switch value.(type) {
	case string:
		strValue = "TEXT"
	case bool:
		strValue = "BIT"
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
		strValue = "INT (10)"
		break
	case float32:
	case float64:
		strValue = "DOUBLE"
		break
	default:
		strValue = "LONGBLOB"
		break

	}

	return strValue
}

// func (repository CloudSqlRepository) sqlToGolang(b []byte, t string) interface{} {

// 	if b == nil {
// 		return nil
// 	}

// 	if len(b) == 0 {
// 		return b
// 	}

// 	var outData interface{}
// 	tmp := string(b)
// 	switch t {
// 	case "bit(1)":
// 		if len(b) == 0 {
// 			outData = false
// 		} else {
// 			if b[0] == 1 {
// 				outData = true
// 			} else {
// 				outData = false
// 			}
// 		}

// 		break
// 	case "double":
// 		fData, err := strconv.ParseFloat(tmp, 64)
// 		if err != nil {
// 			term.Write(err.Error(), 1)
// 			outData = tmp
// 		} else {
// 			outData = fData
// 		}
// 		break
// 	//case "text":
// 	//case "blob":
// 	default:
// 		if len(tmp) == 4 {
// 			if strings.ToLower(tmp) == "null" {
// 				outData = nil
// 				break
// 			}
// 		}

// 		// var m map[string]interface{}
// 		// var ml []map[string]interface{}

// 		// if (string(tmp[0]) == "{"){
// 		// 	err := json.Unmarshal([]byte(tmp), &m)
// 		// 	if err == nil {
// 		// 		outData = m
// 		// 	}else{
// 		// 		term.Write(err.Error())
// 		// 		outData = tmp
// 		// 	}
// 		// }else if (string(tmp[0]) == "["){
// 		// 	err := json.Unmarshal([]byte(tmp), &ml)
// 		// 	if err == nil {
// 		// 		outData = ml
// 		// 	}else{
// 		// 		term.Write(err.Error())
// 		// 		outData = tmp
// 		// 	}
// 		// }else{
// 		// 	outData = tmp
// 		// }

// 		if string(tmp[0]) == "^" {
// 			byteData := []byte(tmp)
// 			bdata := string(byteData[1:])
// 			decData, _ := base64.StdEncoding.DecodeString(bdata)
// 			outData = repository.getInterfaceValue(string(decData))

// 		} else {
// 			outData = repository.getInterfaceValue(tmp)
// 		}

// 		break
// 	}

// 	return outData
// }

func (repository CloudSqlRepository) sqlToGolang(b []byte, t string) interface{} {

	if b == nil {
		return nil
	}

	if len(b) == 0 {
		return b
	}

	// term.Write("--------------")
	// term.Write(t)
	// term.Write(string(b))
	// term.Write("--------------")

	var outData interface{}
	tmp := string(b)
	switch t {
	case "bit(1)":
		if len(b) == 0 {
			outData = false
		} else {
			if b[0] == 1 {
				outData = true
			} else {
				outData = false
			}
		}

		break
	case "double":
		fData, err := strconv.ParseFloat(tmp, 64)
		if err != nil {
			term.Write(err.Error(), 1)
			outData = tmp
		} else {
			outData = fData
		}
		break
	case "BIT":
		if len(b) == 0 {
			outData = false
		} else {
			if b[0] == 1 {
				outData = true
			} else {
				outData = false
			}
		}

		break
	case "DOUBLE":
		fData, err := strconv.ParseFloat(tmp, 64)
		if err != nil {
			term.Write(err.Error(), 1)
			outData = tmp
		} else {
			outData = fData
		}
		break
	//case "text":
	//case "blob":
	default:
		if len(tmp) == 4 {
			if strings.ToLower(tmp) == "null" {
				outData = nil
				break
			}
		}

		// var m map[string]interface{}
		// var ml []map[string]interface{}

		// if (string(tmp[0]) == "{"){
		// 	err := json.Unmarshal([]byte(tmp), &m)
		// 	if err == nil {
		// 		outData = m
		// 	}else{
		// 		term.Write(err.Error())
		// 		outData = tmp
		// 	}
		// }else if (string(tmp[0]) == "["){
		// 	err := json.Unmarshal([]byte(tmp), &ml)
		// 	if err == nil {
		// 		outData = ml
		// 	}else{
		// 		term.Write(err.Error())
		// 		outData = tmp
		// 	}
		// }else{
		// 	outData = tmp
		// }

		if string(tmp[0]) == "^" {
			byteData := []byte(tmp)
			bdata := string(byteData[1:])
			decData, _ := base64.StdEncoding.DecodeString(bdata)
			outData = repository.getInterfaceValue(string(decData))

		} else {
			outData = repository.getInterfaceValue(tmp)
		}

		break
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
			term.Write(err.Error(), 1)
			outData = tmp
		}
	} else {
		outData = tmp
	}
	return
}

func (repository CloudSqlRepository) rowsToMap(rows *sql.Rows, tableName interface{}) (tableMap []map[string]interface{}, err error) {

	columns, _ := rows.Columns()
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

	var cacheItem map[string]string

	if tableName != nil {
		tName := tableName.(string)
		cacheItem = tableCache[tName]
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

func (repository CloudSqlRepository) executeQueryMany(conn *sql.DB, query string, tableName interface{}) (result []map[string]interface{}, err error) {
	rows, err := conn.Query(query)
	term.Write("Query Many : ", term.Debug)

	if len(query) > 1000 {
		term.Write("Query Found but Too Long to STDOUT!", term.Debug)
	} else {
		term.Write(query, term.Debug)
	}

	if err == nil {
		result, err = repository.rowsToMap(rows, tableName)
	} else {
		if strings.HasPrefix(err.Error(), "Error 1146") {
			err = nil
			result = make([]map[string]interface{}, 0)
		}
	}

	return
}

func (repository CloudSqlRepository) executeQueryOne(conn *sql.DB, query string, tableName interface{}) (result map[string]interface{}, err error) {
	rows, err := conn.Query(query)
	term.Write("Query One : ", term.Debug)
	if len(query) > 1000 {
		term.Write("Query Found but Too Long to STDOUT!", term.Debug)
	} else {
		term.Write(query, term.Debug)
	}

	if err == nil {
		var resultSet []map[string]interface{}
		resultSet, err = repository.rowsToMap(rows, tableName)
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

func (repository CloudSqlRepository) executeNonQuery(conn *sql.DB, query string) (err error) {

	term.Write("Executing Non-Query : ", term.Debug)
	if len(query) > 1000 {
		term.Write("Query Found but Too Long to STDOUT!", term.Debug)
	} else {
		term.Write(query, term.Debug)
	}

	var stmt *sql.Stmt
	stmt, err = conn.Prepare(query)
	if err != nil {
		term.Write(err.Error(), term.Debug)
		return err
	}
	_, err = stmt.Exec()

	if err != nil {
		term.Write(err.Error(), term.Debug)
		return err
	}

	_ = stmt.Close()
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
		session, isError := repository.getConnection(request)
		if isError != nil {
			returnID = ""
			repository.closeConnection(session)
			return
		} else {

			_ = repository.checkAvailabilityDb(session, repository.getDatabaseName(request.Controls.Namespace))

			//Reading maxCount from DB
			checkTableQuery := "SELECT DISTINCT TABLE_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='" + repository.getDatabaseName(request.Controls.Namespace) + "' AND TABLE_NAME='domainClassAttributes';"
			tableResultMap, _ := repository.executeQueryOne(session, checkTableQuery, request.Controls.Class)
			if len(tableResultMap) == 0 {
				//Create new domainClassAttributes  table
				createDomainAttrQuery := "create table " + repository.getDatabaseName(request.Controls.Namespace) + ".domainClassAttributes ( class VARCHAR(255) primary key, maxCount text, version text);"
				err := repository.executeNonQuery(session, createDomainAttrQuery)
				if err != nil {
					returnID = "1"
					repository.closeConnection(session)
					return
				} else {
					//insert record with count 1 and return
					insertQuery := "INSERT INTO " + repository.getDatabaseName(request.Controls.Namespace) + ".domainClassAttributes (class, maxCount,version) VALUES ('" + strings.ToLower(request.Controls.Class) + "','1','" + common.GetGUID() + "')"
					err = repository.executeNonQuery(session, insertQuery)
					if err != nil {
						returnID = "1"
						repository.closeConnection(session)
						return
					} else {
						returnID = "1"
						repository.closeConnection(session)
						return
					}
				}
			} else {
				//This is a new Class.. Create New entry
				readQuery := "SELECT maxCount FROM " + repository.getDatabaseName(request.Controls.Namespace) + ".domainClassAttributes where class = '" + strings.ToLower(request.Controls.Class) + "';"
				myMap, _ := repository.executeQueryOne(session, readQuery, (repository.getDatabaseName(request.Controls.Namespace) + ".domainClassAttributes"))

				if len(myMap) == 0 {
					request.Log("New Class! New record for this class will be inserted")
					insertNewClassQuery := "INSERT INTO " + repository.getDatabaseName(request.Controls.Namespace) + ".domainClassAttributes (class,maxCount,version) values ('" + strings.ToLower(request.Controls.Class) + "', '1', '" + common.GetGUID() + "');"
					err := repository.executeNonQuery(session, insertNewClassQuery)
					if err != nil {
						returnID = ""
						repository.closeConnection(session)
						return
					} else {
						returnID = "1"
						repository.closeConnection(session)
						return
					}
				} else {
					//Inrement one and UPDATE
					maxCount := 0
					maxCount, err := strconv.Atoi(myMap["maxCount"].(string))
					maxCount++
					returnID = strconv.Itoa(maxCount)
					updateQuery := "UPDATE " + repository.getDatabaseName(request.Controls.Namespace) + ".domainClassAttributes SET maxCount='" + returnID + "' WHERE class = '" + strings.ToLower(request.Controls.Class) + "' ;"
					err = repository.executeNonQuery(session, updateQuery)
					if err != nil {
						returnID = ""
						repository.closeConnection(session)
						return
					}
				}
			}
		}
		repository.closeConnection(session)
	} else {
		returnID = obj[request.Body.Parameters.KeyProperty].(string)
	}
	return
}

func (repository CloudSqlRepository) closeConnection(conn *sql.DB) {
	// err := conn.Close()
	// if err != nil {
	// 	term.Write(err.Error(), 1)
	// } else {
	// 	term.Write("Connection Closed!", 2)
	// }
}
