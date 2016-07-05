package repositories

import (
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"fmt"
	"github.com/twinj/uuid"
	"gopkg.in/mgo.v2"
	dbUse "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"strings"
)

type MongoRepository struct {
}

func (repository MongoRepository) GetRepositoryName() string {
	return "Mongo DB"
}

func getMongoConnection(request *messaging.ObjectRequest) (client *mgo.Collection, isError bool, errorMessage string) {

	isError = false

	session, err := mgo.Dial(request.Configuration.ServerConfiguration["MONGO"]["Url"])
	if err != nil {
		isError = false
		errorMessage = err.Error()
		request.Log("Mongo connection initilizing failed!")
	}

	namespace := getSQLnamespace(request)
	client = session.DB(namespace).C(request.Controls.Class)
	return
}

func getCustomMongoConnection(request *messaging.ObjectRequest, customNamespace string, customClass string) (client *mgo.Collection, isError bool, errorMessage string) {

	isError = false

	session, err := mgo.Dial(request.Configuration.ServerConfiguration["MONGO"]["Url"])
	if err != nil {
		isError = false
		errorMessage = err.Error()
		request.Log("Mongo connection initilizing failed!")
	}

	//namespace := getSQLnamespace(request)
	client = session.DB(customNamespace).C(customClass)
	request.Log("Reusing existing Mongo connection")
	return
}

func (repository MongoRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-ALL")
	response := RepositoryResponse{}
	collection, isError, errorMessage := getMongoConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		isError = false
		var data []bson.M
		err := collection.Find(bson.M{}).All(&data)

		take := len(data)

		if request.Extras["take"] != nil {
			take, _ = strconv.Atoi(request.Extras["take"].(string))
		}

		skip := 0

		if request.Extras["skip"] != nil {
			skip, _ = strconv.Atoi(request.Extras["skip"].(string))
		}

		if len(data) == 0 {
			response.IsSuccess = true
			response.Message = "No objects found in Mongo"
			var emptyMap map[string]interface{}
			emptyMap = make(map[string]interface{})
			byte, _ := json.Marshal(emptyMap)
			response.GetResponseWithBody(byte)
		}

		for index, _ := range data {
			if request.Controls.SendMetaData == "false" {
				delete(data[index], "__osHeaders")
			}
			delete(data[index], "_id")
		}

		byteValue, errMarshal := json.Marshal(data[skip:(skip + take)])

		if errMarshal != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting values for all objects in mongo" + err.Error())
		} else {
			response.IsSuccess = true
			response.GetResponseWithBody(byteValue)
			response.Message = "Successfully retrieved values for all objects in mongo"
			request.Log(response.Message)
		}

	}
	return response
}

func (repository MongoRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Get Search not implemented in Mongo Db repository")
	return getDefaultNotImplemented()
}

func (repository MongoRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-QUERY!")
	response := RepositoryResponse{}
	queryType := request.Body.Query.Type

	switch queryType {
	case "Query":
		if request.Body.Query.Parameters != "*" {
			request.Log("Support for SQL Query not implemented in Mongo Db repository")
			return getDefaultNotImplemented()
		} else {
			return repository.GetAll(request)
		}
	default:
		request.Log(queryType + " not implemented in Mongo Db repository")
		return getDefaultNotImplemented()
	}

	return response
}

func (repository MongoRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-BY-KEY")
	response := RepositoryResponse{}
	collection, isError, errorMessage := getMongoConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		value := getNoSqlKey(request)

		var data map[string]interface{}
		err := collection.Find(bson.M{"_id": value}).One(&data)

		if err != nil {
			fmt.Println(err.Error())
		}

		if len(data) == 0 {
			response.IsSuccess = true
			response.Message = "No objects found in Mongo"
			var emptyMap map[string]interface{}
			emptyMap = make(map[string]interface{})
			byte, _ := json.Marshal(emptyMap)
			response.GetResponseWithBody(byte)
		}

		if request.Controls.SendMetaData == "false" {
			delete(data, "__osHeaders")
		}
		delete(data, "_id")

		byteValue, errMarshal := json.Marshal(data)
		if errMarshal != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error getting value for a single object in mongo" + errMarshal.Error())
		} else {
			response.IsSuccess = true
			response.GetResponseWithBody(byteValue)
			response.Message = "Successfully retrieved value for a single object in mongo"
			request.Log(response.Message)
		}
	}

	return response
}

