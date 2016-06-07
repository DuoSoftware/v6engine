package keygenerator

import (
	"duov6.com/common"
	"duov6.com/objectstore/keygenerator/drivers"
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xuyu/goredis"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func UpdateCountsToDB() {
	tickCount := 0
	c := time.Tick(1 * time.Minute)
	for _ = range c {
		tickCount++
		if tickCount == 120 {
			if common.GetProcessorUsage() < 20 {
				tickCount = 0
				fmt.Println("Executing KeyGen Sync Routine. ( Every Once 30 Minutes )")
				UpdateKeysInDB()
			} else {
				tickCount = 60
			}
		}
	}
}

func GetIncrementID(request *messaging.ObjectRequest, repository string, amount int) (key string) {
	client, err := GetConnection(request)
	if err != nil {
		RedisConnection = nil
		RedisConnectionTCP = nil
		key = common.GetGUID()
		fmt.Println(err.Error() + " Returning an Unique GUID : " + key)
	} else {
		key = ExecuteKeyGenProcess(client, request, repository, amount)
	}

	return
}

func GetTentativeID(request *messaging.ObjectRequest, repository string, amount int) (key string) {
	client, err := GetConnection(request)
	if err != nil {
		RedisConnection = nil
		RedisConnectionTCP = nil
		key = common.GetGUID()
		fmt.Println(err.Error() + " Returning an Unique GUID : " + key)
	} else {
		key = ExecuteKeyGenProcessForReading(client, request, repository, amount)
	}
	return
}

func ExecuteKeyGenProcess(client *goredis.Redis, request *messaging.ObjectRequest, repository string, amount int) (key string) {
	if status := CheckForKeyGen(request, client); status {
		//Key Available in Database
		// if isLock := CheckKeyGenLock(request, client); isLock {
		// 	for true {
		// 		if isLock = CheckKeyGenLock(request, client); !isLock {
		// 			// Lock is Over
		// 			key = GetKeyGenKey(request, client)
		// 			SetKeyGenTime(request, client)
		// 			return
		// 		}
		// 	}
		// } else {
		key = GetKeyGenKey(request, client)
		//SetKeyGenTime(request, client)
		//	}
	} else {
		if IsLockKey := CheckKeyGenLock(request, client); !IsLockKey {
			LockKeyGen(request, client)
			max := VerifyMaxFromDB(request, repository, amount)
			SetKeyGenKey(request, client, max)
			UnlockKeyGen(request, client)
			//SetKeyGenTime(request, client)
			key = max
		} else {
			for true {
				time.Sleep(1)
				if isLock := CheckKeyGenLock(request, client); !isLock {
					// Lock is Over
					key = GetKeyGenKey(request, client)
					//SetKeyGenTime(request, client)
					return
				}
			}
		}
	}
	return
}

func ExecuteKeyGenProcessForReading(client *goredis.Redis, request *messaging.ObjectRequest, repository string, amount int) (key string) {
	if status := CheckForKeyGen(request, client); status {
		//Key Available in Database
		// if isLock := CheckKeyGenLock(request, client); isLock {
		// 	for true {
		// 		if isLock = CheckKeyGenLock(request, client); !isLock {
		// 			// Lock is Over
		// 			key = GetKeyGenKey(request, client)
		// 			SetKeyGenTime(request, client)
		// 			return
		// 		}
		// 	}
		// } else {
		key = ReadKeyGenKey(request, client)
		//SetKeyGenTime(request, client)
		//	}
	} else {
		if IsLockKey := CheckKeyGenLock(request, client); !IsLockKey {
			LockKeyGen(request, client)
			max := VerifyMaxFromDB(request, repository, amount)
			SetKeyGenKey(request, client, max)
			UnlockKeyGen(request, client)
			//SetKeyGenTime(request, client)
			key = max
		} else {
			for true {
				time.Sleep(1)
				if isLock := CheckKeyGenLock(request, client); !isLock {
					// Lock is Over
					key = ReadKeyGenKey(request, client)
					//SetKeyGenTime(request, client)
					return
				}
			}
		}
	}
	return
}

