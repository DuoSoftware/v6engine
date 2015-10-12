package processes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

func ExportToMySQLServer(ipAddress string, username string, password string) (status bool) {
	status = true
	for _, value := range GetBackupFileList() {
		content, _ := ioutil.ReadFile(value)
		var array []map[string]interface{}
		_ = json.Unmarshal(content, &array)
		namespace, class := getNamespaceAndClass(value)
		status = InsertToMySQL(ipAddress, username, password, namespace, class, array)
	}
	return
}

func InsertToMySQL(ipAddress string, username string, password string, namespace string, class string, array []map[string]interface{}) (status bool) {

	session, isError, _ := getMysqlConnection(ipAddress, username, password, namespace)

	if isError == true {
		fmt.Println("Error Creating Connection to MySQL")
	} else {

		var DataObjects []map[string]interface{}
		DataObjects = make([]map[string]interface{}, len(array))

		//change osheaders
		for i := 0; i < len(array); i++ {
			var tempMapObject map[string]interface{}
			tempMapObject = make(map[string]interface{})

			for key, value := range array[i] {
				if key == "__osHeaders" {
					tempMapObject["osHeaders"] = value
				} else {
					tempMapObject[key] = value
				}
			}

			DataObjects[i] = tempMapObject
		}
		//check for table in MsSql
		if createMySQLTable(namespace, class, session, array) {
			fmt.Println("Table Verified Successfully!")
		} else {
			status = false
			return
		}

		indexNames := getMySQLFieldOrder(session, namespace, class)
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
				var keyArray = make([]string, noOfElements)
				var valueArray = make([]string, noOfElements)

				for index := 0; index < len(indexNames); index++ {

					if indexNames[index] != "osHeaders" {

						if _, ok := DataObjects[i][indexNames[index]].(string); ok {
							keyArray[index] = indexNames[index]
							valueArray[index] = DataObjects[i][indexNames[index]].(string)
						} else {
							//fmt.Println("Non string value detected, Will be strigified!")
							keyArray[index] = indexNames[index]
							valueArray[index] = getStringByObject(DataObjects[i][indexNames[index]])
						}
					} else {
						// __osHeaders Catched!
						keyArray[index] = "osHeaders"
						valueArray[index] = ConvertOsheaders(DataObjects[i][indexNames[index]].(ControlHeaders))
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
			_, err := session.Query("INSERT INTO " + getMySQLnamespace(namespace) + "." + strings.ToLower(class) + " (" + argKeyList + ") VALUES " + argValueList + ";")
			if err != nil {
				setStatus[statusIndex] = false
				fmt.Println("ERROR : " + err.Error())
			} else {
				fmt.Println("INSERTED SUCCESSFULLY")
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
				var keyArray = make([]string, noOfElements)
				var valueArray = make([]string, noOfElements)

				for index := 0; index < len(indexNames); index++ {
					if indexNames[index] != "osHeaders" {

						if _, ok := DataObjects[i][indexNames[index]].(string); ok {
							keyArray[index] = indexNames[index]
							valueArray[index] = DataObjects[i][indexNames[index]].(string)
						} else {
							//fmt.Println("Non string value detected, Will be strigified!")
							keyArray[index] = indexNames[index]
							valueArray[index] = getStringByObject(DataObjects[i][indexNames[index]])
						}
					} else {
						// __osHeaders Catched!
						keyArray[index] = "osHeaders"
						valueArray[index] = ConvertOsheaders(DataObjects[i][indexNames[index]].(ControlHeaders))
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
			_, err := session.Query("INSERT INTO " + getMySQLnamespace(namespace) + "." + strings.ToLower(class) + " (" + argKeyList + ") VALUES " + argValueList + ";")
			if err != nil {
				setStatus[statusIndex] = false
				fmt.Println("ERROR : " + err.Error())
			} else {
				fmt.Println("INSERTED SUCCESSFULLY")
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
			fmt.Println("Successfully inserted many objects in to MYSQL")
		} else {
			fmt.Println("Error inserting many objects in to MYSQL")
		}

	}

	session.Close()
	return
}

func getMySQLFieldOrder(session *sql.DB, Namespace string, class string) []string {
	var returnArray []string
	//read fields
	byteValue := executeMySqlGetFields(session, Namespace, class)

	err := json.Unmarshal(byteValue, &returnArray)
	if err != nil {
		fmt.Println("Converstion of Json Failed!")
		returnArray = make([]string, 1)
		returnArray[0] = "nil"
		return returnArray
	}

	return returnArray
}

func getMySQLnamespace(namespace string) string {
	namespace = strings.Replace(namespace, ".", "", -1)
	return "_" + strings.ToLower(namespace)
}

func getMysqlConnection(ipaddress string, username string, password string, namespace string) (session *sql.DB, isError bool, errorMessage string) {
	ipTokens := strings.Split(ipaddress, ":")
	//creating database out of namespace

	server := ipTokens[0]
	port := ipTokens[1]

	session, err := sql.Open("mysql", username+":"+password+"@tcp("+server+":"+port+")/")

	if err != nil {
		fmt.Println("Failed to create connection to MySql! : " + err.Error())
	} else {
		fmt.Println("Successfully created connection to MySql!")
	}

	//Create schema if not available.
	fmt.Println("Checking if Database " + getMySQLnamespace(namespace) + " is available.")

	isDatabaseAvailbale := false

	rows, err := session.Query("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = '" + getMySQLnamespace(namespace) + "'")

	if err != nil {
		fmt.Println("Error contacting Mysql Server to fetch available databases")
	} else {
		fmt.Println("Successfully retrieved values for all objects in MySQL")

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
				fmt.Println("Check domain : " + getMySQLnamespace(namespace) + " : available schema : " + v.(string))
				if v.(string) == getMySQLnamespace(namespace) {
					//Database available
					isDatabaseAvailbale = true
					break
				}
			}
		}
	}

	if isDatabaseAvailbale {
		fmt.Println("Database already available. Nothing to do. Proceed!")
	} else {
		_, err = session.Query("create schema " + getMySQLnamespace(namespace) + ";")
		if err != nil {
			fmt.Println("Creation of domain matched Schema failed")
		} else {
			fmt.Println("Creation of domain matched Schema Successful")
		}
	}

	fmt.Println("Reusing existing MySQL connection")
	return
}

func createMySQLTable(namespace string, class string, session *sql.DB, array []map[string]interface{}) (status bool) {
	status = false

	//get table list
	classBytes := executeMySqlGetClasses(session, namespace)
	var classList []string
	err := json.Unmarshal(classBytes, &classList)
	if err != nil {
		status = false
	} else {
		for _, className := range classList {
			if strings.ToLower(class) == className {
				fmt.Println("Table Already Available")
				status = true
				//Get all fields
				classBytes := executeMySqlGetFields(session, namespace, class)
				var tableFieldList []string
				_ = json.Unmarshal(classBytes, &tableFieldList)
				//Check For missing fields. If any ALTER TABLE
				var recordFieldList []string
				var recordFieldType []string

				recordFieldList = make([]string, len(array[0]))
				recordFieldType = make([]string, len(array[0]))
				index := 0
				for key, value := range array[0] {
					if key == "__osHeaders" {
						recordFieldList[index] = "osHeaders"
						recordFieldType[index] = "text"
					} else {
						recordFieldList[index] = key
						recordFieldType[index] = getMySqlDataType(value)
					}
					index++
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
					_, er := session.Query("ALTER TABLE " + getMySQLnamespace(namespace) + "." + strings.ToLower(class) + " ADD COLUMN " + newFields[key] + " " + newTypes[key] + ";")
					if er != nil {
						status = false
						fmt.Println("Table Alter Failed : " + er.Error())
						return
					} else {
						status = true
						fmt.Println("Table Alter Success!")
					}
				}

				return
			}
		}

		// if not available
		//get one object
		var dataObject map[string]interface{}
		dataObject = make(map[string]interface{})

		for key, value := range array[0] {
			if key == "__osHeaders" {
				dataObject["osHeaders"] = value
			} else {
				dataObject[key] = value
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
				if keyArray[i] == "OriginalIndex" {
					argKeyList2 = argKeyList2 + keyArray[i] + " VARCHAR(255) PRIMARY KEY, "
				} else {
					argKeyList2 = argKeyList2 + keyArray[i] + " " + dataTypeArray[i] + ", "
				}

			} else {
				if keyArray[i] == "OriginalIndex" {
					argKeyList2 = argKeyList2 + keyArray[i] + " VARCHAR(255) PRIMARY KEY"
				} else {
					argKeyList2 = argKeyList2 + keyArray[i] + " " + dataTypeArray[i]
				}

			}
		}

		//request.Log("create table " + getMySQLnamespace(namespace) + "." + strings.ToLower(request.Controls.Class) + "(" + argKeyList2 + ");")

		_, er := session.Query("create table " + getMySQLnamespace(namespace) + "." + strings.ToLower(class) + "(" + argKeyList2 + ");")
		if er != nil {
			status = false
			fmt.Println("Table Creation Failed : " + er.Error())
			return
		}

		status = true

	}

	return
}

type ControlHeaders struct {
	Version    string
	Namespace  string
	Class      string
	Tenant     string
	LastUdated string
}

func ConvertOsheaders(input ControlHeaders) string {
	myStr := "{\"Class\":\"" + input.Class + "\",\"LastUdated\":\"2" + input.LastUdated + "\",\"Namespace\":\"" + input.Namespace + "\",\"Tenant\":\"" + input.Tenant + "\",\"Version\":\"" + input.Version + "\"}"
	return myStr
}

func getStringByObject(obj interface{}) string {

	result, err := json.Marshal(obj)

	if err == nil {
		return string(result)
	} else {
		return "{}"
	}
}

func getMySqlDataType(item interface{}) (datatype string) {
	datatype = reflect.TypeOf(item).Name()
	if datatype == "bool" {
		datatype = "text"
	} else if datatype == "float64" {
		datatype = "real"
	} else if datatype == "" || datatype == "string" || datatype == "ControlHeaders" {
		datatype = "text"
	}
	return datatype
}

func executeMySqlGetClasses(session *sql.DB, Namespace string) (returnByte []byte) {

	namespace := getMySQLnamespace(Namespace)

	var returnMap map[string]interface{}
	returnMap = make(map[string]interface{})

	rows, err := session.Query("SELECT DISTINCT TABLE_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='" + namespace + "';")

	if err != nil {
		fmt.Println("Error executing query in MySQL")
	} else {
		fmt.Println("Successfully executed query in MySQL")
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
			fmt.Println("Error converting to byte array")
			byteValue = nil
		} else {
			fmt.Println("Successfully converted result to byte array")
		}

		returnByte = byteValue
	}

	return
}

func executeMySqlGetFields(session *sql.DB, Namespace string, class string) (returnByte []byte) {

	namespace := getMySQLnamespace(Namespace)

	var returnMap map[string]interface{}
	returnMap = make(map[string]interface{})

	rows, err := session.Query("describe " + namespace + "." + class)

	if err != nil {
		fmt.Println("Error executing query in MySQL")
	} else {
		fmt.Println("Successfully executed query in MySQL")
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
			fmt.Println("Error converting to byte array")
			byteValue = nil
		} else {
			fmt.Println("Successfully converted result to byte array")
		}

		returnByte = byteValue
	}

	return returnByte
}
