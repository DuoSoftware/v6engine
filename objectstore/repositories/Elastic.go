package repositories

import (
	//"duov6.com/term"
	"duov6.com/objectstore/connmanager"
	"duov6.com/objectstore/messaging"
	"duov6.com/queryparser"
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
	return repository.search(request, "*")
}

func (repository ElasticRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GETSEARCH")
	return repository.search(request, request.Body.Query.Parameters)
}

func (repository ElasticRepository) search(request *messaging.ObjectRequest, searchStr string) RepositoryResponse {
	response := RepositoryResponse{}
	conn := repository.getConnection(request)

	skip := "0"

	if request.Extras["skip"] != nil {
		skip = request.Extras["skip"].(string)
	}

	take := "100"

	if request.Extras["take"] != nil {
		take = request.Extras["take"].(string)
	}

	orderbyfield := ""
	var query string

	if request.Extras["orderby"] != nil {
		orderbyfield = request.Extras["orderby"].(string)
		operator := "asc"
		query = "{\"sort\" : [{\"" + orderbyfield + "\" : {\"order\" : \"" + operator + "\"}}],\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + searchStr + "\"}}}"
	} else if request.Extras["orderbydsc"] != nil {
		orderbyfield = request.Extras["orderbydsc"].(string)
		operator := "desc"
		query = "{\"sort\" : [{\"" + orderbyfield + "\" : {\"order\" : \"" + operator + "\"}}],\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + searchStr + "\"}}}"
	} else {
		query = "{\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + searchStr + "\"}}}"
	}

	data, err := conn.Search(request.Controls.Namespace, request.Controls.Class, nil, query)

	if err != nil {
		response.GetResponseWithBody(getEmptyByteObject())
	} else {
		var allMaps []map[string]interface{}
		allMaps = make([]map[string]interface{}, data.Hits.Len())
		for index, hit := range data.Hits.Hits {
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})

			byteData, _ := hit.Source.MarshalJSON()
			json.Unmarshal(byteData, &currentMap)
			delete(currentMap, "__osHeaders")
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
			fieldsInByte := repository.executeQuery(request)
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
			//Check if just * then execute GET-SEARCH method
			//term.Write("Redirecting to GET-SEARCH!", 2)
			return repository.search(request, request.Body.Query.Parameters)
		}
	default:
		return repository.search(request, request.Body.Query.Parameters)

	}

	return response
}

func (repository ElasticRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-BY-KEY")
	response := RepositoryResponse{}
	conn := repository.getConnection(request)

	key := getNoSqlKey(request)
	data, err := conn.Get(request.Controls.Namespace, request.Controls.Class, key, nil)

	if err != nil {
		response.GetResponseWithBody(getEmptyByteObject())
	} else {
		bytes, err := data.Source.MarshalJSON()
		//Get Data to struct
		var originalData map[string]interface{}
		originalData = make(map[string]interface{})
		json.Unmarshal(bytes, &originalData)
		delete(originalData, "__osHeaders")

		bytes, err = json.Marshal(originalData)
		if err != nil {
			errorMessage := "Elastic search JSON marshal error : " + err.Error()
			//term.Write(err.Error(), 1)
			response.GetErrorResponse(errorMessage)

		} else {
			response.GetResponseWithBody(bytes)
		}

	}

	return response
}

func (repository ElasticRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-MULTIPLE")
	return repository.setManyElastic(request)
}

func (repository ElasticRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-SINGLE")
	return repository.setOneElastic(request)
}

func (repository ElasticRepository) setOneElastic(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	conn := repository.getConnection(request)

	key := ""
	id := ""
	if request.Body.Object["OriginalIndex"] != nil {
		key = request.Body.Object["OriginalIndex"].(string)
	} else {
		id = repository.getRecordID(request, request.Body.Object)
		request.Body.Object[request.Body.Parameters.KeyProperty] = id
		key = request.Controls.Namespace + "." + request.Controls.Class + "." + id
	}
	_, err := conn.Index(request.Controls.Namespace, request.Controls.Class, key, nil, request.Body.Object)
	if err != nil {
		//term.Write(err.Error(), 1)
		errorMessage := "Elastic Search Single Insert Error : " + err.Error()
		response.GetErrorResponse(errorMessage)
		return response
	} else {
		response.IsSuccess = true
		response.Message = "Successfully inserted one to elastic search"
	}

	//Update Response
	var Data []map[string]interface{}
	Data = make([]map[string]interface{}, 1)
	var actualData map[string]interface{}
	actualData = make(map[string]interface{})
	actualData["ID"] = id
	Data[0] = actualData
	response.Data = Data
	return response
}

