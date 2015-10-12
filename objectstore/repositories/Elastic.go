package repositories

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/queryparser"
	"encoding/json"
	"fmt"
	"github.com/mattbaird/elastigo/lib"
	"github.com/twinj/uuid"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ElasticRepository struct {
}

func (repository ElasticRepository) GetRepositoryName() string {
	return "Elastic Search"
}

func (repository ElasticRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GETALL")
	request.Log(request.Controls.SendMetaData)
	return search(request, "*")
}

func (repository ElasticRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GETSEARCH")
	fmt.Println("Params : " + request.Body.Query.Parameters)
	return search(request, request.Body.Query.Parameters)
}

func search(request *messaging.ObjectRequest, searchStr string) RepositoryResponse {
	response := RepositoryResponse{}
	conn := getConnection()(request)
	fmt.Println(searchStr)
	skip := "0"

	if request.Extras["skip"] != nil {
		skip = request.Extras["skip"].(string)
	}

	take := "100000"

	if request.Extras["take"] != nil {
		take = request.Extras["take"].(string)
	}
	query := "{\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + searchStr + "\"}}}"
	//query := "{\"sort\" : [{\"__osHeaders.LastUdated\" : {\"order\" : \"desc\"}}],\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + searchStr + "\"}}}"
	//query := "{\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + searchStr + "\"}},\"sort\" : [{\"__osHeaders.LastUdated\" : {\"order\" : \"desc\"}}]}"

	data, err := conn.Search(request.Controls.Namespace, request.Controls.Class, nil, query)

	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			var emptymap map[string]interface{}
			emptymap = make(map[string]interface{})
			finalBytes, _ := json.Marshal(emptymap)
			response.GetResponseWithBody(finalBytes)
		} else {
			errorMessage := "Error retrieving object from elastic search : " + err.Error()
			request.Log(errorMessage)
			request.Log("Error Query : " + query)
			response.GetErrorResponse(errorMessage)
		}
	} else {
		request.Log("Successfully retrieved object from Elastic Search")

		var allMaps []map[string]interface{}
		allMaps = make([]map[string]interface{}, data.Hits.Len())
		for index, hit := range data.Hits.Hits {
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})

			byteData, _ := hit.Source.MarshalJSON()

			json.Unmarshal(byteData, &currentMap)

			//Check if meta data is not needed
			if request.Controls.SendMetaData == "false" {
				delete(currentMap, "__osHeaders")
			}

			allMaps[index] = currentMap
		}

		finalBytes, _ := json.Marshal(allMaps)
		response.GetResponseWithBody(finalBytes)
	}

	return response
}

func (repository ElasticRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-QUERY!")
	response := RepositoryResponse{}
	queryType := request.Body.Query.Type

	switch queryType {
	case "Query":
		if request.Body.Query.Parameters != "*" {
			fieldsInByte := executeElasticQuery(request)
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
			//Check if just STAR then execute GET-SEARCH method
			request.Log("Redirecting to GET-SEARCH!")
			return search(request, request.Body.Query.Parameters)
		}
	default:
		return search(request, request.Body.Query.Parameters)

	}

	return response
}

func (repository ElasticRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-BY-KEY")
	response := RepositoryResponse{}
	conn := getConnection()(request)

	key := getNoSqlKey(request)
	request.Log("Elastic Search Get By Key : " + key)
	data, err := conn.Get(request.Controls.Namespace, request.Controls.Class, key, nil)

	if err != nil {
		// errorMessage := "Error retrieving object from elastic search : " + err.Error()
		// request.Log(errorMessage)
		// response.GetErrorResponse(errorMessage)
		if strings.Contains(err.Error(), "record not found") {
			var emptymap map[string]interface{}
			emptymap = make(map[string]interface{})
			finalBytes, _ := json.Marshal(emptymap)
			response.GetResponseWithBody(finalBytes)
		}
	} else {
		request.Log("Successfully retrieved object from Elastic Search")
		bytes, err := data.Source.MarshalJSON()

		//Get Data to struct
		var originalData map[string]interface{}
		originalData = make(map[string]interface{})
		json.Unmarshal(bytes, &originalData)

		//Check if meta data is not needed
		if request.Controls.SendMetaData == "false" {
			delete(originalData, "__osHeaders")
			bytes, _ = json.Marshal(originalData)
		}

		if err != nil {
			errorMessage := "Elastic search JSON marshal error : " + err.Error()
			request.Log(errorMessage)
			response.GetErrorResponse(errorMessage)

		} else {
			response.GetResponseWithBody(bytes)
		}

	}

	return response
}

