package repositories

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/connmanager"
	"database/sql";
	_ "github.com/go-sql-driver/mysql";
	"fmt";
	"encoding/json";
	"strings";
	"errors";
	"strconv"
)

type CloudSqlRepository struct {
}

func (repository CloudSqlRepository) GetRepositoryName() string {
	return "CloudSQL"
}

func (repository CloudSqlRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	query := "SELECT * FROM " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class
	return repository.queryCommonMany(query, request)
}

func (repository CloudSqlRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	query := request.Body.Query.Parameters
	return repository.queryCommonMany(query, request)
}

func (repository CloudSqlRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	query := "SELECT * FROM " + repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class + " WHERE __os_id = \"" + getNoSqlKey(request) + "\""
	return repository.queryCommonOne(query, request)
}

func (repository CloudSqlRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("GetSearch not implemented in CloudSQL repository")
	return getDefaultNotImplemented()
}

func (repository CloudSqlRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryStore(request)
}

func (repository CloudSqlRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryStore(request)
}

func (repository CloudSqlRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryStore(request)
}

func (repository CloudSqlRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryStore(request)
}

func (repository CloudSqlRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("DeleteMultiple not implemented in CloudSQL repository")
	return getDefaultNotImplemented()
}

func (repository CloudSqlRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("DeleteSingle not implemented in CloudSQL repository")
	return getDefaultNotImplemented()
}

func (repository CloudSqlRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Special not implemented in CloudSQL repository")
	return getDefaultNotImplemented()
}

func (repository CloudSqlRepository) Test(request *messaging.ObjectRequest) {

}

////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////SQL GENERATORS/////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////
func (repository CloudSqlRepository) queryCommon(query string, request *messaging.ObjectRequest, isOne bool) RepositoryResponse{
	response := RepositoryResponse{}

	conn,err:= repository.getConnection(request)
	if err==nil{
		var err error
		dbName := repository.getDatabaseName(request.Controls.Namespace)
		err = repository.buildTableCache(conn, dbName , request.Controls.Class)
		if err !=nil{

		}
		
		var obj interface{}
		tableName := dbName + "."+  request.Controls.Class
		if isOne{
			obj, err = repository.executeQueryOne(conn, query, tableName)
		}else{
			obj, err = repository.executeQueryMany(conn, query, tableName)
		}
		
		if err==nil{
			response.GetSuccessResByObject(obj)
		}else{
			response.GetErrorResponse("Error querying from CloudSQL : " + err.Error())
		}
	} else{
		response.GetErrorResponse("Error connecting to CloudSQL : " + err.Error())
	}

	return response
}

func (repository CloudSqlRepository) queryCommonMany(query string, request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryCommon(query, request, false)

}

func (repository CloudSqlRepository) queryCommonOne(query string, request *messaging.ObjectRequest) RepositoryResponse{
	return repository.queryCommon(query, request, true)
}

func (repository CloudSqlRepository) queryStore(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	conn, _ := repository.getConnection(request);

	script,err := repository.getStoreScript(conn, request)
	
	if err ==nil{
		//fmt.Println(script)
		err := repository.executeNonQuery(conn, script);
		if err ==nil{
			response.IsSuccess = true
			response.Message = "Successfully stored object(s) in CloudSQL"
		}else {
			response.IsSuccess = false
			response.Message = "Error storing data in CloudSQL : " + err.Error()
		}
	}else{
		response.IsSuccess = true
		response.Message = "Error generating CloudSQL query : " + err.Error()
	}

	return response
}

func (repository CloudSqlRepository) getByKey(conn *sql.DB, namespace string, class string, id string) (obj map[string]interface{}){
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
		allObjects = make([]map[string]interface{} ,1)
		allObjects[0] = schemaObj
	}else {
		if (request.Body.Objects !=nil){
			if (len(request.Body.Objects) != 0) {
				schemaObj = request.Body.Objects [0]
				allObjects = request.Body.Objects
			}else{
				err = errors.New("No objects available to store")
				return
			}
		}else{
			err = errors.New("No objects available to store")
			return
		}
		
	}
	
	repository.checkSchema(conn, namespace, class, schemaObj);

	query = ""

	for _, obj := range allObjects {

		currentObject := repository.getByKey(conn, namespace, class, getNoSqlKey(request))

		if currentObject == nil{
			query += ("INSERT INTO " + repository.getDatabaseName(namespace) + "." + class)
			keyList :="";
			valueList :="";
			for k,v := range obj {
				keyList += ("," +k);
				valueList += ("," + repository.getSqlFieldValue(v));
			}
			query += ( " (__os_id" + keyList + ") VALUES (\""+ getNoSqlKey(request) + "\"" + valueList + ");")
		} else {
			updateValues :="";
			isFirst := true;
			for k,v := range obj {
				if isFirst{
					isFirst = false
				}else{
					updateValues += ","
				}

				updateValues += (k + "=" + repository.getSqlFieldValue(v));
			}
			query += ("UPDATE " + repository.getDatabaseName(namespace) + "." + class + " SET " + updateValues + " WHERE __os_id=\"" + getNoSqlKey(request) + "\";")
		}
	}

	return
}

