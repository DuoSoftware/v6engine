package processes

import (
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"fmt"
	"github.com/mattbaird/elastigo/lib"
	"github.com/twinj/uuid"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func RestoreServer(ipAddress string) (status bool) {
	status = true
	for _, value := range GetBackupFileList() {
		content, _ := ioutil.ReadFile(value)
		var array []map[string]interface{}
		_ = json.Unmarshal(content, &array)
		namespace, class := getNamespaceAndClass(value)
		status = InsertElastic(ipAddress, namespace, class, array)
	}
	return
}

func GetBackupFileList() []string {
	files1, _ := filepath.Glob("*.objectfile")
	return files1
}

/*func InsertElastic(ipAddress string, namespace string, class string, array []map[string]interface{}) (status bool) {
	var Request messaging.ObjectRequest
	Request.Controls = messaging.RequestControls{SecurityToken: "token", Namespace: namespace, Class: class}
	Request.Body.Parameters = messaging.ObjectParameters{}
	Request.Body.Parameters.KeyProperty = "OriginalIndex"
	Request.Controls.Multiplicity = "multiple"
	Request.Body.Objects = array
	Request.IsLogEnabled = true
	var initialSlice []string
	initialSlice = make([]string, 0)
	Request.MessageStack = initialSlice
	request := &Request
	setManyElastic(ipAddress, request)
	status = true
	return
}
*/
func getNoSqlKeyByGUID(request *messaging.ObjectRequest) string {
	namespace := request.Controls.Namespace
	class := request.Controls.Class
	key := namespace + "." + class + "." + uuid.NewV1().String()
	return key
}

func getNoSqlKeyById(request *messaging.ObjectRequest, obj map[string]interface{}) string {
	key := request.Controls.Namespace + "." + request.Controls.Class + "." + obj[request.Body.Parameters.KeyProperty].(string)
	return key
}

func setManyElastic(url string, request *messaging.ObjectRequest) {
	tokens := strings.Split(url, ":")
	conn := getElasticConnection()(tokens[0], tokens[1])
	request.Log("Restoring " + strconv.Itoa(len(request.Body.Objects)) + " records TO Namespace : " + request.Controls.Namespace + " Class : " + request.Controls.Class)
	isGUIDKey := false
	isAutoIncrementKey := false
	currentIndex := 0
	maxCount := ""
	CountIndex := 0

	if (request.Body.Objects[0][request.Body.Parameters.KeyProperty].(string) == "-888") || (request.Body.Parameters.GUIDKey == true) {
		request.Log("GUID Key generation Requested")
		isGUIDKey = true
	} else if (request.Body.Objects[0][request.Body.Parameters.KeyProperty].(string) == "-999") || (request.Body.Parameters.AutoIncrement == true) {
		request.Log("Auto-Increment Key generation Requested")
		isAutoIncrementKey = true

		//set starting index
		//Read maxCount from domainClassAttributes table
		request.Log("Reading the max count")
		classkey := request.Controls.Class
		data, err := conn.Get(request.Controls.Namespace, "domainClassAttributes", classkey, nil)

		if err != nil {
			request.Log("No record Found. This is a NEW record. Inserting new attribute value")
			var newRecord map[string]interface{}
			newRecord = make(map[string]interface{})
			newRecord["class"] = request.Controls.Class
			newRecord["maxCount"] = "0"
			newRecord["version"] = uuid.NewV1().String()

			_, err = conn.Index(request.Controls.Namespace, "domainClassAttributes", classkey, nil, newRecord)

			if err != nil {
				errorMessage := "Failed to create new Domain Class Attribute entry."
				request.Log(errorMessage)
			} else {
				request.Log("Successfully new Domain Class Attribute to elastic search")
				maxCount = "0"
			}

		} else {
			request.Log("Successfully retrieved object from Elastic Search")
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})

			byteData, err := data.Source.MarshalJSON()

			if err != nil {
				request.Log("Data serialization to read maxCount failed")
			}

			json.Unmarshal(byteData, &currentMap)
			maxCount = currentMap["maxCount"].(string)
			intTemp, _ := strconv.Atoi(maxCount)
			currentIndex = intTemp
		}

	} else {
		//	request.Log("Manual Keys supplied!")
	}

	stub := 100

	noOfSets := (len(request.Body.Objects) / stub)
	remainderFromSets := 0
	statusCount := noOfSets
	remainderFromSets = (len(request.Body.Objects) - (noOfSets * stub))
	if remainderFromSets > 0 {
		statusCount++
	}
	var setStatus []bool
	setStatus = make([]bool, statusCount)

	startIndex := 0
	stopIndex := stub
	statusIndex := 0

	waitTime := 2000

	for x := 0; x < noOfSets; x++ {
		fmt.Println("Inserting data stub : " + strconv.Itoa(x))
		indexer := conn.NewBulkIndexer(stub)
		//nowTime := time.Now()

		if isAutoIncrementKey {
			//Read maxCount from domainClassAttributes table
			request.Log("Reading the max count")
			classkey := request.Controls.Class
			data, err := conn.Get(request.Controls.Namespace, "domainClassAttributes", classkey, nil)

			if err != nil {
				request.Log("No record Found. This is a NEW record. Inserting new attribute value")
				var newRecord map[string]interface{}
				newRecord = make(map[string]interface{})
				newRecord["class"] = request.Controls.Class
				newRecord["maxCount"] = "0"
				newRecord["version"] = uuid.NewV1().String()

				_, err = conn.Index(request.Controls.Namespace, "domainClassAttributes", classkey, nil, newRecord)

				if err != nil {
					errorMessage := "Failed to create new Domain Class Attribute entry."
					request.Log(errorMessage)
				} else {
					maxCount = "0"
				}

			} else {
				request.Log("Successfully retrieved object from Elastic Search")
				var currentMap map[string]interface{}
				currentMap = make(map[string]interface{})

				byteData, err := data.Source.MarshalJSON()

				if err != nil {
					request.Log("Data serialization to read maxCount failed")
					//response.Message = "Data serialization to read maxCount failed"
					//return response
				}

				json.Unmarshal(byteData, &currentMap)
				maxCount = currentMap["maxCount"].(string)
			}

			//Increment by 100 and update
			tempCount, err := strconv.Atoi(maxCount)
			maxCount = strconv.Itoa(tempCount + stub)

			request.Log("Updating Domain Class Attribute table")
			var newRecord map[string]interface{}
			newRecord = make(map[string]interface{})
			newRecord["class"] = request.Controls.Class
			newRecord["maxCount"] = maxCount
			newRecord["version"] = uuid.NewV1().String()
			_, err2 := conn.Index(request.Controls.Namespace, "domainClassAttributes", request.Controls.Class, nil, newRecord)
			if err2 != nil {
				request.Log("Inserting to Elastic Failed")
			} else {
				request.Log("Inserting to Elastic Successfull")
			}
		}

		for _, obj := range request.Body.Objects[startIndex:stopIndex] {
			nosqlid := ""
			if isGUIDKey {
				nosqlid = getNoSqlKeyByGUID(request)
				itemArray := strings.Split(nosqlid, (request.Controls.Namespace + "." + request.Controls.Class + "."))
				obj[request.Body.Parameters.KeyProperty] = itemArray[1]
			} else if isAutoIncrementKey {
				currentIndex += 1
				nosqlid = request.Controls.Namespace + "." + request.Controls.Class + "." + strconv.Itoa(currentIndex)
				obj[request.Body.Parameters.KeyProperty] = strconv.Itoa(currentIndex)
			} else {
				nosqlid = getNoSqlKeyById(request, obj)
			}
			CountIndex++
			delete(obj, "OriginalIndex")
			//indexer.Index(request.Controls.Namespace, request.Controls.Class, nosqlid, "10", &nowTime, obj, false)
		}
		indexer.Start()
		numerrors := indexer.NumErrors()
		time.Sleep(time.Duration(waitTime) * time.Millisecond)

		if numerrors != 0 {
			request.Log("Elastic Search bulk insert error")
			setStatus[statusIndex] = false
		} else {
			setStatus[statusIndex] = true
		}
		statusIndex++
		startIndex += stub
		stopIndex += stub
	}

	if remainderFromSets > 0 {
		fmt.Println("Inserting Last Stub of record Set!")
		start := len(request.Body.Objects) - remainderFromSets
		indexer := conn.NewBulkIndexer(stub)
		//nowTime := time.Now()

		if isAutoIncrementKey {
			//Read maxCount from domainClassAttributes table
			request.Log("Reading the max count")
			classkey := request.Controls.Class
			data, err := conn.Get(request.Controls.Namespace, "domainClassAttributes", classkey, nil)

			if err != nil {
				request.Log("No record Found. This is a NEW record. Inserting new attribute value")
				var newRecord map[string]interface{}
				newRecord = make(map[string]interface{})
				newRecord["class"] = request.Controls.Class
				newRecord["maxCount"] = "0"
				newRecord["version"] = uuid.NewV1().String()

				_, err = conn.Index(request.Controls.Namespace, "domainClassAttributes", classkey, nil, newRecord)

				if err != nil {
					errorMessage := "Failed to create new Domain Class Attribute entry."
					request.Log(errorMessage)
				} else {
					maxCount = "0"
				}

			} else {
				request.Log("Successfully retrieved object from Elastic Search")
				var currentMap map[string]interface{}
				currentMap = make(map[string]interface{})

				byteData, err := data.Source.MarshalJSON()

				if err != nil {
					request.Log("Data serialization to read maxCount failed")
				}

				json.Unmarshal(byteData, &currentMap)
				maxCount = currentMap["maxCount"].(string)
			}

			//Increment by 100 and update
			tempCount, err := strconv.Atoi(maxCount)
			maxCount = strconv.Itoa(tempCount + (len(request.Body.Objects) - startIndex))

			request.Log("Updating Domain Class Attribute table")
			var newRecord map[string]interface{}
			newRecord = make(map[string]interface{})
			newRecord["class"] = request.Controls.Class
			newRecord["maxCount"] = maxCount
			newRecord["version"] = uuid.NewV1().String()
			_, err2 := conn.Index(request.Controls.Namespace, "domainClassAttributes", request.Controls.Class, nil, newRecord)
			if err2 != nil {
				request.Log("Inserting to Elastic Failed")
			} else {
				request.Log("Inserting to Elastic Successfull")
			}
		}

		for _, obj := range request.Body.Objects[start:len(request.Body.Objects)] {
			nosqlid := ""
			if isGUIDKey {
				request.Log("GUIDKey keys requested")
				nosqlid = getNoSqlKeyByGUID(request)
				itemArray := strings.Split(nosqlid, (request.Controls.Namespace + "." + request.Controls.Class + "."))
				obj[request.Body.Parameters.KeyProperty] = itemArray[1]
			} else if isAutoIncrementKey {
				currentIndex += 1
				nosqlid = request.Controls.Namespace + "." + request.Controls.Class + "." + strconv.Itoa(currentIndex)
				obj[request.Body.Parameters.KeyProperty] = strconv.Itoa(currentIndex)
			} else {
				nosqlid = getNoSqlKeyById(request, obj)
			}
			CountIndex++
			delete(obj, "OriginalIndex")
			//indexer.Index(request.Controls.Namespace, request.Controls.Class, nosqlid, "1000", &nowTime, obj, false)
		}

		fmt.Println("-----------------")
		fmt.Println(len(request.Body.Objects))
		indexer.Start()
		numerrors := indexer.NumErrors()
		time.Sleep(time.Duration(waitTime) * time.Millisecond)

		if numerrors != 0 {
			request.Log("Elastic Search bulk insert error")
			setStatus[statusIndex] = false
		} else {
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
		request.Log("Hurrah! Done!")
	} else {
		request.Log("nah :( Some Isssue! ")
	}
}

