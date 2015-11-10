package repositories

import (
	"database/sql"
	"duov6.com/objectstore/connmanager"
	"duov6.com/objectstore/messaging"
	"duov6.com/term"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	//"github.com/twinj/uuid"
	"strconv"
	"strings"
)

type PostgresRepository struct {
}

//var availableDbs map[string]interface{} //:= make(map[string]bool)
//var availableTables map[string]interface{}
//var tableCache map[string]map[string]string

func (repository PostgresRepository) GetRepositoryName() string {
	return "Postgres DB"
}

func (repository PostgresRepository) getDatabaseName(namespace string) string {
	return ("_" + strings.Replace(namespace, ".", "", -1))
}

func (repository PostgresRepository) getConnection(request *messaging.ObjectRequest) (session *sql.DB, isError bool, errorMessage string) {
	connInt := connmanager.Get("POSTGRES", request.Controls.Namespace)

	if connInt != nil {
		term.Write("Connection Found!", 2)
		session = connInt.(*sql.DB)
		isError = false
	} else {
		term.Write("Connection Not Found! Creating New Postgres Connection!", 2)
		isError = false
		username := request.Configuration.ServerConfiguration["POSTGRES"]["Username"]
		password := request.Configuration.ServerConfiguration["POSTGRES"]["Password"]
		dbUrl := request.Configuration.ServerConfiguration["POSTGRES"]["Url"]
		dbPort := request.Configuration.ServerConfiguration["POSTGRES"]["Port"]

		session, err := sql.Open("postgres", "host="+dbUrl+" port="+dbPort+" user="+username+" password="+password+" dbname="+"postgres"+" sslmode=disable")

		if err != nil {
			isError = true
			term.Write(err.Error(), 1)
			errorMessage = err.Error()
		}

		//Create schema if not available.
		term.Write("Checking if Database "+repository.getDatabaseName(request.Controls.Namespace)+" is available.", 2)

		isDatabaseAvailbale := false

		rows, err := session.Query("SELECT datname FROM pg_database WHERE datistemplate = false;")

		if err != nil {
			term.Write(err.Error(), 1)
		} else {
			term.Write("Successfully retrieved values for all objects in PostGres", 2)

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
					term.Write("Check domain : "+repository.getDatabaseName(request.Controls.Namespace)+" : available schema : "+v.(string), 2)
					if v.(string) == repository.getDatabaseName(request.Controls.Namespace) {
						//Database available
						isDatabaseAvailbale = true
						break
					}
				}
			}
		}

		if isDatabaseAvailbale {
			term.Write("Database already available. Nothing to do. Proceed!", 2)
			session.Close()
			session, err = sql.Open("postgres", "host="+dbUrl+" port="+dbPort+" user="+username+" password="+password+" dbname="+(repository.getDatabaseName(request.Controls.Namespace))+" sslmode=disable")
			if err != nil {
				term.Write(err.Error(), 1)
				isError = true
			} else {
				isError = false
				errorMessage = ""
				term.Write("Already Relogin successful!", 2)
			}

		} else {
			_, err = session.Query("CREATE DATABASE " + repository.getDatabaseName(request.Controls.Namespace) + ";")
			if err != nil {
				term.Write(err.Error(), 1)
				isError = true
			} else {
				term.Write("Creation of domain matched Schema Successful", 2)
				session.Close()
				session, err = sql.Open("postgres", "host="+dbUrl+" port="+dbPort+" user="+username+" password="+password+" dbname="+(repository.getDatabaseName(request.Controls.Namespace))+" sslmode=disable")
				if err != nil {
					term.Write(err.Error(), 1)
					isError = true
				} else {
					term.Write("Relogin successful!", 2)
					isError = false
				}

			}
		}

		return session, isError, errorMessage
	}
	term.Write("Reusing existing Postgres connection", 2)
	return
}

func (repository PostgresRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
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

	query := "SELECT * FROM " + request.Controls.Class
	if isTakable {
		query += " limit " + take
	}
	if isSkippable {
		query += " offset " + skip
	}

	return repository.queryCommonMany(query, request)
}

