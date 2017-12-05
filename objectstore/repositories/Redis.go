package repositories

import (
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"github.com/twinj/uuid"
	"github.com/xuyu/goredis"
	"strconv"
	"strings"
)

type RedisRepository struct {
}

func (repository RedisRepository) GetRepositoryName() string {
	return "Redis"
}

func getRedisConnection(request *messaging.ObjectRequest) (client *goredis.Redis, isError bool, errorMessage string) {

	password := request.Configuration.ServerConfiguration["REDIS"]["Password"]

	urlStart := "tcp://"
	if password != "" {
		urlStart += "auth:" + password
	}
	urlStart += "@"

	isError = false
	client, err := goredis.DialURL(urlStart + request.Configuration.ServerConfiguration["REDIS"]["Host"] + ":" + request.Configuration.ServerConfiguration["REDIS"]["Port"] + "/4?timeout=60s&maxidle=60")
	if err != nil {
		isError = true
		errorMessage = err.Error()
		request.Log("Error! Can't connect to server!error")

	}
	request.Log("Reusing existing GoRedis connection")
	return
}

func (repository RedisRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-ALL")
	response := RepositoryResponse{}
	client, isError, errorMessage := getRedisConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + "*"
		value, err := client.Keys(key)

		if err != nil {
			response.IsSuccess = false
			request.Log("Error getting value by key for All object in Redis : " + key + ", " + err.Error())
			response.GetErrorResponse("Error getting value by key for All objects in Redis" + err.Error())
		}

		var temp []string
		temp = make([]string, len(value))

		for x := 0; x < len(value); x++ {
			val, _ := client.Get(value[x])
			temp[x] = string(val[:])
		}

		take := len(temp)

		if request.Extras["take"] != nil {
			take, _ = strconv.Atoi(request.Extras["take"].(string))
		}

		skip := 0

		if request.Extras["skip"] != nil {
			skip, _ = strconv.Atoi(request.Extras["skip"].(string))
		}

		if len(temp) == 0 {
			response.IsSuccess = true
			response.Message = "No objects found in Redis"
			var emptyMap map[string]interface{}
			emptyMap = make(map[string]interface{})
			byte, _ := json.Marshal(emptyMap)
			response.GetResponseWithBody(byte)
		}

		var returnValues []map[string]interface{}

		for _, valueIndex := range temp[skip:(skip + take)] {
			var tempMap map[string]interface{}
			tempMap = make(map[string]interface{})
			err = json.Unmarshal([]byte(valueIndex), &tempMap)
			if err != nil {
				request.Log("Error converting Json to Map : " + err.Error())
			} else {
				if request.Controls.SendMetaData == "false" {
					delete(tempMap, "__osHeaders")
				}
				returnValues = append(returnValues, tempMap)
			}
		}

		byteValue, _ := json.Marshal(returnValues)
		if err != nil {
			response.IsSuccess = false
			request.Log("Error getting value by key for All object in Redis : " + key + ", " + err.Error())
			response.GetErrorResponse("Error getting value by key for all objects in Redis" + err.Error())
		} else {
			response.IsSuccess = true
			response.GetResponseWithBody(byteValue)
			response.Message = "Successfully retrieved all object in Redis"
			request.Log(response.Message)
		}

	}
	return response
}

func (repository RedisRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("GetSearch not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-QUERY!")
	response := RepositoryResponse{}
	queryType := request.Body.Query.Type

	switch queryType {
	case "Query":
		if request.Body.Query.Parameters != "*" {
			request.Log("Support for SQL Query not implemented in REDIS repository")
			return getDefaultNotImplemented()
		} else {
			return repository.GetAll(request)
		}
	default:
		request.Log(queryType + " not implemented in Redis Db repository")
		return getDefaultNotImplemented()
	}

	return response
}

func (repository RedisRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-BY-KEY")
	response := RepositoryResponse{}
	client, isError, errorMessage := getRedisConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		key := getNoSqlKey(request)
		value, err := client.Get(key)

		var tempMap map[string]interface{}
		tempMap = make(map[string]interface{})
		err = json.Unmarshal(value, &tempMap)

		if len(tempMap) == 0 {
			response.IsSuccess = true
			response.Message = "No objects found in Redis"
			var emptyMap map[string]interface{}
			emptyMap = make(map[string]interface{})
			byte, _ := json.Marshal(emptyMap)
			response.GetResponseWithBody(byte)
		}

		if err != nil {
			request.Log("Error converting Json to Map : " + err.Error())
		} else {
			if request.Controls.SendMetaData == "false" {
				delete(tempMap, "__osHeaders")
			}
		}

		byteValue, _ := json.Marshal(tempMap)

		if err != nil {
			response.IsSuccess = false
			request.Log("Error getting value by key for object in Redis : " + key + ", " + err.Error())
			response.GetErrorResponse("Error getting value by key for one object in Redis" + err.Error())
		}
		if err != nil {
			response.IsSuccess = false
			request.Log("Error getting value by key for object in Redis : " + key + ", " + err.Error())
			response.GetErrorResponse("Error getting value by key for one object in Redis" + err.Error())
		} else {
			response.IsSuccess = true
			response.GetResponseWithBody(byteValue)
			response.Message = "Successfully retrieved one object in Redis"
			request.Log(response.Message)
		}

	}
	return response
}

func (repository RedisRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-MULTIPLE")
	return setManyRedis(request)
}

func (repository RedisRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-SINGLE")
	return setOneRedis(request)
}