func (repository ElasticRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-MULTIPLE")
	return setManyElastic(request)
}

func (repository ElasticRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-SINGLE")
	return setOneElastic(request)
}

func setOneElastic(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	conn := getConnection()(request)

	isAutoIncrementing := false
	isRandomKeyID := false
	returnID := ""

	if (request.Controls.Id == "-999") || (request.Body.Parameters.AutoIncrement == true) {
		isAutoIncrementing = true
	}

	if (request.Controls.Id == "-888") || (request.Body.Parameters.GUIDKey == true) {
		isRandomKeyID = true
	}

	if isRandomKeyID {
		request.Log("GUID Key Selected!")
		key := getNoSqlKeyByGUID(request)
		itemArray := strings.Split(key, (request.Controls.Namespace + "." + request.Controls.Class + "."))
		request.Body.Object[request.Body.Parameters.KeyProperty] = itemArray[1]
		returnID = itemArray[1]
		_, err := conn.Index(request.Controls.Namespace, request.Controls.Class, key, nil, request.Body.Object)

		if err != nil {
			errorMessage := "Elastic Search Single Insert Error : " + err.Error()
			request.Log(errorMessage)
			response.GetErrorResponse(errorMessage)
		} else {
			response.IsSuccess = true
			request.Log(response.Message)
			response.Message = "Successfully inserted one to elastic search"
			response.Message = key
		}
	} else if isAutoIncrementing {
		//Read maxCount from domainClassAttributes table
		request.Log("Automatic Increment Key Selected!")
		key := request.Controls.Class
		data, err := conn.Get(request.Controls.Namespace, "domainClassAttributes", key, nil)
		maxCount := ""

		if err != nil {
			request.Log("No record Found. This is a NEW record. Inserting new attribute value")
			var newRecord map[string]interface{}
			newRecord = make(map[string]interface{})
			newRecord["class"] = request.Controls.Class
			newRecord["maxCount"] = "0"
			newRecord["version"] = uuid.NewV1().String()

			_, err = conn.Index(request.Controls.Namespace, "domainClassAttributes", key, nil, newRecord)

			if err != nil {
				errorMessage := "Elastic Search Insert Error : " + err.Error()
				request.Log(errorMessage)
				return response
			} else {
				response.IsSuccess = true
				response.Message = "Successfully inserted one to elastic search"
				request.Log(response.Message)
				maxCount = "0"
			}

		} else {
			request.Log("Successfully retrieved object from Elastic Search")
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})

			byteData, err := data.Source.MarshalJSON()

			if err != nil {
				request.Log("Data serialization to read maxCount failed")
				response.Message = "Data serialization to read maxCount failed"
				return response
			}

			json.Unmarshal(byteData, &currentMap)
			maxCount = currentMap["maxCount"].(string)
		}

		tempCount, err := strconv.Atoi(maxCount)
		maxCount = strconv.Itoa(tempCount + 1)

		request.Log("Inserting Actual data Body")
		recordKey := request.Controls.Namespace + "." + request.Controls.Class + "." + maxCount
		request.Body.Object[request.Body.Parameters.KeyProperty] = maxCount
		returnID = maxCount
		_, err1 := conn.Index(request.Controls.Namespace, request.Controls.Class, recordKey, nil, request.Body.Object)

		//Update the Count
		var newRecord map[string]interface{}
		newRecord = make(map[string]interface{})
		newRecord["class"] = request.Controls.Class
		newRecord["maxCount"] = maxCount
		newRecord["version"] = uuid.NewV1().String()
		_, err2 := conn.Index(request.Controls.Namespace, "domainClassAttributes", request.Controls.Class, nil, newRecord)
		if err1 != nil || err2 != nil {
			request.Log("Inserting to Elastic Failed")
			response.Message = "Inserting to Elastic Failed"
			return response
		} else {
			request.Log("Inserting to Elastic Successfull")
			response.Message = "Inserting to Elastic Successfull"
			response.IsSuccess = true
			request.Log(response.Message)
			response.Message = recordKey
		}

	} else {
		request.Log("Manual ID Selected!")
		key := getNoSqlKey(request)
		request.Log("Inserting single object to Elastic Search : " + key)
		returnID = request.Controls.Id
		_, err := conn.Index(request.Controls.Namespace, request.Controls.Class, key, nil, request.Body.Object)

		if err != nil {
			errorMessage := "Elastic Search Single Insert Error : " + err.Error()
			request.Log(errorMessage)
			response.GetErrorResponse(errorMessage)
			return response
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted one to elastic search"
			request.Log(response.Message)
			response.Message = key
		}
	}

	//Update Response
	var Data []map[string]interface{}
	Data = make([]map[string]interface{}, 1)
	var actualData map[string]interface{}
	actualData = make(map[string]interface{})
	actualData["ID"] = returnID
	Data[0] = actualData
	response.Data = Data
	return response
}

