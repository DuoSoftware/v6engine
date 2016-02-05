package repositories

import (
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"fmt"
	"github.com/xuyu/goredis"
	"strconv"
)

func getRedisConnection(request *messaging.ObjectRequest) (client *goredis.Redis, isError bool, errorMessage string) {

	isError = false
	client, err := goredis.DialURL("tcp://@" + request.Configuration.ServerConfiguration["REDIS"]["Host"] + ":" + request.Configuration.ServerConfiguration["REDIS"]["Port"] + "/0?timeout=1s&maxidle=1")
	if err != nil {
		isError = true
		errorMessage = err.Error()
		request.Log("Error! Can't connect to server!")
		return nil, isError, errorMessage

	}
	if client == nil {
		return nil, true, "No REDIS Host Found!"
	}
	return
}

func GetByKey(request *messaging.ObjectRequest) (output []byte) {
	client, isError, errorMessage := getRedisConnection(request)

	if isError == true {
		fmt.Println(errorMessage)
	} else {
		key := getNoSqlKey(request)
		result, err := client.Get(key)
		if err != nil {
			fmt.Println("ERROR : " + err.Error())
		} else {
			if checkEmptyByteArray(result) {
				result = nil
			}
			output = result
		}
		client.ClosePool()
	}

	return
}

func SetOneRedis(request *messaging.ObjectRequest, data map[string]interface{}) (err error) {
	client, isError, errorMessage := getRedisConnection(request)
	if isError == true {
		fmt.Println(errorMessage)
	} else {
		var ttl int
		ttl, _ = strconv.Atoi(request.Configuration.ServerConfiguration["REDIS"]["TTL"])
		key := getNoSqlKeyById(request, data)
		value := getStringByObject(data)
		fmt.Println(ttl)
		err = client.Set(key, value, ttl, 0, false, false)
		client.ClosePool()
	}

	return
}

func SetManyRedis(request *messaging.ObjectRequest, data []map[string]interface{}) (err error) {
	client, isError, errorMessage := getRedisConnection(request)
	if isError == true {
		fmt.Println(errorMessage)
	} else {
		var ttl int
		ttl, _ = strconv.Atoi(request.Configuration.ServerConfiguration["REDIS"]["TTL"])
		fmt.Println(ttl)
		for _, obj := range data {
			key := getNoSqlKeyById(request, obj)
			value := getStringByObject(obj)
			err = client.Set(key, value, ttl, 0, false, false)
			if err != nil {
				return
			}
		}
		client.ClosePool()
	}

	return
}

func getNoSqlKeyById(request *messaging.ObjectRequest, obj map[string]interface{}) string {
	key := request.Controls.Namespace + "." + request.Controls.Class + "." + obj[request.Body.Parameters.KeyProperty].(string)
	return key
}

func getNoSqlKey(request *messaging.ObjectRequest) string {
	key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id
	return key
}

func getStringByObject(obj interface{}) string {
	result, err := json.Marshal(obj)
	if err == nil {
		return string(result)
	} else {
		return "{}"
	}
}

func checkEmptyByteArray(input []byte) (status bool) {
	status = false
	if len(input) == 4 || len(input) == 2 || len(input) < 2 {
		status = true
	}
	return
}
