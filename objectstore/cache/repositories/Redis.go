package repositories

import (
	"duov6.com/objectstore/messaging"
	"errors"
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
			if !checkEmptyByteArray(result) {
				fmt.Println("Retrieved from Cache!")
			}
		}
		client.ClosePool()
	}

	return
}

func GetSearch(request *messaging.ObjectRequest) (output []byte) {
	client, isError, errorMessage := getRedisConnection(request)

	if isError == true {
		fmt.Println(errorMessage)
	} else {
		key := getSearchResultKey(request)
		result, err := client.Get(key)
		if err != nil {
			fmt.Println("ERROR : " + err.Error())
		} else {
			if checkEmptyByteArray(result) {
				result = nil
			}
			output = result
			if !checkEmptyByteArray(result) {
				fmt.Println("Retrieved from Cache!")
			}
		}
		client.ClosePool()
	}

	return
}

func GetQuery(request *messaging.ObjectRequest) (output []byte) {
	client, isError, errorMessage := getRedisConnection(request)

	if isError == true {
		fmt.Println(errorMessage)
	} else {
		key := getQueryResultKey(request)
		result, err := client.Get(key)
		if err != nil {
			fmt.Println("ERROR : " + err.Error())
		} else {
			if checkEmptyByteArray(result) {
				result = nil
			}
			output = result
			if !checkEmptyByteArray(result) {
				fmt.Println("Retrieved from Cache!")
			}
		}
		client.ClosePool()
	}

	return
}

func SetOneRedis(request *messaging.ObjectRequest, data map[string]interface{}) (err error) {

	client, isError, errorMessage := getRedisConnection(request)
	if isError == true {
		err = errors.New(errorMessage)
	} else {
		var ttl int
		ttl, _ = strconv.Atoi(request.Configuration.ServerConfiguration["REDIS"]["TTL"])
		key := ""
		if request.Body.Parameters.KeyProperty == "" {
			key = getNoSqlKey(request)
		} else {
			key = getNoSqlKeyById(request, data)
		}

		value := getStringByObject(data)
		err = client.Set(key, value, ttl, 0, false, false)

		if err != nil {
			fmt.Println("Inserted One Record to Cache!")
		}

		client.ClosePool()
	}

	_ = ResetSearchResultCache(request)

	return
}

func SetResultRedis(request *messaging.ObjectRequest, data interface{}) (err error) {
	client, isError, errorMessage := getRedisConnection(request)
	if isError == true {
		err = errors.New(errorMessage)
	} else {
		var ttl int
		ttl, _ = strconv.Atoi(request.Configuration.ServerConfiguration["REDIS"]["TTL"])
		key := getSearchResultKey(request)
		value := getStringByObject(data)
		err = client.Set(key, value, ttl, 0, false, false)

		if err != nil {
			fmt.Println("Inserted Search Result Set to Cache!")
		}
		client.ClosePool()
	}

	return
}

func SetQueryRedis(request *messaging.ObjectRequest, data interface{}) (err error) {
	client, isError, errorMessage := getRedisConnection(request)
	if isError == true {
		err = errors.New(errorMessage)
	} else {
		var ttl int
		ttl, _ = strconv.Atoi(request.Configuration.ServerConfiguration["REDIS"]["TTL"])
		key := getQueryResultKey(request)
		value := getStringByObject(data)
		err = client.Set(key, value, ttl, 0, false, false)

		if err != nil {
			fmt.Println("Inserted Query Result Set to Cache!")
		}
		client.ClosePool()
	}

	return
}

func SetManyRedis(request *messaging.ObjectRequest, data []map[string]interface{}) (err error) {
	client, isError, errorMessage := getRedisConnection(request)
	if isError == true {
		err = errors.New(errorMessage)
	} else if request.Body.Parameters.KeyProperty != "" {
		var ttl int
		ttl, _ = strconv.Atoi(request.Configuration.ServerConfiguration["REDIS"]["TTL"])
		for _, obj := range data {
			key := getNoSqlKeyById(request, obj)
			value := getStringByObject(obj)
			err = client.Set(key, value, ttl, 0, false, false)
			if err != nil {
				client.ClosePool()
				return
			}
		}
		if err != nil {
			fmt.Println("Inserted Many Records to Cache!")
		}

		client.ClosePool()
	}

	_ = ResetSearchResultCache(request)

	return
}

func RemoveOneRedis(request *messaging.ObjectRequest, data map[string]interface{}) (err error) {

	client, isError, errorMessage := getRedisConnection(request)
	if isError == true {
		err = errors.New(errorMessage)
	} else {

		fmt.Println("------------")
		fmt.Println(request.Body.Parameters.KeyProperty)
		fmt.Println(request)
		fmt.Println("------------")

		key := ""
		if request.Body.Parameters.KeyProperty == "" || request.Controls.Id != "" {
			key = getNoSqlKey(request)
		} else {
			key = getNoSqlKeyById(request, data)
		}
		reply, _ := client.ExecuteCommand("DEL", key)

		_ = reply.OKValue()

		client.ClosePool()
	}

	_ = ResetSearchResultCache(request)

	return
}

func RemoveManyRedis(request *messaging.ObjectRequest, data []map[string]interface{}) (err error) {
	client, isError, errorMessage := getRedisConnection(request)
	if isError == true {
		err = errors.New(errorMessage)
	} else if request.Body.Parameters.KeyProperty != "" {
		for _, obj := range data {
			key := getNoSqlKeyById(request, obj)
			reply, _ := client.ExecuteCommand("DEL", key)
			_ = reply.OKValue()
		}

		client.ClosePool()
	}

	_ = ResetSearchResultCache(request)

	return
}

func ResetSearchResultCache(request *messaging.ObjectRequest) (err error) {
	client, isError, errorMessage := getRedisConnection(request)
	if isError == true {
		err = errors.New(errorMessage)
	} else {
		namespace := request.Controls.Namespace
		class := request.Controls.Class

		pattern := namespace + ":" + class + ":" + "keyword=*"

		if keySet, err := client.Keys(pattern); err == nil {
			for _, keyValue := range keySet {
				reply, err := client.ExecuteCommand("DEL", keyValue)
				err = reply.OKValue()
				if err != nil {
					client.ClosePool()
					return err
				}
			}
		}

		queryPattern := namespace + ":" + class + ":Query:*"

		if keySet, err := client.Keys(queryPattern); err == nil {
			for _, keyValue := range keySet {
				reply, err := client.ExecuteCommand("DEL", keyValue)
				err = reply.OKValue()
				if err != nil {
					client.ClosePool()
					return err
				}
			}
		}

		if err != nil {
			fmt.Println("Resetted the pattern Key Set in Cache!")
		}

		client.ClosePool()
	}

	return err
}