func setManyElastic(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	conn := getConnection()(request)

	request.Log("Starting Elastic Search bulk insert")

	request.Log("Determining which Key indexing method is used")

	isGUIDKey := false
	isAutoIncrementKey := false
	currentIndex := 0
	maxCount := ""
	CountIndex := 0

	var Data map[string]interface{}
	Data = make(map[string]interface{})

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
				return response
			} else {
				response.IsSuccess = true
				response.Message = "Successfully new Domain Class Attribute to elastic search"
				request.Log(response.Message)
				maxCount = "0"
			}

		} else {
			request.Log("Successfully retrieved object from Elastic Search")
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})

			byteData, err := data.Source.MarshalJSON()

			if err != nil {
				request.Log("Data serialization to read maxCount failed")
				response.Message = "Data serialization to read maxCount failed"
				return response
			}

			json.Unmarshal(byteData, &currentMap)
			maxCount = currentMap["maxCount"].(string)
			intTemp, _ := strconv.Atoi(maxCount)
			currentIndex = intTemp
		}

	} else {
		request.Log("Manual Keys supplied!")
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

	//numberOfSets := (len(request.Body.Objects) / 100) + 1
	startIndex := 0
	stopIndex := stub
	statusIndex := 0

	for x := 0; x < noOfSets; x++ {

		indexer := conn.NewBulkIndexer(stub)
		nowTime := time.Now()

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
					return response
				} else {
					response.IsSuccess = true
					response.Message = "Successfully new Domain Class Attribute to elastic search"
					request.Log(response.Message)
					maxCount = "0"
				}

			} else {
				request.Log("Successfully retrieved object from Elastic Search")
				var currentMap map[string]interface{}
				currentMap = make(map[string]interface{})

				byteData, err := data.Source.MarshalJSON()

				if err != nil {
					request.Log("Data serialization to read maxCount failed")
					response.Message = "Data serialization to read maxCount failed"
					return response
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
				response.Message = "Inserting to Elastic Failed"
				return response
			} else {
				request.Log("Inserting to Elastic Successfull")
				response.Message = "Inserting to Elastic Successfull"
				response.IsSuccess = true
				request.Log(response.Message)
			}
		}

		for _, obj := range request.Body.Objects[startIndex:stopIndex] {
			nosqlid := ""
			temp := ""
			if isGUIDKey {
				nosqlid = getNoSqlKeyByGUID(request)
				itemArray := strings.Split(nosqlid, (request.Controls.Namespace + "." + request.Controls.Class + "."))
				obj[request.Body.Parameters.KeyProperty] = itemArray[1]
				temp = itemArray[1]
			} else if isAutoIncrementKey {
				currentIndex += 1
				nosqlid = request.Controls.Namespace + "." + request.Controls.Class + "." + strconv.Itoa(currentIndex)
				obj[request.Body.Parameters.KeyProperty] = strconv.Itoa(currentIndex)
				temp = strconv.Itoa(currentIndex)
			} else {
				nosqlid = getNoSqlKeyById(request, obj)
				temp = obj[request.Body.Parameters.KeyProperty].(string)
			}
			fmt.Println(temp)
			CountIndex++
			Data[strconv.Itoa(CountIndex)] = temp
			indexer.Index(request.Controls.Namespace, request.Controls.Class, nosqlid, "10", &nowTime, obj, false)
		}
		indexer.Start()
		numerrors := indexer.NumErrors()
		time.Sleep(1200 * time.Millisecond)

		if numerrors != 0 {
			request.Log("Elastic Search bulk insert error")
			response.GetErrorResponse("Elastic Search bulk insert error")
			setStatus[statusIndex] = false
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted bulk to Elastic Search"
			request.Log(response.Message)
			setStatus[statusIndex] = true
		}
		statusIndex++
		startIndex += stub
		stopIndex += stub
	}

	if remainderFromSets > 0 {
		start := len(request.Body.Objects) - remainderFromSets
		indexer := conn.NewBulkIndexer(stub)
		nowTime := time.Now()

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
					return response
				} else {
					response.IsSuccess = true
					response.Message = "Successfully new Domain Class Attribute to elastic search"
					request.Log(response.Message)
					maxCount = "0"
				}

			} else {
				request.Log("Successfully retrieved object from Elastic Search")
				var currentMap map[string]interface{}
				currentMap = make(map[string]interface{})

				byteData, err := data.Source.MarshalJSON()

				if err != nil {
					request.Log("Data serialization to read maxCount failed")
					response.Message = "Data serialization to read maxCount failed"
					return response
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
				response.Message = "Inserting to Elastic Failed"
				return response
			} else {
				request.Log("Inserting to Elastic Successfull")
				response.Message = "Inserting to Elastic Successfull"
				response.IsSuccess = true
				request.Log(response.Message)
			}
		}

		for _, obj := range request.Body.Objects[start:len(request.Body.Objects)] {
			temp := ""
			nosqlid := ""
			if isGUIDKey {
				request.Log("GUIDKey keys requested")
				nosqlid = getNoSqlKeyByGUID(request)
				itemArray := strings.Split(nosqlid, (request.Controls.Namespace + "." + request.Controls.Class + "."))
				obj[request.Body.Parameters.KeyProperty] = itemArray[1]
				temp = itemArray[1]
			} else if isAutoIncrementKey {
				currentIndex += 1
				nosqlid = request.Controls.Namespace + "." + request.Controls.Class + "." + strconv.Itoa(currentIndex)
				obj[request.Body.Parameters.KeyProperty] = strconv.Itoa(currentIndex)
				temp = strconv.Itoa(currentIndex)
			} else {
				nosqlid = getNoSqlKeyById(request, obj)
				temp = obj[request.Body.Parameters.KeyProperty].(string)
			}
			fmt.Println(temp)
			CountIndex++
			Data[strconv.Itoa(CountIndex)] = temp
			indexer.Index(request.Controls.Namespace, request.Controls.Class, nosqlid, "10", &nowTime, obj, false)
		}
		indexer.Start()
		numerrors := indexer.NumErrors()
		time.Sleep(1200 * time.Millisecond)

		if numerrors != 0 {
			request.Log("Elastic Search bulk insert error")
			response.GetErrorResponse("Elastic Search bulk insert error")
			setStatus[statusIndex] = false
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted bulk to Elastic Search"
			request.Log(response.Message)
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
		response.Message = "Successfully inserted many objects in to Elastic"
		request.Log(response.Message)
	} else {
		response.IsSuccess = false
		response.Message = "Error inserting many objects in to Elastic"
		request.Log(response.Message)
		response.GetErrorResponse("Error inserting many objects in to Elastic")
	}

	//Update Response
	var DataMap []map[string]interface{}
	DataMap = make([]map[string]interface{}, 1)
	var actualInput map[string]interface{}
	actualInput = make(map[string]interface{})
	actualInput["ID"] = Data
	DataMap[0] = actualInput
	response.Data = DataMap
	return response
}

