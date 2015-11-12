package repositories

import (
	"database/sql"
	"duov6.com/objectstore/connmanager"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/queryparser"
	"duov6.com/term"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/twinj/uuid"
	"strconv"
	"strings"
)

type CloudSqlRepository struct {
}

func (repository CloudSqlRepository) GetRepositoryName() string {
	return "CloudSQL"
}

func (repository CloudSqlRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Get-All!", 2)
	isSkippable := false
	isTakable := false
	skip := "0"
	take := "100000"

	if request.Extras["skip"] != nil {
		skip = request.Extras["skip"].(string)
	}

	if request.Extras["take"] != nil {
		take = request.Extras["take"].(string)
	}

	query := "SELECT * FROM " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class
	if isTakable {
		query += " limit " + take
	}
	if isSkippable {
		query += " offset " + skip
	}

	return repository.queryCommonMany(query, request)
}

func (repository CloudSqlRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Get-Query!", 2)
	response := RepositoryResponse{}
	if request.Body.Query.Parameters != "*" {
		formattedQuery := queryparser.GetFormattedQuerywithDB(request.Body.Query.Parameters, repository.getDatabaseName(request.Controls.Namespace))
		term.Write(("Formatted Query : " + formattedQuery), 2)
		query := formattedQuery
		response = repository.queryCommonMany(query, request)
	} else {
		response = repository.GetAll(request)
	}
	return response
}

func (repository CloudSqlRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Get-By-Key!", 2)
	query := "SELECT * FROM " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class + " WHERE __os_id = \"" + getNoSqlKey(request) + "\""
	return repository.queryCommonOne(query, request)
}

func (repository CloudSqlRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Get-Search!", 2)
	response := RepositoryResponse{}
	query := ""
	if strings.Contains(request.Body.Query.Parameters, ":") {
		tokens := strings.Split(request.Body.Query.Parameters, ":")
		fieldName := tokens[0]
		fieldValue := tokens[1]
		query = "select * from " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class + " where " + fieldName + "='" + fieldValue + "';"
	} else {
		query = "select * from " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class + ";"
	}
	request.Log(query)
	response = repository.queryCommonMany(query, request)
	return response
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
	return response
}

func (repository CloudSqlRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Special!", 2)
	response := RepositoryResponse{}
	queryType := request.Body.Special.Type

	switch queryType {
	case "getFields":
		request.Log("Starting GET-FIELDS sub routine!")
		query := "describe " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class
		return repository.queryCommonMany(query, request)
	case "getClasses":
		request.Log("Starting GET-CLASSES sub routine")
		query := "SELECT DISTINCT TABLE_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='" + repository.getDatabaseName(request.Controls.Namespace) + "';"
		return repository.queryCommonMany(query, request)
	case "getNamespaces":
		request.Log("Starting GET-NAMESPACES sub routine")
		query := "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME != 'information_schema' AND SCHEMA_NAME !='mysql' AND SCHEMA_NAME !='performance_schema';"
		return repository.queryCommonMany(query, request)
	case "getSelected":
		request.Log("Get-Selected not implemented in CloudSQL repository. Use Get-Query for custom querying in CloudSQL Repository")
		return getDefaultNotImplemented()
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
			response.GetSuccessResByObject(obj)
		} else {
			var empty map[string]interface{}
			empty = make(map[string]interface{})
			response.GetSuccessResByObject(empty)
		}
	} else {
		response.GetErrorResponse("Error connecting to CloudSQL : " + err.Error())
	}

	return response
}

func (repository CloudSqlRepository) queryCommonMany(query string, request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryCommon(query, request, false)

}

func (repository CloudSqlRepository) queryCommonOne(query string, request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryCommon(query, request, true)
}