/*func (repository ElasticRepository) setManyElastic(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	conn := repository.getConnection(request)

	indexer := conn.NewBulkIndexer(200)
	nowTime := time.Now()

	CountIndex := 0
	var Data map[string]interface{}
	Data = make(map[string]interface{})

	noOfElementsPerSet := 100
	noOfSets := (len(request.Body.Objects) / noOfElementsPerSet)
	remainderFromSets := 0
	remainderFromSets = (len(request.Body.Objects) - (noOfSets * noOfElementsPerSet))

	startIndex := 0
	stopIndex := noOfElementsPerSet

	var status []bool

	if remainderFromSets == 0 {
		status = make([]bool, noOfSets)
	} else {
		status = make([]bool, (noOfSets + 1))
	}

	statusIndex := 0

	for x := 0; x < noOfSets; x++ {
		for index, obj := range request.Body.Objects[startIndex:stopIndex] {
			nosqlid := ""
			if obj["OriginalIndex"] != nil {
				nosqlid = obj["OriginalIndex"].(string)
				Data[strconv.Itoa(CountIndex)] = nosqlid
			} else {
				id := repository.getRecordID(request, obj)
				nosqlid = request.Controls.Namespace + "." + request.Controls.Class + "." + id
				request.Body.Objects[index][request.Body.Parameters.KeyProperty] = id
				Data[strconv.Itoa(CountIndex)] = id
			}
			CountIndex++
			indexer.Index(request.Controls.Namespace, request.Controls.Class, nosqlid, "10", &nowTime, obj, false)
		}
		indexer.Start()
		numerrors := indexer.NumErrors()

		if numerrors != 0 {
			status[statusIndex] = false
		} else {
			status[statusIndex] = true
			fmt.Println("Inserted Stub : " + strconv.Itoa(statusIndex))
		}
		statusIndex++
		startIndex += noOfElementsPerSet
		stopIndex += noOfElementsPerSet

		time.Sleep(1200 * time.Millisecond)

	}

	if remainderFromSets > 0 {
		start := len(request.Body.Objects) - remainderFromSets

		for index, obj := range request.Body.Objects[start:len(request.Body.Objects)] {
			nosqlid := ""
			if obj["OriginalIndex"] != nil {
				nosqlid = obj["OriginalIndex"].(string)
				Data[strconv.Itoa(CountIndex)] = nosqlid
			} else {
				id := repository.getRecordID(request, obj)
				nosqlid = request.Controls.Namespace + "." + request.Controls.Class + "." + id
				request.Body.Objects[index][request.Body.Parameters.KeyProperty] = id
				Data[strconv.Itoa(CountIndex)] = id
			}
			CountIndex++
			indexer.Index(request.Controls.Namespace, request.Controls.Class, nosqlid, "10", &nowTime, obj, false)
		}
		indexer.Start()
		numerrors := indexer.NumErrors()

		if numerrors != 0 {
			status[statusIndex] = false
		} else {
			status[statusIndex] = true
			fmt.Println("Inserted Last Stub!")
		}
		statusIndex++
	}

	isAllCompleted := true

	for _, val := range status {
		if !val {
			isAllCompleted = false
			break
		}
	}

	if isAllCompleted {
		response.IsSuccess = true
		response.Message = "Successfully inserted bulk to Elastic Search"
	} else {
		response.IsSuccess = false
		response.Message = "Error Inserting Some Objects"
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
}*/