func (repository ElasticRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-MULTIPLE")
	return setManyElastic(request)
}

func (repository ElasticRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-SINGLE")
	return setOneElastic(request)
}

func (repository ElasticRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-MULTIPLE")
	response := RepositoryResponse{}

	conn := getConnection()(request)

	for _, object := range request.Body.Objects {
		key := getNoSqlKeyById(request, object)
		request.Log("Deleting single object from Elastic Search : " + key)
		_, err := conn.Delete(request.Controls.Namespace, request.Controls.Class, key, nil)
		if err != nil {
			errorMessage := "Elastic Search single delete error : " + err.Error()
			request.Log(errorMessage)
			response.GetErrorResponse(errorMessage)
		} else {
			response.IsSuccess = true
			response.Message = "Successfully deleted one in elastic search"
			request.Log(response.Message)
		}
	}

	return response

}

func (repository ElasticRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-SINGLE")
	response := RepositoryResponse{}

	conn := getConnection()(request)
	key := getNoSqlKey(request)
	request.Log("Deleting single object from Elastic Search : " + key)
	_, err := conn.Delete(request.Controls.Namespace, request.Controls.Class, getNoSqlKey(request), nil)
	if err != nil {
		errorMessage := "Elastic Search single delete error : " + err.Error()
		request.Log(errorMessage)
		response.GetErrorResponse(errorMessage)
	} else {
		response.IsSuccess = true
		response.Message = "Successfully deleted one in elastic search"
		request.Log(response.Message)
	}

	return response
}