func getElasticConnection() func(host string, port string) *elastigo.Conn {

	var connection *elastigo.Conn

	return func(host string, port string) *elastigo.Conn {
		if connection == nil {
			fmt.Println("Establishing new connection for Elastic Search " + host + ":" + port)
			conn := elastigo.NewConn()
			conn.SetHosts([]string{host})
			conn.Port = port
			connection = conn
		}

		fmt.Println("Reusing existing Elastic Search connection ")
		return connection
	}
}

func getRecordID(inputMap map[string]interface{}) (key string) {
	var keyproperty string

	for key, value := range inputMap {
		if key == "OriginalIndex" {
			keyproperty = value.(string)
		}
	}
	key = keyproperty
	return
}

func getNamespaceAndClass(fileName string) (namespace string, class string) {
	nameWithoutExt := strings.Replace(fileName, ".objectfile", "", -1)
	tokens := strings.Split(nameWithoutExt, "-")
	namespace = tokens[0]
	class = tokens[1]
	return
}

//OLDER INSERT METHOD... DONT DELETE!!!!!!!!!!!!!!!!!!!!!!!!!
//SATAN SHALL HUNT YOU AND MAKE YOU HIS PREY IF YOU DELETE THIS!!!!!!!!!!!!!!!

func InsertElastic(ipAddress string, namespace string, class string, array []map[string]interface{}) (status bool) {
	conn := getConnection(ipAddress)
	status = true
	fmt.Println("Restoring Namespace : " + namespace + " Class : " + class)
	for _, obj := range array {
		nosqlid := getRecordID(obj)

		var allMaps map[string]interface{}
		allMaps = make(map[string]interface{})

		for key, value := range obj {

			if key != "OriginalIndex" {
				allMaps[key] = value
			} else {
				//do nothing
			}
		}

		_, err := conn.Index(namespace, class, nosqlid, nil, allMaps)

		if err != nil {
			status = false
		} else {
			status = true
		}
	}
	return
}