func (repository CloudSqlRepository) queryStore(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	conn, _ := repository.getConnection(request)

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
				response.Message = "Error Updating All Objects in CloudSQL. Check Data!"
				return response
			}
		}

		response.IsSuccess = true
		response.Message = "Successfully stored object(s) in CloudSQL"

	} else {
		if err == nil {
			err := repository.executeNonQuery(conn, script)
			if err == nil {
				response.IsSuccess = true
				response.Message = "Successfully stored object(s) in CloudSQL"
			} else {
				response.IsSuccess = false
				response.Message = "Error storing data in CloudSQL : " + err.Error()
			}
		} else {
			response.IsSuccess = false
			response.Message = "Error generating CloudSQL query : " + err.Error()
		}
	}

	return response
}

func (repository CloudSqlRepository) getByKey(conn *sql.DB, namespace string, class string, id string) (obj map[string]interface{}) {
	query := "SELECT * FROM " + repository.getDatabaseName(namespace) + "." + class + " WHERE __os_id = \"" + id + "\""
	obj, _ = repository.executeQueryOne(conn, query, nil)
	return
}

func (repository CloudSqlRepository) getStoreScript(conn *sql.DB, request *messaging.ObjectRequest) (query string, err error) {
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

	query = ""

	isFirstRow := true
	var keyArray []string

	for _, obj := range allObjects {

		currentObject := repository.getByKey(conn, namespace, class, getNoSqlKeyById(request, obj))

		if currentObject == nil {
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
			//fmt.Println(keyArray)
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

		} else {
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
			query += ("UPDATE " + repository.getDatabaseName(namespace) + "." + class + " SET " + updateValues + " WHERE __os_id=\"" + getNoSqlKeyById(request, obj) + "\";###")
		}

		if isFirstRow {
			isFirstRow = false
		}
	}

	return
}

func (repository CloudSqlRepository) getDeleteScript(namespace string, class string, id string) string {
	return "DELETE FROM " + repository.getDatabaseName(namespace) + "." + class + " WHERE __os_id = \"" + id + "\""
}

func (repository CloudSqlRepository) getCreateScript(namespace string, class string, obj map[string]interface{}) string {
	query := "CREATE TABLE IF NOT EXISTS " + repository.getDatabaseName(namespace) + "." + class + "(__os_id TEXT"

	for k, v := range obj {
		if k != "OriginalIndex" {
			query += (", " + k + " " + repository.golangToSql(v))
		}
	}

	query += ")"
	fmt.Println(query)
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
		availableDbs[dbName] = true
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

			availableTables[dbName+"."+class] = true

		} else {
			return
		}
	}

	err = repository.buildTableCache(conn, dbName, class)

	alterColumns := ""
	cacheItem := tableCache[dbName+"."+class]
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
			tableCache[dbName+"."+class] = newMap
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

func (repository CloudSqlRepository) getConnection(request *messaging.ObjectRequest) (conn *sql.DB, err error) {
	connInt := connmanager.Get("MYSQL", request.Controls.Namespace)

	if connInt != nil {
		term.Write("Connection Already Available.. Pinging Now....", 2)
		temp := connInt.(*sql.DB)
		err2 := temp.Ping()
		if err2 != nil {
			term.Write(err2.Error(), 1)
			term.Write("Ping Failed! Creating a new Connection!", 2)
			var c *sql.DB
			mysqlConf := request.Configuration.ServerConfiguration["MYSQL"]
			c, err = sql.Open("mysql", mysqlConf["Username"]+":"+mysqlConf["Password"]+"@tcp("+mysqlConf["Url"]+":"+mysqlConf["Port"]+")/")
			connmanager.Set("MYSQL", request.Controls.Namespace, c)
			conn = c
			return
		}
		term.Write("Ping Successful! Reusing Same Connection!", 2)
		conn = temp
		//conn = connInt.(*sql.DB)

	} else {
		term.Write("!No Connection Found! Creating Brand New Connection!", 2)
		var c *sql.DB
		mysqlConf := request.Configuration.ServerConfiguration["MYSQL"]
		c, err = sql.Open("mysql", mysqlConf["Username"]+":"+mysqlConf["Password"]+"@tcp("+mysqlConf["Url"]+":"+mysqlConf["Port"]+")/")
		connmanager.Set("MYSQL", request.Controls.Namespace, c)
		conn = c
	}

	return
}

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
	//fmt.Println(reflect.TypeOf(value))
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
		strValue = "BLOB"
		break

	}

	return strValue
}