func (repository ElasticRepository) setManyElastic(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	conn := repository.getConnection(request)

	var Data map[string]interface{}
	Data = make(map[string]interface{})

	for index, obj := range request.Body.Objects {
		id := repository.getRecordID(request, obj)
		Data[strconv.Itoa(index)] = id
		request.Body.Objects[index][request.Body.Parameters.KeyProperty] = id
	}

	noOfElementsPerSet := 500
	noOfSets := (len(request.Body.Objects) / noOfElementsPerSet)
	remainderFromSets := 0
	remainderFromSets = (len(request.Body.Objects) - (noOfSets * noOfElementsPerSet))

	startIndex := 0
	stopIndex := noOfElementsPerSet

	var status []bool

	if remainderFromSets == 0 {
		status = make([]bool, noOfSets)
	} else {
		status = make([]bool, (noOfSets + 1))
	}

	statusIndex := 0

	for x := 0; x < noOfSets; x++ {
		tempStatus := repository.insertRecordStub(request, request.Body.Objects[startIndex:stopIndex], conn)
		status[statusIndex] = tempStatus

		if tempStatus {
			fmt.Println("Inserted Stub : " + strconv.Itoa(statusIndex))
		} else {
			fmt.Println("Inserting Failed Stub : " + strconv.Itoa(statusIndex))
		}

		statusIndex += 1
		startIndex += noOfElementsPerSet
		stopIndex += noOfElementsPerSet

		time.Sleep(1 * time.Millisecond)

	}

	if remainderFromSets > 0 {
		start := len(request.Body.Objects) - remainderFromSets

		tempStatus := repository.insertRecordStub(request, request.Body.Objects[start:len(request.Body.Objects)], conn)
		status[statusIndex] = tempStatus

		if tempStatus {
			fmt.Println("Inserted Stub : " + strconv.Itoa(statusIndex))
		} else {
			fmt.Println("Inserting Failed Stub : " + strconv.Itoa(statusIndex))
		}

		statusIndex += 1
	}

	isAllCompleted := true

	for _, val := range status {
		if !val {
			isAllCompleted = false
			break
		}
	}

	if isAllCompleted {
		response.IsSuccess = true
		response.Message = "Successfully inserted bulk to Elastic Search"
	} else {
		response.IsSuccess = false
		response.Message = "Error Inserting Some Objects"
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

func (repository ElasticRepository) insertRecordStub(request *messaging.ObjectRequest, records []map[string]interface{}, conn *elastigo.Conn) (status bool) {
	status = true
	indexer := conn.NewBulkIndexerErrors(1000, 60)
	indexer.Start()
	for _, obj := range records {
		nosqlid := ""
		if obj["OriginalIndex"] != nil {
			nosqlid = obj["OriginalIndex"].(string)
		} else {
			nosqlid = getNoSqlKeyById(request, obj)
		}
		indexer.Index(request.Controls.Namespace, request.Controls.Class, nosqlid, "", "", nil, obj)
	}
	indexer.Stop()

	return

}

// Elastic v1.7 code -- Don't Delete
// func (repository ElasticRepository) insertRecordStub(request *messaging.ObjectRequest, records []map[string]interface{}, conn *elastigo.Conn) (status bool) {
// 	status = true

// 	indexer := conn.NewBulkIndexer(100)
// 	nowTime := time.Now()

// 	for _, obj := range records {
// 		nosqlid := ""
// 		if obj["OriginalIndex"] != nil {
// 			nosqlid = obj["OriginalIndex"].(string)
// 		} else {
// 			nosqlid = getNoSqlKeyById(request, obj)
// 		}
// 		//indexer.Index(request.Controls.Namespace, request.Controls.Class, nosqlid, "10", &nowTime, obj, false)
// 		//func (b *BulkIndexer) Index(index string, _type string, id, parent, ttl string, date *time.Time, data interface{}) error
// 		indexer.Index(request.Controls.Namespace, request.Controls.Class, nosqlid, "", "10", &nowTime, obj)
// 	}
// 	indexer.Start()
// 	numerrors := indexer.NumErrors()

// 	if numerrors != 0 {
// 		status = false
// 	}
// 	return
// }

func (repository ElasticRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-MULTIPLE")
	return repository.setManyElastic(request)
}

func (repository ElasticRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-SINGLE")
	return repository.setOneElastic(request)
}

func (repository ElasticRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-MULTIPLE")
	response := RepositoryResponse{}

	conn := repository.getConnection(request)

	for _, object := range request.Body.Objects {
		key := getNoSqlKeyById(request, object)
		_, err := conn.Delete(request.Controls.Namespace, request.Controls.Class, key, nil)
		if err != nil {
			errorMessage := "Elastic Search single delete error : " + err.Error()
			//term.Write(err.Error(), 1)
			response.GetErrorResponse(errorMessage)
		} else {
			response.IsSuccess = true
			response.Message = "Successfully deleted one in elastic search"
		}
	}

	return response

}

func (repository ElasticRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-SINGLE")
	response := RepositoryResponse{}

	conn := repository.getConnection(request)
	_, err := conn.Delete(request.Controls.Namespace, request.Controls.Class, getNoSqlKey(request), nil)
	if err != nil {
		errorMessage := "Elastic Search single delete error : " + err.Error()
		request.Log(errorMessage)
		//term.Write(err.Error(), 1)
		response.GetErrorResponse(errorMessage)
	} else {
		response.IsSuccess = true
		response.Message = "Successfully deleted one in elastic search"
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
		fieldsInByte := repository.executeGetFields(request)
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
		fieldsInByte := repository.executeGetClasses(request)
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
		fieldsInByte := repository.executeGetNamespaces(request)
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
		fieldsInByte := repository.executeGetSelectedFields(request)
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
		return repository.search(request, request.Body.Special.Parameters)

	}

	return response

}

func (repository ElasticRepository) Test(request *messaging.ObjectRequest) {

}

//SUB FUNCTIONS
//Functions from SPECIAL and QUERY

func (repository ElasticRepository) executeQuery(request *messaging.ObjectRequest) (returnByte []byte) {
	conn := repository.getConnection(request)

	var query string

	query, err := queryparser.GetElasticQuery(request.Body.Query.Parameters, request.Controls.Namespace, request.Controls.Class)

	if err != nil {
		returnByte = getEmptyByteObject()
		return
	}

	//fmt.Print("Elastic JSON Query : ")
	//fmt.Println(query)

	data, err := conn.Search(request.Controls.Namespace, request.Controls.Class, nil, query)

	if err != nil {
		returnByte = getEmptyByteObject()
		return
	} else {
		var allMaps []map[string]interface{}
		allMaps = make([]map[string]interface{}, data.Hits.Len())
		for index, hit := range data.Hits.Hits {
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})
			byteData, _ := hit.Source.MarshalJSON()
			json.Unmarshal(byteData, &currentMap)
			delete(currentMap, "__osHeaders")
			allMaps[index] = currentMap
		}

		returnByte, _ = json.Marshal(allMaps)
	}

	return returnByte
}