func (repository MongoRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-MULTIPLE")
	response := RepositoryResponse{}
	collection, isError, errorMessage := getMongoConnection(request)
	var idData map[string]interface{}
	idData = make(map[string]interface{})
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		isError = false
		if isError == true {
			response.IsSuccess = false
			request.Log("Error inserting multiple objects in Mongo : " + errorMessage)
			response.GetErrorResponse("Error inserting multiple objects in Mongo" + errorMessage)
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted multiple objects in Mongo"
			request.Log(response.Message)
		}

		for i := 0; i < len(request.Body.Objects); i++ {
			key := getMongoDBRecordID(request, request.Body.Objects[i])

			if key == "" {
				continue
			}

			request.Body.Objects[i]["_id"] = request.Controls.Namespace + "." + request.Controls.Class + "." + key
			request.Body.Objects[i][request.Body.Parameters.KeyProperty] = key
			idData[strconv.Itoa(i)] = key

			err := collection.Insert(bson.M(request.Body.Objects[i]))

			if err != nil {
				response.IsSuccess = false
				response.GetErrorResponse("Error inserting many object in mongo" + err.Error())
			} else {
				response.IsSuccess = true
				response.Message = "Successfully inserted many object in Mongo"
				request.Log(response.Message)
			}
		}

	}
	var DataMap []map[string]interface{}
	DataMap = make([]map[string]interface{}, 1)
	var actualInput map[string]interface{}
	actualInput = make(map[string]interface{})
	actualInput["ID"] = idData
	DataMap[0] = actualInput
	response.Data = DataMap
	return response
}

func (repository MongoRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-SINGLE")
	response := RepositoryResponse{}
	collection, isError, errorMessage := getMongoConnection(request)
	key := getMongoDBRecordID(request, nil)
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else if key != "" {

		request.Body.Object[request.Body.Parameters.KeyProperty] = key
		request.Body.Object["_id"] = request.Controls.Namespace + "." + request.Controls.Class + "." + key

		err := collection.Insert(bson.M(request.Body.Object))
		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error inserting one object in mongo" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted one object in Mongo"
			request.Log(response.Message)
		}
	}
	var Data []map[string]interface{}
	Data = make([]map[string]interface{}, 1)
	var actualData map[string]interface{}
	actualData = make(map[string]interface{})
	actualData["ID"] = key
	Data[0] = actualData
	response.Data = Data
	return response
}

func (repository MongoRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-MULTIPLE")
	response := RepositoryResponse{}
	collection, isError, errorMessage := getMongoConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		isError = false
		for _, obj := range request.Body.Objects {
			key := request.Body.Parameters.KeyProperty
			value := obj[request.Body.Parameters.KeyProperty]

			collection.UpdateAll(bson.M{key: value}, bson.M{"$set": obj})
			if isError == true {
				response.IsSuccess = false
				request.Log("Error updating objects in Mongo : " + key + ", " + errorMessage)
				response.GetErrorResponse("Error updating multiple objects in Mongo  because no match was found!" + errorMessage)
			} else {
				response.IsSuccess = true
				response.Message = "Successfully updating multiple objects in Mongo "
				request.Log(response.Message)
			}
		}

	}

	return response
}

func (repository MongoRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-SINGLE")
	response := RepositoryResponse{}
	collection, isError, errorMessage := getMongoConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		key := request.Body.Parameters.KeyProperty
		value := request.Body.Object[request.Body.Parameters.KeyProperty]

		err := collection.Update(bson.M{key: value}, bson.M{"$set": request.Body.Object})
		if err != nil {
			response.IsSuccess = false
			request.Log("Error updating object in Mongo  : " + key + ", " + err.Error())
			response.GetErrorResponse("Error updating one object in Mongo because no match was found!" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully updating one object in Mongo "
			request.Log(response.Message)
		}

	}

	return response
}

