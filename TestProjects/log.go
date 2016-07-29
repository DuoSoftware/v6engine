package main

import (
	"duov6.com/common"
	"duov6.com/gorest"
	"errors"
	"fmt"
	"github.com/xuyu/goredis"
	"net/http"
	"sync"
)

func main() {
	fmt.Println("uehueheu")
	runRestFul()
}

func runRestFul() {
	gorest.RegisterService(new(Auth))

	err := http.ListenAndServe(":8899", gorest.Handle())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

}

type Auth struct {
	gorest.RestService
	verify gorest.EndPoint `method:"GET" path:"/" output:"string"`
}

// func (A Auth) Verify() (output string) {
// 	err := SetOneRedis()
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		output = err.Error()
// 		A.ResponseBuilder().SetResponseCode(401).WriteAndOveride([]byte(err.Error()))
// 	} else {
// 		fmt.Println("Inserted Successfully")
// 		output = "Inserted Successfully"
// 		A.ResponseBuilder().SetResponseCode(200).WriteAndOveride([]byte("Inserted Successfully"))
// 	}

// 	return
// }

func (A Auth) Verify() (output string) {
	fmt.Println("yay")
	addToMap()
	fmt.Println(len(m))
	A.ResponseBuilder().SetResponseCode(200).WriteAndOveride([]byte("Inserted Successfully"))
	return
}

var RedisConnection *goredis.Redis

func GetConnection() (client *goredis.Redis, err error) {

	host := "localhost"
	port := "6379"

	if RedisConnection == nil {
		client, err = goredis.DialURL("tcp://@" + host + ":" + port + "/" + "10" + "?timeout=300s&maxidle=300")
		if err != nil {
			return nil, err
		} else {
			if client == nil {
				return nil, errors.New("Connection to REDIS Failed!")
			}
			RedisConnection = client
		}

	} else {
		if err = RedisConnection.Ping(); err != nil {
			RedisConnection = nil
			client, err = goredis.DialURL("tcp://@" + host + ":" + port + "/" + "10" + "?timeout=60s&maxidle=60")
			if err != nil {
				return nil, err
			} else {
				if client == nil {
					return nil, errors.New("Connection to REDIS Failed!")
				}
				RedisConnection = client
			}
		} else {
			client = RedisConnection
		}
	}

	return
}

func SetOneRedis() (err error) {
	client, isError := GetConnection()
	if isError == nil {
		key := common.GetGUID()
		value := "{\"name\": \"DuoAuth\",\"version\": \"6.0.22-a\",\"Change Log\":\"Added doc cache!\",\"author\": {\"name\": \"Duo Software\",\"url\": \"http://www.duosoftware.com/\"},\"repository\": {\"type\": \"git\",\"url\": \"https://github.com/DuoSoftware/v6engine/\"}}"
		err = client.Set(key, value, 0, 0, false, false)

		if err != nil {
			fmt.Println("Inserted One Record to Cache!")
		}
	}

	return
}

var m map[string]interface{}
var lock = sync.RWMutex{}

func addToMap() {

	if m == nil {
		m = make(map[string]interface{})
	}

	lock.Lock()
	defer lock.Unlock()
	guid := common.GetGUID()
	m[guid] = guid
}

func Read() {
	lock.RLock()
	defer lock.RUnlock()
	_ = m["a"]
}