func (repository CloudSqlRepository) sqlToGolang(b []byte, t string) interface{} {
	if b == nil {
		return nil
	}

	if len(b) == 0 {
		return b
	}

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
	//case "text":
	//case "blob":
	default:
		if len(tmp) == 4 {
			if strings.ToLower(tmp) == "null" {
				outData = nil
				break
			}
		}

		/*
			var m map[string]interface{}
			var ml []map[string]interface{}


			if (string(tmp[0]) == "{"){
				err := json.Unmarshal([]byte(tmp), &m)
				if err == nil {
					outData = m
				}else{
					fmt.Println(err.Error())
					outData = tmp
				}
			}else if (string(tmp[0]) == "["){
				err := json.Unmarshal([]byte(tmp), &ml)
				if err == nil {
					outData = ml
				}else{
					fmt.Println(err.Error())
					outData = tmp
				}
			}else{
				outData = tmp
			}
		*/
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
	fmt.Println(query)
	var stmt *sql.Stmt
	stmt, err = conn.Prepare(query)
	_, err = stmt.Exec()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Successfully Executed Query!")
	}
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
		returnID = uuid.NewV1().String()
	} else if isAutoIncrementId {
		session, isError := repository.getConnection(request)
		if isError != nil {
			returnID = ""
			return
		} else {
			//Reading maxCount from DB
			checkTableQuery := "SELECT DISTINCT TABLE_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='" + repository.getDatabaseName(request.Controls.Namespace) + "' AND TABLE_NAME='domainClassAttributes';"
			tableResultMap, _ := repository.executeQueryOne(session, checkTableQuery, request.Controls.Class)
			if len(tableResultMap) == 0 {
				//Create new domainClassAttributes  table
				createDomainAttrQuery := "create table " + repository.getDatabaseName(request.Controls.Namespace) + ".domainClassAttributes ( class VARCHAR(255) primary key, maxCount text, version text);"
				err := repository.executeNonQuery(session, createDomainAttrQuery)
				if err != nil {
					returnID = ""
					return
				} else {
					//insert record with count 1 and return
					insertQuery := "INSERT INTO " + repository.getDatabaseName(request.Controls.Namespace) + ".domainClassAttributes (class, maxCount,version) VALUES ('" + strings.ToLower(request.Controls.Class) + "','1','" + uuid.NewV1().String() + "')"
					err = repository.executeNonQuery(session, insertQuery)
					if err != nil {
						returnID = ""
						return
					} else {
						returnID = "1"
						return
					}
				}
			} else {
				//This is a new Class.. Create New entry
				readQuery := "SELECT maxCount FROM " + getMySQLnamespace(request) + ".domainClassAttributes where class = '" + strings.ToLower(request.Controls.Class) + "';"
				myMap, _ := repository.executeQueryOne(session, readQuery, (getMySQLnamespace(request) + ".domainClassAttributes"))

				if len(myMap) == 0 {
					request.Log("New Class! New record for this class will be inserted")
					insertNewClassQuery := "INSERT INTO " + getMySQLnamespace(request) + ".domainClassAttributes (class,maxCount,version) values ('" + strings.ToLower(request.Controls.Class) + "', '1', '" + uuid.NewV1().String() + "');"
					err := repository.executeNonQuery(session, insertNewClassQuery)
					if err != nil {
						returnID = ""
						return
					} else {
						returnID = "1"
						return
					}
				} else {
					//Inrement one and UPDATE
					maxCount := 0
					maxCount, err := strconv.Atoi(myMap["maxCount"].(string))
					maxCount++
					returnID = strconv.Itoa(maxCount)
					updateQuery := "UPDATE " + getMySQLnamespace(request) + ".domainClassAttributes SET maxCount='" + returnID + "' WHERE class = '" + strings.ToLower(request.Controls.Class) + "' ;"
					err = repository.executeNonQuery(session, updateQuery)
					if err != nil {
						returnID = ""
						return
					}
				}
			}
		}
	} else {
		returnID = obj[request.Body.Parameters.KeyProperty].(string)
	}
	return
}
