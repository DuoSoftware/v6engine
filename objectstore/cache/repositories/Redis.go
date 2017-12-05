package repositories

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/term"
	"errors"
	"fmt"
	"github.com/xuyu/goredis"
	"strconv"
)

func getRedisConnection(request *messaging.ObjectRequest, database int) (client *goredis.Redis, isError bool, errorMessage string) {
	client, err := GetConnection(request, database)
	if err != nil {
		isError = true
		errorMessage = err.Error()
	}
	return
}

var RedisCacheConnection []*goredis.Redis

func GetConnection(request *messaging.ObjectRequest, database int) (client *goredis.Redis, err error) {

	if RedisCacheConnection == nil {
		RedisCacheConnection = make([]*goredis.Redis, 16)
	}

	host := request.Configuration.ServerConfiguration["REDIS"]["Host"]
	port := request.Configuration.ServerConfiguration["REDIS"]["Port"]
	password := request.Configuration.ServerConfiguration["REDIS"]["Password"]

	urlStart := "tcp://"
	if password != "" {
		urlStart += "auth:" + password
	}
	urlStart += "@"

	if RedisCacheConnection[database] == nil {
		client, err = goredis.DialURL(urlStart + host + ":" + port + "/" + strconv.Itoa(database) + "?timeout=60s&maxidle=60")
		if err != nil {
			return nil, err
		} else {
			if client == nil {
				return nil, errors.New("Connection to REDIS Failed!")
			}
			RedisCacheConnection[database] = client
		}

	} else {
		if err = RedisCacheConnection[database].Ping(); err != nil {
			RedisCacheConnection[database] = nil
			client, err = goredis.DialURL(urlStart + host + ":" + port + "/" + strconv.Itoa(database) + "?timeout=60s&maxidle=60")
			if err != nil {
				return nil, err
			} else {
				if client == nil {
					return nil, errors.New("Connection to REDIS Failed!")
				}
				RedisCacheConnection[database] = client
			}
		} else {
			client = RedisCacheConnection[database]
		}
	}

	return
}

func GetByKey(request *messaging.ObjectRequest, database int) (output []byte) {
	client, isError, errorMessage := getRedisConnection(request, database)

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
		//client.ClosePool()
	}

	return
}

func GetSearch(request *messaging.ObjectRequest, database int) (output []byte) {
	client, isError, errorMessage := getRedisConnection(request, database)

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
		//client.ClosePool()
	}

	return
}

func GetQuery(request *messaging.ObjectRequest, database int) (output []byte) {
	client, isError, errorMessage := getRedisConnection(request, database)

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
		//client.ClosePool()
	}

	return
}

func SetOneRedis(request *messaging.ObjectRequest, data map[string]interface{}, database int) (err error) {
	_ = ResetSearchResultCache(request, database)

	client, isError, errorMessage := getRedisConnection(request, database)
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
		//err = client.Set(key, value, 0, 0, false, false)

		if err != nil {
			term.Write("Inserted One Record to Cache!", term.Debug)
		}

		//client.ClosePool()
	}

	return
}

func SetResultRedis(request *messaging.ObjectRequest, data interface{}, database int) (err error) {
	client, isError, errorMessage := getRedisConnection(request, database)
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
		//client.ClosePool()
	}

	return
}

func SetQueryRedis(request *messaging.ObjectRequest, data interface{}, database int) (err error) {
	client, isError, errorMessage := getRedisConnection(request, database)
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
		//client.ClosePool()
	}

	return
}

func SetManyRedis(request *messaging.ObjectRequest, data []map[string]interface{}, database int) (err error) {
	_ = ResetSearchResultCache(request, database)

	client, isError, errorMessage := getRedisConnection(request, database)
	if isError == true {
		err = errors.New(errorMessage)
	} else if request.Body.Parameters.KeyProperty != "" {
		var ttl int
		ttl, _ = strconv.Atoi(request.Configuration.ServerConfiguration["REDIS"]["TTL"])
		for _, obj := range data {
			key := getNoSqlKeyById(request, obj)
			value := getStringByObject(obj)
			err = client.Set(key, value, ttl, 0, false, false)
			//err = client.Set(key, value, 0, 0, false, false)

			if err != nil {
				//client.ClosePool()
				return
			}
		}
		if err != nil {
			term.Write("Inserted Many Records to Cache!", term.Debug)
		}

		//client.ClosePool()
	}

	return
}

func RemoveOneRedis(request *messaging.ObjectRequest, data map[string]interface{}, database int) (err error) {

	client, isError, errorMessage := getRedisConnection(request, database)
	if isError == true {
		err = errors.New(errorMessage)
	} else {

		key := ""
		if request.Body.Parameters.KeyProperty == "" || request.Controls.Id != "" {
			key = getNoSqlKey(request)
		} else {
			key = getNoSqlKeyById(request, data)
		}
		//reply, _ := client.ExecuteCommand("DEL", key)

		//_ = reply.OKValue()

		_, err = client.Expire(key, 0)
	}

	_ = ResetSearchResultCache(request, database)

	return
}

