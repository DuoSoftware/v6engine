package repositories

import (
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	//	"reflect"
)

type MongoRepository struct {
}

func (repository MongoRepository) GetRepositoryName() string {
	return "Mongo DB"
}

func getMongoConnection(request *messaging.ObjectRequest) (client *mgo.Collection, isError bool, errorMessage string) {

	isError = false

	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		isError = false
		errorMessage = err.Error()
		request.Log("Mongo connection initilizing failed!")
	}
	client = session.DB(request.Controls.Namespace).C(request.Controls.Class)
	request.Log("Reusing existing Mongo connection")
	return
}

func (repository MongoRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	collection, isError, errorMessage := getMongoConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		isError = false
		var data []bson.M
		err := collection.Find(bson.M{}).All(&data)

		byteValue, errMarshal := json.Marshal(data)
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
	request.Log("DeleteMultiple not implemented in Mongo Db repository")
	return getDefaultNotImplemented()
}

func (repository MongoRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("DeleteMultiple not implemented in Mongo Db repository")
	return getDefaultNotImplemented()
}

func (repository MongoRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	collection, isError, errorMessage := getMongoConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		//key := request.Body.Parameters.KeyProperty
		value := getNoSqlKey(request)
		fmt.Println(value)

		var data map[string]interface{}
		err := collection.Find(bson.M{"_id": value}).One(&data)

		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println(data)

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
	response := RepositoryResponse{}

	collection, isError, errorMessage := getMongoConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		isError := false
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
			for key := range request.Body.Objects[i] {
				request.Body.Objects[i]["_id"] = getNoSqlKeyById(request, request.Body.Objects[i])
				fmt.Println(key)
			}
		}

		for i := 0; i < len(request.Body.Objects); i++ {
			err := collection.Insert(bson.M(request.Body.Objects[i]))
			if err != nil {
				response.IsSuccess = false
				//request.Log("Error inserting/updating object in Mongo : " + key + ", " + err.Error())
				response.GetErrorResponse("Error inserting many object in mongo" + err.Error())
			} else {
				response.IsSuccess = true
				response.Message = "Successfully inserted many object in Mongo"
				request.Log(response.Message)
			}
		}

	}

	return response
}

func (repository MongoRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	collection, isError, errorMessage := getMongoConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		key := getNoSqlKey(request)
		request.Body.Object["_id"] = key

		for key := range request.Body.Object {
			request.Log(key)
		}

		err := collection.Insert(bson.M(request.Body.Object))
		//collection.Upsert(bson.M{"_id": "500"}, update)
		if err != nil {
			response.IsSuccess = false
			//request.Log("Error inserting/updating object in Mongo : " + key + ", " + err.Error())
			response.GetErrorResponse("Error inserting one object in mongo" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted one object in Mongo"
			request.Log(response.Message)
		}
	}
	return response
}

func (repository MongoRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	collection, isError, errorMessage := getMongoConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		isError = false

		key := request.Body.Parameters.KeyProperty
		value := request.Body.Object[request.Body.Parameters.KeyProperty]

		collection.UpdateAll(bson.M{key: value}, bson.M{"$set": request.Body.Object})
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

	return response
}

func (repository MongoRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
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
	response := RepositoryResponse{}
	collection, isError, errorMessage := getMongoConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		key := request.Body.Parameters.KeyProperty
		value := request.Body.Object[request.Body.Parameters.KeyProperty]
		collection.RemoveAll(bson.M{key: value})
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

	return response
}

func (repository MongoRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
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
	request.Log("DeleteMultiple not implemented in Mongo Db repository")
	return getDefaultNotImplemented()
}

func (repository MongoRepository) Test(request *messaging.ObjectRequest) {

}