func VerifyMaxFromDB(request *messaging.ObjectRequest, repository string, count int) (max string) {

	client, _ := GetConnection(request)
	if lock := CheckKeyGenLock(request, client); lock {
		for true {
			if isLock := CheckKeyGenLock(request, client); !isLock {
				max = GetKeyGenKey(request, client)
			}
		}
		return
	}

	fmt.Println("Syncing " + request.Controls.Namespace + ".DomainClassAttributes - " + repository)
	switch repository {
	case "CLOUDSQL":
		var driver drivers.CloudSql
		max = driver.VerifyMaxValueDB(request, count)
		break
	case "ELASTIC":
		var driver drivers.ElasticSearch
		max = driver.VerifyMaxValueDB(request, count)
		break
	default:
		fmt.Println("Error! No such Repository : " + repository + " exists!")
		break
	}
	return
}

var RedisConnection *goredis.Redis

func GetConnection(request *messaging.ObjectRequest) (client *goredis.Redis, err error) {
	host := request.Configuration.ServerConfiguration["REDIS"]["Host"]
	port := request.Configuration.ServerConfiguration["REDIS"]["Port"]
	if RedisConnection == nil {
		//client, err := goredis.Dial(&goredis.DialConfig{"tcp", (host + ":" + port), 1, "", 1 * time.Second, 1})
		client, err = goredis.DialURL("tcp://@" + host + ":" + port + "/5?timeout=60s&maxidle=60")
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
			//client, err := goredis.Dial(&goredis.DialConfig{"tcp", (host + ":" + port), 1, "", 1 * time.Second, 1})
			client, err = goredis.DialURL("tcp://@" + host + ":" + port + "/5?timeout=60s&maxidle=60")
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

var RedisConnectionTCP *goredis.Redis

func GetConnectionTCP(host string, port string) (client *goredis.Redis, err error) {
	if RedisConnectionTCP == nil {
		//client, err := goredis.Dial(&goredis.DialConfig{"tcp", (host + ":" + port), 1, "", 1 * time.Second, 1})
		client, err = goredis.DialURL("tcp://@" + host + ":" + port + "/5?timeout=60s&maxidle=60")
		if err != nil {
			return nil, err
		} else {
			if client == nil {
				return nil, errors.New("Connection to REDIS Failed!")
			}
			RedisConnectionTCP = client
		}
	} else {
		if err = RedisConnectionTCP.Ping(); err != nil {
			RedisConnectionTCP = nil
			//client, err := goredis.Dial(&goredis.DialConfig{"tcp", (host + ":" + port), 1, "", 1 * time.Second, 1})
			client, err = goredis.DialURL("tcp://@" + host + ":" + port + "/5?timeout=60s&maxidle=60")
			if err != nil {
				return nil, err
			} else {
				if client == nil {
					return nil, errors.New("Connection to REDIS Failed!")
				}
				RedisConnectionTCP = client
			}
		} else {
			client = RedisConnectionTCP
		}
	}

	return
}

func CheckForKeyGen(request *messaging.ObjectRequest, client *goredis.Redis) (status bool) {
	incrementKey := "KeyGenKey." + request.Controls.Namespace + "." + request.Controls.Class
	//lockKey := "KeyGenLock." + request.Controls.Namespace + "." + request.Controls.Class
	//timeKey := "KeyGenTime." + request.Controls.Namespace + "." + request.Controls.Class
	status, err := client.Exists(incrementKey)
	if err != nil {
		RedisConnection = nil
		RedisConnectionTCP = nil
		fmt.Println(err.Error())
	}
	return
}

func CheckKeyGenLock(request *messaging.ObjectRequest, client *goredis.Redis) (status bool) {
	status = false
	key := "KeyGenLock." + request.Controls.Namespace + "." + request.Controls.Class
	val, err := client.Get(key)
	if err != nil {
		RedisConnection = nil
		RedisConnectionTCP = nil
		fmt.Println(err.Error())
		return
	}

	if val == nil {
		_ = client.Set(key, "false", 0, 0, false, false)
		return
	}

	err = json.Unmarshal(val, &status)
	if err != nil {
		RedisConnection = nil
		RedisConnectionTCP = nil
		fmt.Println(err.Error())
	}

	return

}

func LockKeyGen(request *messaging.ObjectRequest, client *goredis.Redis) {
	key := "KeyGenLock." + request.Controls.Namespace + "." + request.Controls.Class
	err := client.Set(key, "false", 0, 0, false, false)
	if err != nil {
		RedisConnection = nil
		RedisConnectionTCP = nil
		fmt.Println(err.Error())
	}
}

func UnlockKeyGen(request *messaging.ObjectRequest, client *goredis.Redis) {
	key := "KeyGenLock." + request.Controls.Namespace + "." + request.Controls.Class
	err := client.Set(key, "false", 0, 0, false, false)
	if err != nil {
		RedisConnection = nil
		RedisConnectionTCP = nil
		fmt.Println(err.Error())
	}
}

func SetKeyGenTime(request *messaging.ObjectRequest, client *goredis.Redis) {
	key := "KeyGenTime." + request.Controls.Namespace + "." + request.Controls.Class
	nowTime := time.Now().UTC().Format("2006-01-02 15:04:05")
	err := client.Set(key, nowTime, 0, 0, false, false)
	if err != nil {
		RedisConnection = nil
		RedisConnectionTCP = nil
		fmt.Println(err.Error())
	}
}

func CheckIfTimeToUpdateDB(request *messaging.ObjectRequest, client *goredis.Redis, timeInMinutes float64) (status bool) {
	status = false
	timeKey := "KeyGenTime." + request.Controls.Namespace + "." + request.Controls.Class

	val, err := client.Get(timeKey)
	if err != nil {
		RedisConnection = nil
		RedisConnectionTCP = nil
		fmt.Println(err.Error())
		return
	}

	if val == nil {
		nowTime := time.Now().UTC().Format("2006-01-02 15:04:05")
		_ = client.Set(timeKey, nowTime, 0, 0, false, false)
		return
	}

	KeyGenTime, err := time.Parse("2006-01-02 15:04:05", string(val))
	if err != nil {
		RedisConnection = nil
		RedisConnectionTCP = nil
		fmt.Println(err.Error())
	} else {
		difference := time.Now().UTC().Sub(KeyGenTime)
		if difference.Minutes() > timeInMinutes {
			SetKeyGenTime(request, client)
			status = true
		}
	}
	return
}

func SetKeyGenKey(request *messaging.ObjectRequest, client *goredis.Redis, value string) {
	key := "KeyGenKey." + request.Controls.Namespace + "." + request.Controls.Class
	err := client.Set(key, value, 0, 0, false, false)
	if err != nil {
		RedisConnection = nil
		RedisConnectionTCP = nil
		fmt.Println(err.Error())
	}
}

func GetKeyGenKey(request *messaging.ObjectRequest, client *goredis.Redis) (value string) {
	key := "KeyGenKey." + request.Controls.Namespace + "." + request.Controls.Class
	val, err := client.Incr(key)
	if err != nil {
		RedisConnection = nil
		RedisConnectionTCP = nil
		fmt.Println(err.Error())
		return
	}
	value = strconv.FormatInt(val, 10)
	return
}

func ReadKeyGenKey(request *messaging.ObjectRequest, client *goredis.Redis) (value string) {
	key := "KeyGenKey." + request.Controls.Namespace + "." + request.Controls.Class
	bvalue, err := client.Get(key)
	if err != nil {
		RedisConnection = nil
		RedisConnectionTCP = nil
		fmt.Println(err.Error())
	}
	err = json.Unmarshal(bvalue, &value)
	if err != nil {
		var intVal int
		err = json.Unmarshal(bvalue, &intVal)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			value = strconv.Itoa(intVal)
		}
	}
	return
}

func CreateNewKeyGenBundle(request *messaging.ObjectRequest) {
	client, err := GetConnection(request)
	if err != nil {
		RedisConnection = nil
		RedisConnectionTCP = nil
		fmt.Println(err.Error())
		return
	}

	if !CheckForKeyGen(request, client) {
		SetKeyGenKey(request, client, "0")
		SetKeyGenTime(request, client)
		_ = CheckForKeyGen(request, client)
	}
}

func FlushCache(request *messaging.ObjectRequest) {
	client, err := GetConnection(request)
	if err != nil {
		RedisConnection = nil
		RedisConnectionTCP = nil
		fmt.Println(err.Error())
		return
	}

	_ = client.FlushAll()
}

//support functions

func GetPatternAttributes(request *messaging.ObjectRequest) (prefix, value string) {
	classname := request.Controls.Class
	classLowered := strings.ToLower(classname)

	isIndexFound := false
	index := 0

	for x := 0; x < len(classname); x++ {
		_, err := strconv.Atoi(string(classname[x]))

		if err == nil {
			if !isIndexFound {
				match, _ := regexp.MatchString("([a-z]+)", classLowered[x:])
				if !match {
					index = x
					isIndexFound = true
				}
			}
		}
	}

	prefix = classname[:index]
	value = classname[index:]
	return
}