func (repository ElasticRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	request.Log("Starting SPECIAL!")
	queryType := request.Body.Special.Type

	switch queryType {
	case "getFields":
		request.Log("Starting GET-FIELDS sub routine!")
		fieldsInByte := executeElasticGetFields(request)
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
		fieldsInByte := executeElasticGetClasses(request)
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
		fieldsInByte := executeElasticGetNamespaces(request)
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
		request.Log("Starting GET-SELECTED_FIELDS sub routine")
		fieldsInByte := executeElasticGetSelectedFields(request)
		if fieldsInByte != nil {
			response.IsSuccess = true
			response.Message = "Successfully Retrieved All selected Field data"
			response.GetResponseWithBody(fieldsInByte)
		} else {
			response.IsSuccess = false
			response.Message = "Aborted! Unsuccessful Retrieving All selected field data"
			errorMessage := response.Message
			response.GetErrorResponse(errorMessage)
		}
	default:
		return search(request, request.Body.Special.Parameters)

	}

	return response

}

func (repository ElasticRepository) Test(request *messaging.ObjectRequest) {

}

//SUB FUNCTIONS
//Functions from SPECIAL and QUERY

func executeElasticQuery(request *messaging.ObjectRequest) (returnByte []byte) {
	conn := getConnection()(request)
	skip := "0"

	if request.Extras["skip"] != nil {
		skip = request.Extras["skip"].(string)
	}

	take := "100000"

	if request.Extras["take"] != nil {
		take = request.Extras["take"].(string)
	}

	searchStr, isSelectedFields, selectedFields, fromClass := queryparser.GetQuery(request.Body.Query.Parameters)
	request.Log("Search String : " + searchStr)
	fmt.Print("Selcted Fields ? : ")
	fmt.Println(isSelectedFields)
	fmt.Print("Selected Fields List : ")
	fmt.Println(selectedFields)
	request.Log("Class if mentioned : " + fromClass)

	//searchStr := "productName : STB001 AND Namespace : com.duosoftware.com  "
	//searchStr := "NOT Id : 700"
	//searchStr = "Id:>=700"
	//query := "{\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + "*" + "\"}}}"
	query := "{\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + searchStr + "\"}}}"

	var data elastigo.SearchResult
	var err error
	fmt.Println(request.Controls.Class)
	if fromClass == "" {
		data, err = conn.Search(request.Controls.Namespace, request.Controls.Class, nil, query)
	} else {
		data, err = conn.Search(request.Controls.Namespace, fromClass, nil, query)
	}
	if err != nil {
		errorMessage := "Error retrieving object from elastic search : " + err.Error()
		request.Log(errorMessage)
		request.Log("Error Query : " + query)
	} else {
		request.Log("Successfully retrieved object from Elastic Search")

		var allMaps []map[string]interface{}
		allMaps = make([]map[string]interface{}, data.Hits.Len())

		for index, hit := range data.Hits.Hits {
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})

			byteData, _ := hit.Source.MarshalJSON()

			json.Unmarshal(byteData, &currentMap)

			allMaps[index] = currentMap
		}

		for _, value := range selectedFields {
			if value == "*" {
				isSelectedFields = false
			}
		}

		if isSelectedFields {
			var fields []string
			fields = selectedFields

			fmt.Print("Fields List : ")
			fmt.Println(fields)

			//create map to store data

			var outMap []map[string]interface{}
			outMap = make([]map[string]interface{}, len(allMaps))

			for index, value := range allMaps {

				var currentMap map[string]interface{}
				currentMap = make(map[string]interface{})

				for key, value2 := range value {

					for _, value3 := range fields {
						if key == value3 {
							currentMap[key] = value2
						}
					}
				}
				outMap[index] = currentMap
			}
			returnByte, _ = json.Marshal(outMap)
		} else {
			returnByte, _ = json.Marshal(allMaps)
		}
	}
	return
}