func (repository MongoRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-MULTIPLE")
	response := RepositoryResponse{}
	collection, isError, errorMessage := getMongoConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		for _, obj := range request.Body.Objects {
			key := request.Body.Parameters.KeyProperty
			value := obj[request.Body.Parameters.KeyProperty]
			collection.Remove(bson.M{key: value})
			if isError == true {
				response.IsSuccess = false
				request.Log("Error deleting one object in Mongo : " + errorMessage)
				response.GetErrorResponse("Error deleting one object in Mongo" + errorMessage)
			} else {
				response.IsSuccess = true
				response.Message = "Successfully deleting one object in mongo"
				request.Log(response.Message)
			}
		}

	}

	return response
}

func (repository MongoRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-SINGLE")
	response := RepositoryResponse{}
	collection, isError, errorMessage := getMongoConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		key := request.Body.Parameters.KeyProperty
		value := request.Body.Object[request.Body.Parameters.KeyProperty]

		err := collection.Remove(bson.M{key: value})
		if err != nil {
			response.IsSuccess = false
			request.Log("Error deleting object in Mongo  : " + err.Error())
			response.GetErrorResponse("Error deleting one object in Mongo because no match was found!" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully deleted one object in Mongo"
			request.Log(response.Message)
		}
	}

	return response
}

func (repository MongoRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	request.Log("Starting SPECIAL!")
	queryType := request.Body.Special.Type

	switch queryType {
	case "getFields":
		request.Log("Starting GET-FIELDS sub routine!")
		fieldsInByte := executeMongoGetFields(request)
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
		fieldsInByte := executeMongoGetClasses(request)
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
		fieldsInByte := executeMongoGetNamespaces(request)
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
		fieldsInByte := executeMongoGetSelectedFields(request)
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
	}

	return response
}

func (repository MongoRepository) Test(request *messaging.ObjectRequest) {

}

// HELPER functions

func getMongoDBRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
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
		collection, isError, _ := getCustomMongoConnection(request, getSQLnamespace(request), "domainClassAttributes")
		if isError {
			returnID = ""
			request.Log("Connecting to MongoDB Failed!")
		} else {
			//read Attributes table
			key := request.Controls.Class

			var data map[string]interface{}
			err := collection.Find(bson.M{"_id": key}).One(&data)
			fmt.Println(data)
			if err != nil {
				request.Log("This is a freshly created Class. Inserting new Class record.")
				var ObjectBody map[string]interface{}
				ObjectBody = make(map[string]interface{})
				ObjectBody["_id"] = request.Controls.Class
				ObjectBody["maxCount"] = "1"
				ObjectBody["version"] = uuid.NewV1().String()
				err = collection.Insert(bson.M(ObjectBody))
				if err != nil {
					request.Log("Inserting New DomainClassAttributes failed")
					returnID = ""
				} else {
					returnID = "1"
				}
			} else {
				var UpdatedCount int

				for fieldName, fieldvalue := range data {
					if strings.ToLower(fieldName) == "maxcount" {
						UpdatedCount, _ = strconv.Atoi(fieldvalue.(string))
						UpdatedCount++
						returnID = strconv.Itoa(UpdatedCount)
						break
					}
				}

				//update the table
				//save to attributes table
				data["maxCount"] = returnID
				data["version"] = uuid.NewV1().String()
				err := collection.Update(bson.M{"_id": key}, bson.M{"$set": data})
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

//SUB FUNCTIONS
//Functions from SPECIAL and QUERY

func executeMongoGetFields(request *messaging.ObjectRequest) (returnByte []byte) {
	collection, isError, _ := getMongoConnection(request)

	if isError == true {
		returnByte = []byte("Error getting values for all objects in mongo")
	} else {
		isError = false
		var data []bson.M
		err := collection.Find(bson.M{}).All(&data)
		if err != nil {
			request.Log("Error getting data from Mongo")
			returnByte = []byte("Error getting data from Mongo")
		} else {

			var fields []string
			fields = make([]string, len(data[0]))
			index := 0
			for key, _ := range data[0] {
				fields[index] = key
				index++
			}

			byteValue, errMarshal := json.Marshal(fields)
			if errMarshal != nil {
				returnByte = []byte("Error getting values for all objects in mongo")
				request.Log("Error getting values for all objects in mongo")
			} else {
				returnByte = byteValue
				request.Log("Successfully retrieved values for all objects in mongo")
			}
		}

	}
	return
}

func executeMongoGetClasses(request *messaging.ObjectRequest) (returnByte []byte) {

	session, err := mgo.Dial(request.Configuration.ServerConfiguration["MONGO"]["Url"])
	if err != nil {
		returnByte = []byte("Error getting values for all objects in mongo")
	} else {
		db := dbUse.Database{}
		db.Session = session
		db.Name = getSQLnamespace(request)
		data, err := db.CollectionNames()
		if err != nil {
			request.Log("Error getting data from Mongo")
			returnByte = []byte("Error getting data from Mongo")
		} else {

			var newData []string
			newData = make([]string, (len(data) - 1))

			index := 0
			for _, value := range data {
				if value != "system.indexes" {
					newData[index] = value
					index++
				}
			}

			byteValue, errMarshal := json.Marshal(newData)
			if errMarshal != nil {
				returnByte = []byte("Error getting values for all objects in mongo")
				request.Log("Error getting values for all objects in mongo")
			} else {
				returnByte = byteValue
				request.Log("Successfully retrieved values for all objects in mongo")
			}
		}

	}
	return
}

func executeMongoGetNamespaces(request *messaging.ObjectRequest) (returnByte []byte) {

	session, err := mgo.Dial(request.Configuration.ServerConfiguration["MONGO"]["Url"])
	if err != nil {
		returnByte = []byte("Error getting values for all objects in mongo")
	} else {

		data, err := session.DatabaseNames()
		if err != nil {
			request.Log("Error getting data from Mongo")
			returnByte = []byte("Error getting data from Mongo")
		} else {

			byteValue, errMarshal := json.Marshal(data)
			if errMarshal != nil {
				returnByte = []byte("Error getting values for all objects in mongo")
				request.Log("Error getting values for all objects in mongo")
			} else {
				returnByte = byteValue
				request.Log("Successfully retrieved values for all objects in mongo")
			}
		}

	}
	return
}

func executeMongoGetSelectedFields(request *messaging.ObjectRequest) (returnByte []byte) {
	collection, isError, _ := getMongoConnection(request)

	if isError == true {
		returnByte = []byte("Error getting values for all objects in mongo")
	} else {
		isError = false
		var data []bson.M
		err := collection.Find(bson.M{}).All(&data)
		if err != nil {
			request.Log("Error getting data from Mongo")
			returnByte = []byte("Error getting data from Mongo")
		} else {
			request.Log("Requested Field List : " + request.Body.Special.Parameters)
			if request.Body.Special.Parameters == "*" {
				request.Log("All fields requested")
				byteValue, errMarshal := json.Marshal(data)
				if errMarshal != nil {
					returnByte = []byte("Error getting values for all objects in mongo")
					request.Log("Error getting values for all objects in mongo")
				} else {
					returnByte = byteValue
					request.Log("Successfully retrieved values for all objects in mongo")
				}
			} else {
				requestedFields := strings.Split(request.Body.Special.Parameters, " ")

				var allRecords []map[string]interface{}
				allRecords = make([]map[string]interface{}, len(data))

				for key, value := range data {

					var temp map[string]interface{}
					temp = make(map[string]interface{})

					for field, dataValue := range value {
						for _, requestField := range requestedFields {
							if field == requestField {
								temp[field] = dataValue
								continue
							}
						}
					}

					allRecords[key] = temp

				}

				byteValue, errMarshal := json.Marshal(allRecords)
				if errMarshal != nil {
					returnByte = []byte("Error getting values for all objects in mongo")
					request.Log("Error getting values for all objects in mongo")
				} else {
					returnByte = byteValue
					request.Log("Successfully retrieved values for all objects in mongo")
				}

			}

		}

	}
	return
}

func (repository MongoRepository) ClearCache(request *messaging.ObjectRequest) {
}