func (repository PostgresRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Get-Search!", 2)
	response := RepositoryResponse{}
	query := ""
	if strings.Contains(request.Body.Query.Parameters, ":") {
		tokens := strings.Split(request.Body.Query.Parameters, ":")
		fieldName := tokens[0]
		fieldValue := tokens[1]
		query = "select * from " + request.Controls.Class + " where " + fieldName + "='" + fieldValue + "';"
	} else {
		query = "select * from " + request.Controls.Class + ";"
	}
	request.Log(query)
	response = repository.queryCommonMany(query, request)
	return response
}

func (repository PostgresRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Get-Query!", 2)
	response := RepositoryResponse{}
	if request.Body.Query.Parameters != "*" {
		query := request.Body.Query.Parameters
		term.Write(query, 2)
		response = repository.queryCommonMany(query, request)
	} else {
		response = repository.GetAll(request)
	}
	return response
}

func (repository PostgresRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Get-By-Key!", 2)
	query := "SELECT * FROM " + request.Controls.Class + " WHERE __os_id = '" + getNoSqlKey(request) + "'"
	term.Write(query, 2)
	return repository.queryCommonOne(query, request)
}

func (repository PostgresRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Insert-Multiple!", 2)
	//return repository.queryStore(request)
	var idData map[string]interface{}
	idData = make(map[string]interface{})

	for index, obj := range request.Body.Objects {
		//id := repository.getRecordID(request, obj)
		fmt.Println(obj)
		id := request.Body.Objects[index][request.Body.Parameters.KeyProperty]
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

func (repository PostgresRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	term.Write("Executing Insert-Single!", 2)
	//return repository.queryStore(request)
	//id := repository.getRecordID(request, request.Body.Object)
	id := request.Body.Object[request.Body.Parameters.KeyProperty]
	request.Controls.Id = id.(string)
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

func (repository PostgresRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryStore(request)
}

func (repository PostgresRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryStore(request)
}

func (repository PostgresRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	conn, err, _ := repository.getConnection(request)
	if !err {
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
			response.Message = "Successfully Deleted all objects from Postgres repository!"
		}
	} else {
		response.IsSuccess = false
		response.Message = "Error deleting all objects!"
	}
	return response
}

func (repository PostgresRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	conn, err, _ := repository.getConnection(request)
	if !err {
		query := repository.getDeleteScript(request.Controls.Namespace, request.Controls.Class, getNoSqlKey(request))
		fmt.Println(query)
		err := repository.executeNonQuery(conn, query)
		if err != nil {
			response.IsSuccess = false
			response.Message = "Failed Deleting from Postgres repository : " + err.Error()
		} else {
			response.IsSuccess = true
			response.Message = "Successfully Deleted from Postgres repository!"
		}
	} else {
		response.IsSuccess = false
		response.Message = "Failed Deleting from Postgres repository"
	}
	return response
}

func (repository PostgresRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	queryType := request.Body.Special.Type
	switch queryType {
	case "getFields":
		request.Log("Starting GET-FIELDS sub routine!")
		query := "select column_name from information_schema.columns where table_name='" + request.Controls.Class + "';"
		return repository.queryCommonMany(query, request)
	case "getClasses":
		request.Log("Starting GET-CLASSES sub routine")
		query := "SELECT table_name FROM information_schema.tables WHERE table_schema='public';"
		return repository.queryCommonMany(query, request)
	case "getNamespaces":
		request.Log("Starting GET-NAMESPACES sub routine")
		query := "SELECT datname FROM pg_database WHERE datistemplate = false;"
		return repository.queryCommonMany(query, request)
	case "getSelected":
		request.Log("Get-Selected not implemented in Postgres repository. Use Get-Query for custom querying in Postgres Repository")
		return getDefaultNotImplemented()
	case "DropClass":
		request.Log("Starting Delete-Class sub routine")
		conn, err, _ := repository.getConnection(request)
		if !err {
			query := "DROP TABLE " + request.Controls.Class
			err := repository.executeNonQuery(conn, query)
			if err != nil {
				response.IsSuccess = false
				response.Message = "Error Dropping Table in Postgres Repository : " + err.Error()
			} else {
				delete(availableTables, (repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class))
				delete(tableCache, (repository.getDatabaseName(request.Controls.Namespace) + "." + request.Controls.Class))
				response.IsSuccess = true
				response.Message = "Successfully Dropped Table : " + request.Controls.Class
			}
		} else {
			response.IsSuccess = false
			response.Message = "Connection Failed to Postgres Server"
		}
	case "DropNamespace":
		request.Log("Starting Delete-Database sub routine")
		conn, err, _ := repository.getConnection(request)
		if !err {
			query := "DROP DATABASE " + repository.getDatabaseName(request.Controls.Namespace)
			err := repository.executeNonQuery(conn, query)
			if err != nil {
				response.IsSuccess = false
				response.Message = "Error Dropping Table in Postgres Repository : " + err.Error()
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
			response.Message = "Connection Failed to Postgres Server"
		}
	default:
		return repository.GetAll(request)

	}

	return response
}

func (repository PostgresRepository) Test(request *messaging.ObjectRequest) {
}

//Sub Routines

/*
func (repository PostgresRepository) getRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
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
		session, isError, _ := repository.getConnection(request)
		if isError {
			returnID = ""
			request.Log("Connecting to MySQL Failed!")
		} else {
			//Read Table domainClassAttributes
			request.Log("Reading maxCount from DB")
			rows, err := session.Query("SELECT maxCount FROM domainClassAttributes where class = '" + strings.ToLower(request.Controls.Class) + "';")

			if err != nil {
				//If err create new domainClassAttributes  table
				request.Log("No Class found.. Must be a new namespace")
				_, err = session.Query("create table domainClassAttributes ( class text primary key, maxcount text, version text);")
				if err != nil {
					returnID = ""
					return
				} else {
					//insert record with count 1 and return
					_, err := session.Query("INSERT INTO domainClassAttributes (class, maxcount,version) VALUES ('" + strings.ToLower(request.Controls.Class) + "','1','" + uuid.NewV1().String() + "')")
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
					_, err = session.Query("INSERT INTO domainClassAttributes (class,maxcount,version) values ('" + strings.ToLower(request.Controls.Class) + "', '1', '" + uuid.NewV1().String() + "');")
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
					_, err = session.Query("UPDATE domainClassAttributes SET maxcount='" + returnID + "' WHERE class = '" + strings.ToLower(request.Controls.Class) + "' ;")
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
*/
func (repository PostgresRepository) queryCommonMany(query string, request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryCommon(query, request, false)
}

func (repository PostgresRepository) queryCommonOne(query string, request *messaging.ObjectRequest) RepositoryResponse {
	return repository.queryCommon(query, request, true)
}

func (repository PostgresRepository) queryCommon(query string, request *messaging.ObjectRequest, isOne bool) RepositoryResponse {
	response := RepositoryResponse{}

	conn, err, _ := repository.getConnection(request)
	if !err {
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
		response.GetErrorResponse("Error connecting to Postgres")
	}

	return response
}

func (repository PostgresRepository) buildTableCache(conn *sql.DB, dbName string, class string) (err error) {
	if tableCache == nil {
		tableCache = make(map[string]map[string]string)
	}

	_, ok := tableCache[dbName+"."+class]

	if !ok {
		var exResult []map[string]interface{}
		exResult, err = repository.executeQueryMany(conn, "select column_name, data_type from information_schema.columns where table_name = '"+strings.ToLower(class)+"';", nil)
		if err == nil {
			newMap := make(map[string]string)

			for _, cRow := range exResult {
				newMap[cRow["column_name"].(string)] = cRow["data_type"].(string)
			}
			tableCache[dbName+"."+class] = newMap
		}
	}

	return
}

func (repository PostgresRepository) executeQueryOne(conn *sql.DB, query string, tableName interface{}) (result map[string]interface{}, err error) {
	rows, err := conn.Query(query)

	if err == nil {
		var resultSet []map[string]interface{}
		resultSet, err = repository.rowsToMap(rows, tableName)
		if len(resultSet) > 0 {
			result = resultSet[0]
		}

	} else {
		term.Write(err.Error(), 1)
		if strings.HasPrefix(err.Error(), "does not exist") {
			err = nil
			result = make(map[string]interface{})
		}
	}

	return
}

func (repository PostgresRepository) executeQueryMany(conn *sql.DB, query string, tableName interface{}) (result []map[string]interface{}, err error) {
	fmt.Println(query)
	rows, err := conn.Query(query)

	if err == nil {
		result, err = repository.rowsToMap(rows, tableName)
	} else {
		if strings.HasPrefix(err.Error(), "does not exist") {
			err = nil
			result = make([]map[string]interface{}, 0)
		}
	}

	return
}

func (repository PostgresRepository) rowsToMap(rows *sql.Rows, tableName interface{}) (tableMap []map[string]interface{}, err error) {

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

func (repository PostgresRepository) sqlToGolang(b []byte, t string) interface{} {
	if b == nil {
		return nil
	}

	if len(b) == 0 {
		return b
	}

	var outData interface{}

	tmp := string(b)
	switch t {
	case "boolean":
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
	case "double precision":
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

func (repository PostgresRepository) queryStore(request *messaging.ObjectRequest) RepositoryResponse {
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
				response.Message = "Error Updating All Objects in Postgres. Check Data!"
				return response
			}
		}

		response.IsSuccess = true
		response.Message = "Successfully stored object(s) in Postgres"

	} else {
		if err == nil {
			err := repository.executeNonQuery(conn, script)
			if err == nil {
				response.IsSuccess = true
				response.Message = "Successfully stored object(s) in Postgres"
			} else {
				response.IsSuccess = false
				response.Message = "Error storing data in Postgres : " + err.Error()
			}
		} else {
			response.IsSuccess = false
			response.Message = "Error generating Postgres query : " + err.Error()
		}
	}

	return response
}

func (repository PostgresRepository) getByKey(conn *sql.DB, namespace string, class string, id string) (obj map[string]interface{}) {
	query := "SELECT * FROM " + strings.ToLower(class) + " WHERE __os_id = '" + id + "'"
	fmt.Println(query)
	obj, _ = repository.executeQueryOne(conn, query, nil)
	return
}

func (repository PostgresRepository) getStoreScript(conn *sql.DB, request *messaging.ObjectRequest) (query string, err error) {
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
				query += ("INSERT INTO " + class)
			}

			id := ""

			if obj["OriginalIndex"] == nil {
				id = getNoSqlKeyById(request, obj)
			} else {
				id = obj["OriginalIndex"].(string)
			}

			keyList := ""
			valueList := ""

			if isFirstRow {
				for k, _ := range obj {
					keyList += ("," + k)
					keyArray = append(keyArray, k)
				}
			}
			fmt.Println(keyArray)
			for _, k := range keyArray {
				v := obj[k]
				valueList += ("," + repository.getSqlFieldValue(v))
			}

			if isFirstRow {
				query += "(__os_id" + keyList + ") VALUES "
			} else {
				query += ","
			}

			//query += ("('" + getNoSqlKeyById(request, obj) + "'" + valueList + ")")
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

				updateValues += (k + "=" + repository.getSqlFieldValue(v))
			}
			query += ("UPDATE " + class + " SET " + updateValues + " WHERE __os_id='" + getNoSqlKeyById(request, obj) + "';###")
		}

		if isFirstRow {
			isFirstRow = false
		}
	}

	return
}