func executeElasticGetFields(request *messaging.ObjectRequest) (returnByte []byte) {

	conn := getConnection()(request)
	skip := "0"

	if request.Extras["skip"] != nil {
		skip = request.Extras["skip"].(string)
	}

	take := "1"

	if request.Extras["take"] != nil {
		take = request.Extras["take"].(string)
	}

	query := "{\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + "*" + "\"}}}"

	data, err := conn.Search(request.Controls.Namespace, request.Controls.Class, nil, query)

	if err != nil {
		errorMessage := "Error retrieving object from elastic search : " + err.Error()
		request.Log(errorMessage)
		request.Log("Error Query : " + query)
	} else {
		request.Log("Successfully retrieved object from Elastic Search")

		var allMaps []map[string]interface{}
		allMaps = make([]map[string]interface{}, data.Hits.Len())

		for index, hit := range data.Hits.Hits {
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})

			byteData, _ := hit.Source.MarshalJSON()

			json.Unmarshal(byteData, &currentMap)

			allMaps[index] = currentMap
		}

		//get Number of fields
		noOfFields := 0
		for _, value := range allMaps {
			for key, _ := range value {

				if key != "__osHeaders" {
					noOfFields++
				}
			}
		}
		//create array to store
		var fieldList []string
		fieldList = make([]string, noOfFields)

		//store fields in array
		index := 0
		for _, value := range allMaps {
			for key, _ := range value {

				if key != "__osHeaders" {
					fmt.Println(key)
					fieldList[index] = key
					index++
				}
			}
		}

		returnByte, _ = json.Marshal(fieldList)

	}

	return
}

func executeElasticGetClasses(request *messaging.ObjectRequest) (returnByte []byte) {
	host := request.Configuration.ServerConfiguration["ELASTIC"]["Host"]
	port := request.Configuration.ServerConfiguration["ELASTIC"]["Port"]
	returnByte = getElasticByCURL(host, port, (request.Controls.Namespace + "/_mapping"))
	var mainMap map[string]interface{}
	mainMap = make(map[string]interface{})
	_ = json.Unmarshal(returnByte, &mainMap)
	var retArray []string
	//range through namespaces
	for _, index := range mainMap {
		for feature, typeDef := range index.(map[string]interface{}) {
			//if feature is MAPPING
			if feature == "mappings" {
				for typeName, _ := range typeDef.(map[string]interface{}) {
					retArray = append(retArray, typeName)
				}
			}
		}
	}
	returnByte, _ = json.Marshal(retArray)
	return
}

func executeElasticGetNamespaces(request *messaging.ObjectRequest) (returnByte []byte) {
	host := request.Configuration.ServerConfiguration["ELASTIC"]["Host"]
	port := request.Configuration.ServerConfiguration["ELASTIC"]["Port"]
	returnByte = getElasticByCURL(host, port, ("_mapping"))
	var mainMap map[string]interface{}
	mainMap = make(map[string]interface{})
	_ = json.Unmarshal(returnByte, &mainMap)
	var retArray []string
	//range through namespaces
	for index, _ := range mainMap {
		retArray = append(retArray, index)
	}
	returnByte, _ = json.Marshal(retArray)
	return
}