func setOneRedis(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	client, isError, errorMessage := getRedisConnection(request)
	keyValue := getRedisRecordID(request, nil)
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else if keyValue != "" {
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + keyValue
		request.Body.Object[request.Body.Parameters.KeyProperty] = keyValue
		value := getStringByObject(request.Body.Object)

		err := client.Set(key, value, 0, 0, false, false)

		if err != nil {
			response.IsSuccess = false
			request.Log("Error inserting/updating object in Redis : " + key + ", " + err.Error())
			response.GetErrorResponse("Error inserting/updating one object in Redis" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted/updated one object in Redis"
			request.Log(response.Message)
		}
	}

	//Update Response
	var Data []map[string]interface{}
	Data = make([]map[string]interface{}, 1)
	var actualData map[string]interface{}
	actualData = make(map[string]interface{})
	actualData["ID"] = keyValue
	Data[0] = actualData
	response.Data = Data

	return response
}

func setManyRedis(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	client, isError, errorMessage := getRedisConnection(request)
	var idData map[string]interface{}
	idData = make(map[string]interface{})
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		isError := false
		index := 0
		for _, object := range request.Body.Objects {
			index++
			keyValue := getRedisRecordID(request, object)

			if keyValue == "" {
				response.IsSuccess = false
				response.Message = "Failed inserting multiple object in Cassandra"
				request.Log(response.Message)
				request.Log("Inavalid ID request")
				return response
			}
			key := request.Controls.Namespace + "." + request.Controls.Class + "." + keyValue
			object[request.Body.Parameters.KeyProperty] = keyValue

			idData[strconv.Itoa(index)] = keyValue

			value := getStringByObject(object)
			err := client.Set(key, value, 0, 0, false, false)

			if err != nil {
				isError = true
				errorMessage = err.Error()
				break
			}
		}

		if isError == true {
			response.IsSuccess = false
			request.Log("Error inserting/updating multiple objects in Redis : " + errorMessage)
			response.GetErrorResponse("Error inserting/updating multiple objects in Redis" + errorMessage)
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted/updated multiple objects in Redis"
			request.Log(response.Message)
		}
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

func (repository RedisRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-MULTIPLE")
	return setManyRedis(request)
}

func (repository RedisRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-SINGLE")
	return setOneRedis(request)
}

func (repository RedisRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-MULTIPLE")
	response := RepositoryResponse{}
	client, isError, errorMessage := getRedisConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		isError := false

		for _, object := range request.Body.Objects {
			key := getNoSqlKeyById(request, object)
			status, err := client.Expire(key, 0)
			if !status {
				isError = true
				request.Log("Error deleting object in Redis!" + err.Error())
			}
		}

		if isError == true {
			response.IsSuccess = false
			request.Log("Error deleting All multiple objects in Redis, some deletions failed! : " + errorMessage)
			response.GetErrorResponse("Error deleting All multiple objects in Redis, some deletions failed!" + errorMessage)
		} else {
			response.IsSuccess = true
			response.Message = "Successfully deleted multiple objects in Redis"
			request.Log(response.Message)
		}
	}

	return response
}

func (repository RedisRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-SINGLE")
	response := RepositoryResponse{}
	client, isError, errorMessage := getRedisConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id

		isAvailable, err := client.Exists(key)
		if err != nil {
			response.IsSuccess = false
			request.Log("Error deleting object in Redis : " + key + ", " + err.Error())
			response.GetErrorResponse("Error deleting one object in Redis" + err.Error())
		}
		if isAvailable {
			status, err := client.Expire(key, 0)
			if !status {
				response.IsSuccess = false
				request.Log("Error deleting object in Redis!" + err.Error())
				response.GetErrorResponse("Error deleting object in Redis!" + err.Error())
			} else {
				response.IsSuccess = true
				request.Log("Successfully deleted object in Redis!")
			}
		} else {
			response.IsSuccess = false
			response.Message = "No such value available to delete"
			request.Log(response.Message)
		}

	}

	return response
}

func (repository RedisRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Special not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) Test(request *messaging.ObjectRequest) {

}

func getRedisRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
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
		client, isError, _ := getRedisConnection(request)
		if isError {
			returnID = ""
			request.Log("Connecting to REDIS Failed! ")
		} else {
			//read Attributes table

			key := request.Controls.Namespace + "." + request.Controls.Class + "#domainClassAttributes"
			rawBytes, err := client.Get(key)

			if err != nil || string(rawBytes) == "" {
				request.Log("This is a freshly created Class. Inserting new Class record.")
				var ObjectBody map[string]interface{}
				ObjectBody = make(map[string]interface{})
				ObjectBody["maxCount"] = "1"
				ObjectBody["version"] = uuid.NewV1().String()
				err = client.Set(key, getStringByObject(ObjectBody), 0, 0, false, false)
				if err != nil {
					request.Log("Update of maxCount Failed : " + err.Error())
					returnID = ""
				} else {
					returnID = "1"
				}
			} else {
				var UpdatedCount int
				var returnData map[string]interface{}
				returnData = make(map[string]interface{})

				json.Unmarshal(rawBytes, &returnData)

				for fieldName, fieldvalue := range returnData {
					if strings.ToLower(fieldName) == "maxcount" {
						UpdatedCount, _ = strconv.Atoi(fieldvalue.(string))
						UpdatedCount++
						returnID = strconv.Itoa(UpdatedCount)
						break
					}
				}

				//update the table
				//save to attributes table
				returnData["maxCount"] = returnID
				returnData["version"] = uuid.NewV1().String()
				err = client.Set(key, getStringByObject(returnData), 0, 0, false, false)
				if err != nil {
					request.Log("Update of maxCount Failed")
					returnID = ""
				}

			}

		}
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

func (repository RedisRepository) ClearCache(request *messaging.ObjectRequest) {
}