func (repository CloudSqlRepository) getDeleteScript(namespace string, class string, id string) string {
	return "DELETE * FROM " + repository.getDatabaseName(namespace) + "." + class + " WHERE __os_id = \"" + id + "\""
}

func (repository CloudSqlRepository) getCreateScript(namespace string, class string, obj map[string]interface{}) string {
	query := "CREATE TABLE " + repository.getDatabaseName(namespace) + "." + class +"(__os_id TEXT"

	for k,v:= range obj{
		query += (", " +k + " " + repository.golangToSql(v))
	}

	query += ")"
	return query
}


var availableDbs map[string]interface{} //:= make(map[string]bool)
var availableTables map[string]interface{}
var tableCache map[string]map[string]string

func (repository CloudSqlRepository) checkAvailabilityDb(conn *sql.DB, dbName string) (err error){
	if availableDbs == nil{
		availableDbs = make (map[string]interface{})
	}

	if (availableDbs[dbName] != nil){
		return
	} 

	dbQuery := "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '" +  dbName + "'"
	dbResult, err := repository.executeQueryOne(conn, dbQuery, nil)

	if err == nil{
		if (dbResult["SCHEMA_NAME"] == nil) {
			repository.executeNonQuery(conn, "CREATE DATABASE IF NOT EXISTS " + dbName)
		}
		availableDbs[dbName] = true
	}else{
		fmt.Println(err.Error())
	}

	return 
}

func (repository CloudSqlRepository) checkAvailabilityTable(conn *sql.DB, dbName string, namespace string, class string, obj map[string]interface{}) (err error){
	
	if availableTables == nil{
		availableTables = make (map[string]interface{})
	}

	
	if (availableTables[dbName + "." + class] == nil){
		var tableResult map[string]interface{}
		tableResult, err = repository.executeQueryOne(conn, "SHOW TABLES FROM " + dbName + " LIKE \"" + class +"\"", nil)

		if (err ==nil){
			if tableResult["Tables_in_" + dbName] == nil {
				script := repository.getCreateScript(namespace, class, obj);
				err = repository.executeNonQuery(conn, script)	

				if err != nil{
					return
				}			
			}

			availableTables[dbName + "." + class] = true
			
		} else {
			return
		}
	} 
	

	err = repository.buildTableCache(conn, dbName, class)

	alterColumns := ""
	cacheItem := tableCache[dbName + "." + class]
	isFirst := true
	for k,v := range obj {
		_, ok := cacheItem[k]
		if (!ok){
			if isFirst{
				isFirst = false
			}else{
				alterColumns  += ", "
			}

			alterColumns += ("ADD COLUMN " + k + " " + repository.golangToSql(v))
		}
	}

	if len(alterColumns) !=0{
		alterQuery := "ALTER TABLE " + dbName + "." + class + " " + alterColumns
		err = repository.executeNonQuery(conn, alterQuery)
	}

	return
}

func (repository CloudSqlRepository) buildTableCache(conn *sql.DB, dbName string, class string) (err error){
	if tableCache == nil{
			tableCache = make (map[string]map[string]string)
	}
	
	_,ok := tableCache[dbName + "." + class]

	if (!ok){
		var exResult []map[string]interface{}
		exResult, err = repository.executeQueryMany(conn, "EXPLAIN " + dbName + "." + class, nil)	
		if err == nil {
			newMap := make(map[string]string)

			for _,cRow := range exResult {
				newMap[cRow["Field"].(string)] = cRow["Type"].(string)
			}
			tableCache[dbName + "." + class] = newMap
		}
	}

	return
}

