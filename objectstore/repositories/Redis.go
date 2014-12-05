package repositories

import (
	"duov6.com/objectstore/messaging"
	"github.com/xuyu/goredis"
)

type RedisRepository struct {
}

func (repository RedisRepository) GetRepositoryName() string {
	return "Redis"
}

func getRedisConnection(request *messaging.ObjectRequest) (client *goredis.Redis, isError bool, errorMessage string) {

	isError = false

	host := request.Configuration.ServerConfiguration["REDIS"]["Host"]
	port := request.Configuration.ServerConfiguration["REDIS"]["Port"]
	client, err := goredis.DialURL("tcp://@" + host + ":" + port + "/0?timeout=10s&maxidle=1")
	if err != nil {
		isError = true
		errorMessage = err.Error()
		request.Log("Error! Can't connect to server!error")

	}
	request.Log("Reusing existing GoRedis connection")
	return
}

func (repository RedisRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("GetAll not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("GetSearch not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("GetQuery not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	client, isError, errorMessage := getRedisConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		key := getNoSqlKey(request) //request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id

		//client := getConnection()
		value, err := client.Get(key)

		if err != nil {
			response.IsSuccess = false
			request.Log("Error getting value by key for object in Redis : " + key + ", " + err.Error())
			response.GetErrorResponse("Error getting value by key for one object in Redis" + err.Error())
		}
		//convert ASCII output to string
		//result := string(value[:])
		if err != nil {
			response.IsSuccess = false
			request.Log("Error getting value by key for object in Redis : " + key + ", " + err.Error())
			response.GetErrorResponse("Error getting value by key for one object in Redis" + err.Error())
		} else {
			response.IsSuccess = true
			response.GetResponseWithBody(value)
			response.Message = "Successfully retrieved one object in Redis"
			request.Log(response.Message)
		}

	}
	return response
}

func (repository RedisRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return setManyRedis(request)
}

func (repository RedisRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	return setOneRedis(request)
}

func setOneRedis(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	client, isError, errorMessage := getRedisConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id

		value := getStringByObject(request, request.Body.Object)

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

	return response
}

func setManyRedis(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	client, isError, errorMessage := getRedisConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		isError := false

		for _, object := range request.Body.Objects {
			key := getNoSqlKeyById(request, object)

			value := getStringByObject(request, object)
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

	return response
}

func (repository RedisRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return setManyRedis(request)
}

func (repository RedisRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	return setOneRedis(request)
}

func (repository RedisRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	client, isError, errorMessage := getRedisConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {

		isError := false

		for _, object := range request.Body.Objects {
			key := getNoSqlKeyById(request, object)

			//value := getStringByObject(request, object)
			//err := client.Set(key, value, 0, 0, false, false)
			reply, err := client.ExecuteCommand("DEL", key)
			err2 := reply.OKValue()
			if err != nil {
				isError = true
				errorMessage = err.Error()
				response.IsSuccess = false
				request.Log("Error deleting object in Redis!" + err2.Error())
				response.GetErrorResponse("Error deleting object in Redis!" + err2.Error())
				break
			} else {
				response.IsSuccess = true
				response.Message = "Successfully deleted multiple objects in Redis"
				request.Log("Successfully deleted all objects in Redis!")
			}
		}

		if isError == true {
			response.IsSuccess = false
			request.Log("Error deleting multiple objects in Redis : " + errorMessage)
			response.GetErrorResponse("Error deleting multiple objects in Redis" + errorMessage)
		} else {
			response.IsSuccess = true
			response.Message = "Successfully deleted multiple objects in Redis"
			request.Log(response.Message)
		}
	}

	return response
}

func (repository RedisRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	client, isError, errorMessage := getRedisConnection(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id

		//value := getStringByObject(request, request.Body.Object)

		isAvailable, err := client.Exists(key)
		if err != nil {
			response.IsSuccess = false
			request.Log("Error deleting object in Redis : " + key + ", " + err.Error())
			response.GetErrorResponse("Error deleting one object in Redis" + err.Error())
		}
		if isAvailable {
			reply, err := client.ExecuteCommand("DEL", key)
			err2 := reply.OKValue()
			if err != nil {
				response.IsSuccess = false
				request.Log("Error deleting object in Redis!" + err2.Error())
				response.GetErrorResponse("Error deleting object in Redis!" + err2.Error())
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
