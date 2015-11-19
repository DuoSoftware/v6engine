package repositories

import (
	"duov6.com/objectstore/connmanager"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/queryparser"
	"duov6.com/term"
	"encoding/json"
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

	take := "100000"

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
		if strings.Contains(err.Error(), "record not found") {
			var empty map[string]interface{}
			empty = make(map[string]interface{})
			response.GetSuccessResByObject(empty)
		} else {
			errorMessage := "Error retrieving object from elastic search : " + err.Error()
			term.Write(errorMessage, 1)
			response.GetErrorResponse(errorMessage)
		}
	} else {
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
			//Check if just STAR then execute GET-SEARCH method
			term.Write("Redirecting to GET-SEARCH!", 2)
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
		if strings.Contains(err.Error(), "record not found") {
			var empty map[string]interface{}
			empty = make(map[string]interface{})
			response.GetSuccessResByObject(empty)
		}
	} else {
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
			term.Write(err.Error(), 1)
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
		term.Write(err.Error(), 1)
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

func (repository ElasticRepository) setManyElastic(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	conn := repository.getConnection(request)

	indexer := conn.NewBulkIndexer(200)
	nowTime := time.Now()

	CountIndex := 0
	var Data map[string]interface{}
	Data = make(map[string]interface{})

	for index, obj := range request.Body.Objects {
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
		term.Write("Elastic Search bulk insert error!", 1)
		response.GetErrorResponse("Elastic Search bulk insert error")
	} else {
		response.IsSuccess = true
		response.Message = "Successfully inserted bulk to Elastic Search"
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
			term.Write(err.Error(), 1)
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
		term.Write(err.Error(), 1)
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

	searchStr, isSelectedFields, selectedFields, fromClass := queryparser.GetQuery(request.Body.Query.Parameters)
	query := "{\"query\":{\"query_string\" : {\"query\" : \"" + searchStr + "\"}}}"
	term.Write(query, 2)
	var data elastigo.SearchResult
	var err error

	if fromClass == "" {
		data, err = conn.Search(request.Controls.Namespace, request.Controls.Class, nil, query)
	} else {
		data, err = conn.Search(request.Controls.Namespace, fromClass, nil, query)
	}
	if err != nil {
		term.Write(err.Error(), 1)
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

		for _, value := range selectedFields {
			if value == "*" {
				isSelectedFields = false
			}
		}

		if isSelectedFields {
			var fields []string
			fields = selectedFields

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

func (repository ElasticRepository) executeGetFields(request *messaging.ObjectRequest) (returnByte []byte) {

	conn := repository.getConnection(request)

	query := "{\"query\":{\"query_string\" : {\"query\" : \"" + "*" + "\"}}}"

	data, err := conn.Search(request.Controls.Namespace, request.Controls.Class, nil, query)

	if err != nil {
		term.Write(err.Error(), 1)
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
					fieldList[index] = key
					index++
				}
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
	conn := repository.getConnection(request)

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
		//Key Field Detected
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Body.Special.Extras["KeyValue"].(string)
		data, err := conn.Get(request.Controls.Namespace, request.Controls.Class, key, nil)
		if err != nil {
			term.Write(err.Error(), 1)
		} else {
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})
			bytes, _ := data.Source.MarshalJSON()
			json.Unmarshal(bytes, &currentMap)

			if request.Body.Special.Parameters != "*" {
				//get fields list
				var fields []string
				fields = strings.Split(request.Body.Special.Parameters, " ")

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
		//Keyword Detected
		query = "{\"query\":{\"query_string\" : {\"query\" : \"" + request.Body.Special.Extras["Keyword"].(string) + "\"}}}"
	} else {
		//GET-ALL Detected!
		query = "{\"query\":{\"query_string\" : {\"query\" : \"" + "*" + "\"}}}"
	}

	data, err := conn.Search(request.Controls.Namespace, request.Controls.Class, nil, query)

	if err != nil {
		term.Write(err.Error(), 1)
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

		if request.Body.Special.Parameters != "*" {
			//get fields list
			var fields []string
			fields = strings.Split(request.Body.Special.Parameters, " ")

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

func (repository ElasticRepository) getByCURL(host string, port string, path string) (returnByte []byte) {
	url := "http://" + host + ":" + port + "/" + path
	req, err := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		term.Write(err.Error(), 1)
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
				term.Write(err.Error(), 1)
				return ""
			} else {
				return "1"
			}

		} else {
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})
			byteData, err := data.Source.MarshalJSON()
			if err != nil {
				term.Write(err.Error(), 1)
				return ""
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
				term.Write(err.Error(), 1)
				return ""
			} else {
				return maxCount
			}
		}
	} else {
		return obj[request.Body.Parameters.KeyProperty].(string)
	}
	return
}