func (repository PostgresRepository) getDeleteScript(namespace string, class string, id string) string {
	return "DELETE FROM " + strings.ToLower(class) + " WHERE __os_id = '" + id + "'"
}

func (repository PostgresRepository) getCreateScript(namespace string, class string, obj map[string]interface{}) string {
	query := "CREATE TABLE IF NOT EXISTS " + class + "(__os_id TEXT"
	for k, v := range obj {
		query += (", " + k + " " + repository.golangToSql(v))
	}
	query += ")"
	return query
}

func (repository PostgresRepository) checkAvailabilityDb(conn *sql.DB, dbName string) (err error) {
	if availableDbs == nil {
		availableDbs = make(map[string]interface{})
	}

	if availableDbs[dbName] != nil {
		return
	}

	dbQuery := "SELECT datname FROM pg_database WHERE datistemplate = false AND datname='" + dbName + "';"
	dbResult, err := repository.executeQueryOne(conn, dbQuery, nil)

	if err == nil {
		if dbResult["datname"] == nil {
			repository.executeNonQuery(conn, "CREATE DATABASE "+dbName+";")
		}
		availableDbs[dbName] = true
	} else {
		term.Write(err.Error(), 1)
	}

	return
}

func (repository PostgresRepository) checkAvailabilityTable(conn *sql.DB, dbName string, namespace string, class string, obj map[string]interface{}) (err error) {

	if availableTables == nil {
		availableTables = make(map[string]interface{})
	}

	if availableTables[dbName+"."+class] == nil {
		var tableResult map[string]interface{}
		tableResult, err = repository.executeQueryOne(conn, "SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_name like '"+strings.ToLower(class)+"';", nil)
		fmt.Println(tableResult)
		if err == nil {
			if tableResult["table_name"] == nil {
				script := repository.getCreateScript(namespace, class, obj)
				term.Write(script, 2)
				err = repository.executeNonQuery(conn, script)

				if err != nil {
					term.Write(err.Error(), 1)
					return
				}
			}

			availableTables[dbName+"."+class] = true

		} else {
			return
		}
	}
	fmt.Println(1)
	err = repository.buildTableCache(conn, dbName, class)
	fmt.Println(2)
	alterColumns := ""
	cacheItem := tableCache[dbName+"."+class]
	isFirst := true
	for k, v := range obj {
		_, ok := cacheItem[strings.ToLower(k)]
		if !ok {
			if isFirst {
				isFirst = false
			} else {
				alterColumns += ", "
			}

			alterColumns += ("ADD COLUMN " + k + " " + repository.golangToSql(v))
		}
	}

	if len(alterColumns) != 0 {
		alterQuery := "ALTER TABLE " + class + " " + alterColumns
		term.Write(alterQuery, 2)
		err = repository.executeNonQuery(conn, alterQuery)
	}

	return
}