func (repository ElasticRepository) executeGetFields(request *messaging.ObjectRequest) (returnByte []byte) {

	conn := repository.getConnection(request)

	query := "{\"from\": 0, \"size\": 1,\"query\":{\"query_string\" : {\"query\" : \"" + "*" + "\"}}}"

	data, err := conn.Search(request.Controls.Namespace, request.Controls.Class, nil, query)

	if err != nil {
		//term.Write(err.Error(), 1)
		returnByte = getEmptyByteObject()
	} else {
		var allMaps []map[string]interface{}
		allMaps = make([]map[string]interface{}, data.Hits.Len())

		for index, hit := range data.Hits.Hits {
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})
			byteData, _ := hit.Source.MarshalJSON()
			json.Unmarshal(byteData, &currentMap)
			allMaps[index] = currentMap
		}
		//create array to store
		var fieldList []string

		//store fields in array
		for key, _ := range allMaps[0] {
			if key != "__osHeaders" {
				fieldList = append(fieldList, key)
			}
		}
		returnByte, _ = json.Marshal(fieldList)
	}

	return
}

func (repository ElasticRepository) executeGetClasses(request *messaging.ObjectRequest) (returnByte []byte) {
	host := request.Configuration.ServerConfiguration["ELASTIC"]["Host"]
	port := request.Configuration.ServerConfiguration["ELASTIC"]["Port"]
	returnByte = repository.getByCURL(host, port, (request.Controls.Namespace + "/_mapping"))
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

func (repository ElasticRepository) executeGetNamespaces(request *messaging.ObjectRequest) (returnByte []byte) {
	host := request.Configuration.ServerConfiguration["ELASTIC"]["Host"]
	port := request.Configuration.ServerConfiguration["ELASTIC"]["Port"]
	returnByte = repository.getByCURL(host, port, ("_mapping"))
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

func (repository ElasticRepository) executeGetSelectedFields(request *messaging.ObjectRequest) (returnByte []byte) {

	skip := "0"

	if request.Extras["skip"] != nil {
		skip = request.Extras["skip"].(string)
	}

	take := "100"

	if request.Extras["take"] != nil {
		take = request.Extras["take"].(string)
	}
	conn := repository.getConnection(request)

	fieldNames := strings.Split(request.Body.Special.Parameters, " ")

	fieldString := "\"" + fieldNames[0] + "\""

	for index := 1; index < len(fieldNames); index++ {
		fieldString += "," + "\"" + fieldNames[index] + "\""
	}

	query := "{\"_source\":[" + fieldString + "],\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + "*" + "\"}}}"

	//fmt.Println(query)

	data, err := conn.Search(request.Controls.Namespace, request.Controls.Class, nil, query)

	if err != nil {
		//term.Write(err.Error(), 1)
	} else {
		var allMaps []map[string]interface{}
		allMaps = make([]map[string]interface{}, data.Hits.Len())

		for index, hit := range data.Hits.Hits {
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})
			byteData, _ := hit.Source.MarshalJSON()
			json.Unmarshal(byteData, &currentMap)
			allMaps[index] = currentMap
		}

		//fmt.Println(allMaps)

		returnByte, _ = json.Marshal(allMaps)

	}

	return
}