func executeElasticGetSelectedFields(request *messaging.ObjectRequest) (returnByte []byte) {
	conn := getConnection()(request)
	skip := "0"

	if request.Extras["skip"] != nil {
		skip = request.Extras["skip"].(string)
	}

	take := "100000"

	if request.Extras["take"] != nil {
		take = request.Extras["take"].(string)
	}

	isKeyDefined := false
	isKeywordDefined := false

	if (request.Body.Special.Extras["KeyValue"] != "") && (request.Body.Special.Extras["KeyValue"] != " ") && (request.Body.Special.Extras["KeyValue"] != nil) {
		isKeyDefined = true
	}

	if (request.Body.Special.Extras["Keyword"] != "") && (request.Body.Special.Extras["Keyword"] != " ") && (request.Body.Special.Extras["Keyword"] != nil) {
		isKeywordDefined = true
	}

	if (request.Body.Special.Extras["KeyValue"] == "*") || (request.Body.Special.Extras["Keyword"] == "*") {
		isKeyDefined = false
		isKeywordDefined = false
	}

	var query string

	if isKeyDefined {
		request.Log("Key Field Detected.")
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Body.Special.Extras["KeyValue"].(string)
		request.Log("Elastic Search GetSelected for key : " + key)
		data, err := conn.Get(request.Controls.Namespace, request.Controls.Class, key, nil)

		if err != nil {
			errorMessage := "Error retrieving object from elastic search : " + err.Error()
			request.Log(errorMessage)
		} else {
			request.Log("Successfully retrieved object from Elastic Search")
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})

			bytes, _ := data.Source.MarshalJSON()

			json.Unmarshal(bytes, &currentMap)

			if request.Body.Special.Parameters != "*" {

				//get fields list
				var fields []string
				fields = strings.Split(request.Body.Special.Parameters, " ")

				fmt.Print("Fields List : ")
				fmt.Println(fields)

				//create map to store data
				var outMap map[string]interface{}
				outMap = make(map[string]interface{})

				for key, value2 := range currentMap {

					for _, value3 := range fields {
						if key == value3 {
							outMap[key] = value2
						}
					}

				}
				returnByte, _ = json.Marshal(outMap)
			} else {
				returnByte, _ = json.Marshal(currentMap)
			}

		}

		return

	} else if isKeywordDefined {
		request.Log("Keyword Detected")
		query = "{\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + request.Body.Special.Extras["Keyword"].(string) + "\"}}}"
		request.Log(query)
	} else {
		request.Log("GET-ALL Detected!")
		query = "{\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + "*" + "\"}}}"
		request.Log(query)
	}

	data, err := conn.Search(request.Controls.Namespace, request.Controls.Class, nil, query)

	if err != nil {
		errorMessage := "Error retrieving object from elastic search : " + err.Error()
		request.Log(errorMessage)
		request.Log("Error Query : " + query)
	} else {
		request.Log("Successfully retrieved object from Elastic Search")

		var allMaps []map[string]interface{}
		allMaps = make([]map[string]interface{}, data.Hits.Len())

		for index, hit := range data.Hits.Hits {
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})

			byteData, _ := hit.Source.MarshalJSON()

			json.Unmarshal(byteData, &currentMap)

			allMaps[index] = currentMap
		}

		if request.Body.Special.Parameters != "*" {

			//get fields list
			var fields []string
			fields = strings.Split(request.Body.Special.Parameters, " ")

			fmt.Print("Fields List : ")
			fmt.Println(fields)

			//create map to store data
			var outMap []map[string]interface{}
			outMap = make([]map[string]interface{}, len(allMaps))

			for index, value := range allMaps {

				var currentMap map[string]interface{}
				currentMap = make(map[string]interface{})

				for key, value2 := range value {

					for _, value3 := range fields {
						if key == value3 {
							currentMap[key] = value2
						}
					}

				}

				outMap[index] = currentMap
			}

			returnByte, _ = json.Marshal(outMap)
		} else {
			returnByte, _ = json.Marshal(allMaps)
		}

	}

	return
}

// Helper Functions

func getElasticByCURL(host string, port string, path string) (returnByte []byte) {
	url := "http://" + host + ":" + port + "/" + path
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("CURL Request Failed")
	} else {
		fmt.Println("CURL Request Success!")
		body, _ := ioutil.ReadAll(resp.Body)
		returnByte = body
	}
	defer resp.Body.Close()

	return
}

func getConnection() func(request *messaging.ObjectRequest) *elastigo.Conn {

	var connection *elastigo.Conn

	return func(request *messaging.ObjectRequest) *elastigo.Conn {
		if connection == nil {
			host := request.Configuration.ServerConfiguration["ELASTIC"]["Host"]
			port := request.Configuration.ServerConfiguration["ELASTIC"]["Port"]
			request.Log("Establishing new connection for Elastic Search " + host + ":" + port)

			conn := elastigo.NewConn()
			conn.SetHosts([]string{host})
			conn.Port = port
			connection = conn
		}

		request.Log("Reusing existing Elastic Search connection ")
		return connection
	}
}

func getUniqueRecordMap(inputMap map[int]string) map[int]string {

	var outputMap map[int]string
	outputMap = make(map[int]string)

	count := 0
	for _, value := range inputMap {
		if len(outputMap) == 0 {
			outputMap[count] = value
			count++
		} else {
			isAvailable := false
			for _, value2 := range outputMap {
				if value == value2 {
					isAvailable = true
				}
			}
			if isAvailable {
				//Do Nothing
			} else {
				outputMap[count] = value
				count++
			}
		}
	}
	return outputMap
}