func (repository PostgresRepository) checkSchema(conn *sql.DB, namespace string, class string, obj map[string]interface{}) {
	dbName := repository.getDatabaseName(namespace)
	err := repository.checkAvailabilityDb(conn, dbName)

	if err == nil {
		err := repository.checkAvailabilityTable(conn, dbName, namespace, class, obj)

		if err != nil {
			term.Write(err.Error(), 1)
		}
	}
}

func (repository PostgresRepository) getJson(m interface{}) string {
	bytes, _ := json.Marshal(m)
	return string(bytes[:len(bytes)])
}

func (repository PostgresRepository) getSqlFieldValue(value interface{}) string {
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

func (repository PostgresRepository) golangToSql(value interface{}) string {
	var strValue string
	//fmt.Println(reflect.TypeOf(value))
	switch value.(type) {
	case string:
		strValue = "TEXT"
	case bool:
		strValue = "Boolean"
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
		strValue = "bigint"
		break
	case float32:
	case float64:
		strValue = "double precision"
		break
	default:
		strValue = "bytea"
		break

	}

	return strValue
}

func (repository PostgresRepository) getInterfaceValue(tmp string) (outData interface{}) {
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

func (repository PostgresRepository) executeNonQuery(conn *sql.DB, query string) (err error) {
	var stmt *sql.Stmt
	fmt.Println(query)
	stmt, err = conn.Prepare(query)
	_, err = stmt.Exec()
	if err != nil {
		fmt.Println(err.Error())
		term.Write(err.Error(), 1)
	}
	return
}