// Helper Functions

func (repository ElasticRepository) getByCURL(host string, port string, path string) (returnByte []byte) {
	url := "http://" + host + ":" + port + "/" + path
	req, err := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		//term.Write(err.Error(), 1)
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		returnByte = body
	}
	defer resp.Body.Close()
	return
}

func (repository ElasticRepository) getConnection(request *messaging.ObjectRequest) (connection *elastigo.Conn) {
	connInt := connmanager.Get("ELASTIC", request.Controls.Namespace)
	if connInt != nil {
		connection = connInt.(*elastigo.Conn)
	} else {
		host := request.Configuration.ServerConfiguration["ELASTIC"]["Host"]
		port := request.Configuration.ServerConfiguration["ELASTIC"]["Port"]
		request.Log("Establishing new connection for Elastic Search " + host + ":" + port)

		conn := elastigo.NewConn()
		conn.SetHosts([]string{host})
		conn.Port = port
		connection = conn
		connmanager.Set("ELASTIC", request.Controls.Namespace, connection)
	}
	return
}

func (repository ElasticRepository) getRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
	conn := repository.getConnection(request)
	isAutoIncrementing := false
	isRandomKeyID := false

	if (obj[request.Body.Parameters.KeyProperty].(string) == "-999") || (request.Body.Parameters.AutoIncrement == true) {
		isAutoIncrementing = true
	} else if (obj[request.Body.Parameters.KeyProperty].(string) == "-888") || (request.Body.Parameters.GUIDKey == true) {
		isRandomKeyID = true
	}

	if isRandomKeyID {
		returnID = uuid.NewV1().String()
	} else if isAutoIncrementing {
		key := request.Controls.Class
		data, err := conn.Get(request.Controls.Namespace, "domainClassAttributes", key, nil)
		maxCount := ""

		if err != nil {
			request.Log("No record Found. This is a NEW record. Inserting new attribute value")
			var newRecord map[string]interface{}
			newRecord = make(map[string]interface{})
			newRecord["class"] = request.Controls.Class
			newRecord["maxCount"] = "1"
			newRecord["version"] = uuid.NewV1().String()
			_, err = conn.Index(request.Controls.Namespace, "domainClassAttributes", key, nil, newRecord)

			if err != nil {
				//term.Write(err.Error(), 1)
				return uuid.NewV1().String()
			} else {
				return "1"
			}

		} else {
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})
			byteData, err := data.Source.MarshalJSON()
			if err != nil {
				//term.Write(err.Error(), 1)
				return uuid.NewV1().String()
			}
			json.Unmarshal(byteData, &currentMap)
			maxCount = currentMap["maxCount"].(string)
			tempCount, err := strconv.Atoi(maxCount)
			maxCount = strconv.Itoa(tempCount + 1)

			//Update Table
			var newRecord map[string]interface{}
			newRecord = make(map[string]interface{})
			newRecord["class"] = request.Controls.Class
			newRecord["maxCount"] = maxCount
			newRecord["version"] = uuid.NewV1().String()
			_, err = conn.Index(request.Controls.Namespace, "domainClassAttributes", request.Controls.Class, nil, newRecord)
			if err != nil {
				//term.Write(err.Error(), 1)
				return uuid.NewV1().String()
			} else {
				return maxCount
			}
		}
	} else {
		return obj[request.Body.Parameters.KeyProperty].(string)
	}
	return
}