func RemoveManyRedis(request *messaging.ObjectRequest, data []map[string]interface{}, database int) (err error) {
	client, isError, errorMessage := getRedisConnection(request, database)
	if isError == true {
		err = errors.New(errorMessage)
	} else if request.Body.Parameters.KeyProperty != "" {
		for _, obj := range data {
			key := getNoSqlKeyById(request, obj)
			_, err = client.Expire(key, 0)
		}

		//client.ClosePool()
	}

	_ = ResetSearchResultCache(request, database)

	return
}

func ResetSearchResultCache(request *messaging.ObjectRequest, database int) (err error) {
	client, isError, errorMessage := getRedisConnection(request, database)
	if isError == true {
		err = errors.New(errorMessage)
	} else {
		namespace := request.Controls.Namespace
		class := request.Controls.Class

		pattern := namespace + ":" + class + ":" + "keyword=*"

		if keySet, err := client.Keys(pattern); err == nil {
			for _, keyValue := range keySet {
				_, err = client.Expire(keyValue, 0)
				if err != nil {
					//client.ClosePool()
					return err
				}
			}
		}

		queryPattern := namespace + ":" + class + ":Query:*"

		if keySet, err := client.Keys(queryPattern); err == nil {
			for _, keyValue := range keySet {
				//reply, err := client.ExecuteCommand("DEL", keyValue)
				//err = reply.OKValue()
				_, err = client.Expire(keyValue, 0)
				if err != nil {
					//client.ClosePool()
					return err
				}
			}
		}

		keyPattern := namespace + "." + class + ".*"

		if keySet, err := client.Keys(keyPattern); err == nil {
			for _, keyValue := range keySet {
				//reply, err := client.ExecuteCommand("DEL", keyValue)
				//err = reply.OKValue()
				_, err = client.Expire(keyValue, 0)
				if err != nil {
					//client.ClosePool()
					return err
				}
			}
		}

		if err != nil {
			term.Write("Resetted the pattern Key Set in Cache!", term.Debug)
		}

		//client.ClosePool()
	}

	return err
}

func StoreKeyValue(request *messaging.ObjectRequest, key string, value string, database int) (err error) {
	client, err := GetConnection(request, database)
	if err != nil {
		return
	}
	err = client.Set(key, value, 0, 0, false, false)
	return
}

func GetKeyValue(request *messaging.ObjectRequest, key string, database int) (value []byte) {
	client, err := GetConnection(request, database)
	if err != nil {
		return
	}
	value, err = client.Get(key)
	return
}

func GetKeyListPattern(request *messaging.ObjectRequest, key string, database int) (value []string) {
	client, err := GetConnection(request, database)
	if err != nil {
		return
	}
	value, err = client.Keys(key)
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}

func ExistsKeyValue(request *messaging.ObjectRequest, key string, database int) (status bool) {
	client, err := GetConnection(request, database)
	if err != nil {
		return
	}
	status, err = client.Exists(key)
	return
}

func DeleteKey(request *messaging.ObjectRequest, key string, database int) (status bool) {
	client, err := GetConnection(request, database)
	if err != nil {
		return
	}
	status, err = client.Expire(key, 0)
	return
}

func DeletePattern(request *messaging.ObjectRequest, pattern string, database int) (status bool) {
	keys := GetKeyListPattern(request, pattern, database)
	client, err := GetConnection(request, database)
	if err != nil {
		return
	}
	for _, key := range keys {
		_, _ = client.Expire(key, 0)
	}
	status = true
	return
}

func RPush(request *messaging.ObjectRequest, list string, value string, database int) (err error) {
	client, err := GetConnection(request, database)
	if err != nil {
		return
	}
	_, err = client.RPush(list, value)
	return
}

func LPush(request *messaging.ObjectRequest, list string, value string, database int) (err error) {
	client, err := GetConnection(request, database)
	if err != nil {
		return
	}
	_, err = client.LPush(list, value)
	return
}

func GetListLength(request *messaging.ObjectRequest, key string, database int) (length int64) {
	client, err := GetConnection(request, database)
	if err != nil {
		return
	}
	length, _ = client.LLen(key)
	return
}

func RPop(request *messaging.ObjectRequest, key string, database int) (result []byte, err error) {
	client, err := GetConnection(request, database)
	if err != nil {
		return
	}
	result, err = client.RPop(key)
	return
}

func LPop(request *messaging.ObjectRequest, key string, database int) (result []byte, err error) {
	client, err := GetConnection(request, database)
	if err != nil {
		return
	}
	result, err = client.LPop(key)
	return
}

func Flush(request *messaging.ObjectRequest) {

	for x := 0; x < 3; x++ {
		client, err := GetConnection(request, x)
		if err != nil {
			return
		}
		_ = client.FlushDB()
	}

	//Flush Logs
	client, err := GetConnection(request, 8)
	if err != nil {
		return
	}
	_ = client.FlushDB()

}

func LRange(request *messaging.ObjectRequest, key string, database, start, end int) (result []string, err error) {
	client, err := GetConnection(request, database)
	if err != nil {
		return
	}
	result, err = client.LRange(key, start, end)
	return
}

func GetIncrValue(request *messaging.ObjectRequest, key string, database int) (val int64) {
	client, err := GetConnection(request, database)
	if err != nil {
		return
	}
	val, _ = client.Incr(key)
	_, _ = client.Expire(key, 300)
	return
}
