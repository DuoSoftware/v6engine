package repositories

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/term"
	"errors"
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
		term.Write(errorMessage, term.Debug)
	} else {
		key := getNoSqlKey(request)
		result, err := client.Get(key)
		if err != nil {
			term.Write("ERROR : "+err.Error(), term.Debug)
		} else {
			if checkEmptyByteArray(result) {
				result = nil
			}
			output = result
			if !checkEmptyByteArray(result) {
				term.Write("Retrieved from Cache!", term.Debug)
			}
		}
		client.ClosePool()
	}

	return
}

func GetSearch(request *messaging.ObjectRequest) (output []byte) {
	client, isError, errorMessage := getRedisConnection(request)

	if isError == true {
		term.Write(errorMessage, term.Debug)
	} else {
		key := getSearchResultKey(request)
		result, err := client.Get(key)
		if err != nil {
			term.Write("ERROR : "+err.Error(), term.Debug)
		} else {
			if checkEmptyByteArray(result) {
				result = nil
			}
			output = result
			if !checkEmptyByteArray(result) {
				term.Write("Retrieved from Cache!", term.Debug)
			}
		}
		client.ClosePool()
	}

	return
}

func GetQuery(request *messaging.ObjectRequest) (output []byte) {
	client, isError, errorMessage := getRedisConnection(request)

	if isError == true {
		term.Write(errorMessage, term.Debug)
	} else {
		key := getQueryResultKey(request)
		result, err := client.Get(key)
		if err != nil {
			term.Write("ERROR : "+err.Error(), term.Debug)
		} else {
			if checkEmptyByteArray(result) {
				result = nil
			}
			output = result
			if !checkEmptyByteArray(result) {
				term.Write("Retrieved from Cache!", term.Debug)
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
			term.Write("Inserted One Record to Cache!", term.Debug)
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
			term.Write("Inserted Search Result Set to Cache!", term.Debug)
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
			term.Write("Inserted Query Result Set to Cache!", term.Debug)
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
			term.Write("Inserted Many Records to Cache!", term.Debug)
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

		term.Write("------------", term.Debug)
		term.Write(request.Body.Parameters.KeyProperty, term.Debug)
		term.Write(request, term.Debug)
		term.Write("------------", term.Debug)

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
			term.Write("Resetted the pattern Key Set in Cache!", term.Debug)
		}

		client.ClosePool()
	}

	return err
}