func (repository CloudSqlRepository) checkSchema(conn *sql.DB, namespace string, class string, obj map[string]interface{}){
	dbName := repository.getDatabaseName(namespace)
	err := repository.checkAvailabilityDb (conn, dbName)
	
	if err ==nil{
		err := repository.checkAvailabilityTable(conn, dbName ,namespace, class, obj)

		if err !=nil{

		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////Helper functions/////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////

func (repository CloudSqlRepository) getConnection(request *messaging.ObjectRequest)(conn *sql.DB,  err error){
	connInt := connmanager.Get("MYSQL", request.Controls.Namespace);
	
	if (connInt != nil){
		conn = connInt.(*sql.DB)
	}else{
		var c *sql.DB
		mysqlConf := request.Configuration.ServerConfiguration["MYSQL"]
		c, err = sql.Open("mysql", mysqlConf["Username"]+":"+mysqlConf["Password"]+"@tcp("+mysqlConf["Url"]+":"+mysqlConf["Port"]+")/")
		connmanager.Set("MYSQL", request.Controls.Namespace , c)
		conn = c
	}

	return
}

func (repository CloudSqlRepository) getDatabaseName(namespace string) string {
	return "_" + strings.ToLower(strings.Replace(namespace, ".", "", -1))
}

func (repository CloudSqlRepository) getJson(m interface{}) string{
	bytes,_ := json.Marshal(m)
	return string(bytes[:len(bytes)])
}

func (repository CloudSqlRepository) getSqlFieldValue(value interface{}) string{
	var strValue string
    switch v := value.(type) {
    		
    		case []map[string]interface{}:
    		case []interface{}:
            case map[string]interface{}:
            	strValue = "'" + repository.getJson(v) + "'"
            	break;
            case string:
            	strValue = "'" + fmt.Sprint(value) + "'"
            	break;
            default:
            	strValue = "'" + repository.getJson(v) + "'";
                break;
                
    }

    return strValue;
}


func (repository CloudSqlRepository) golangToSql(value interface{}) string{
	var strValue string
    switch value.(type) {
    		case string:
    			strValue = "TEXT"
            case bool:
            	strValue = "BIT"
            	break;
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
				break;
			case float32:
			case float64:
				strValue = "DOUBLE"
				break;
    		case []map[string]interface{}:
    		case []interface{}:
            case map[string]interface{}:
            	strValue = "BLOB"
            	break;
            default:
            	strValue = "TEXT"
                break;
                
    }

    return strValue;
}

func (repository CloudSqlRepository) sqlToGolang(b []byte, t string) (interface{}) {
	if b == nil{
		return nil
	}

	if len(b) ==0 {
		return b
	}

	var outData interface{}

	tmp := string(b)
	switch (t){
		case "bit(1)":
			fmt.Println("BIT : " + t)
			if (tmp == "1"){
				outData = true
			}else{
				outData = false
			}
			break
		case "text":
			outData = tmp
			break
		case "double":
			fData,err := strconv.ParseFloat(tmp,64)
			if err !=nil{
				fmt.Println(err.Error())
				outData = tmp
			}else{
				outData=fData
			}
			break
		case "blob":
			outData = tmp
			break
	}
	

	return outData
}

func (repository CloudSqlRepository) rowsToMap(rows *sql.Rows, tableName interface{}) (tableMap []map[string]interface{}, err error) {
	
	columns, _ := rows.Columns()
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

	var cacheItem map[string]string

	if (tableName !=nil){
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
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				//fmt.Println("cacheItem")
				//fmt.Println(tableCache)
				if (cacheItem !=nil){
					t, ok := cacheItem[col]
					if (ok){
						v  = repository.sqlToGolang(b, t)
					}
				}

				if v == nil{
					v = string(b)
				}
			} else {
				v = val
			}
			rowMap[col] = v
		}
		tableMap = append (tableMap, rowMap)
	}

	return;
}


func (repository CloudSqlRepository) executeQueryMany(conn *sql.DB, query string, tableName interface{}) (result []map[string]interface{}, err error){
	rows, err := conn.Query(query)

	if err == nil{
		result, err = repository.rowsToMap(rows, tableName);
	}else {
		if (strings.HasPrefix(err.Error(), "Error 1146")){
			err = nil
			result = make ([]map[string]interface{},0)
		}
	}

	return
}

func (repository CloudSqlRepository) executeQueryOne(conn *sql.DB, query string, tableName interface{}) (result map[string]interface{}, err error){
	rows, err := conn.Query(query)

	if err ==nil{
		var resultSet []map[string]interface{}
		resultSet, err = repository.rowsToMap(rows, tableName);
		if (len(resultSet) > 0 ){
			result = resultSet[0]
		}


	}else{
		if (strings.HasPrefix(err.Error(), "Error 1146")){
			err = nil
			result = make (map[string]interface{})
		}
	}

	return
}

func (repository CloudSqlRepository) executeNonQuery(conn *sql.DB, query string) (err error){
	var stmt *sql.Stmt
    stmt, err = conn.Prepare(query);
    _, err = stmt.Exec()
    return
}
