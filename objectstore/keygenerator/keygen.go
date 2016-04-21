package keygenerator

import (
	"duov6.com/common"
	"duov6.com/objectstore/keygenerator/drivers"
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xuyu/goredis"
	"strconv"
	"time"
	//"strings"
)

func GetIncrementID(request *messaging.ObjectRequest, repository string) (key string) {

	client, err := GetConnection(request)
	if err != nil {
		key = common.GetGUID()
		fmt.Println(err.Error() + " Returning an Unique GUID : " + key)
	} else {
		key = ExecuteKeyGenProcess(client, request, repository)
	}

	return
}

func ExecuteKeyGenProcess(client *goredis.Redis, request *messaging.ObjectRequest, repository string) (key string) {
	if status := CheckForKeyGen(request, client); status {
		//Key Available in Database
		if isLock := CheckKeyGenLock(request, client); isLock {
			for true {
				time.Sleep(1 * time.Second)
				if isLock = CheckKeyGenLock(request, client); !isLock {
					// Lock is Over
					key = GetKeyGenKey(request, client)
					return
				}
			}
		} else {
			if isUpdateTime := CheckIfTimeToUpdateDB(request, client, float64(5.0)); isUpdateTime {
				if isLocked := CheckKeyGenLock(request, client); isLocked {
					for true {
						time.Sleep(1 * time.Second)
						if isLocked = CheckKeyGenLock(request, client); !isLocked {
							// Lock is Over
							key = GetKeyGenKey(request, client)
							return
						}
					}
				} else {
					//Not locked. Ready to update
					LockKeyGen(request, client)
					currentVal := ReadKeyGenKey(request, client)
					IntCurrentVal, err := strconv.Atoi(currentVal)
					if err != nil {
						fmt.Println(err.Error())
						return
					}

					max := VerifyMaxFromDB(request, repository, IntCurrentVal, false)

					intMax, err := strconv.Atoi(max)
					if err != nil {
						key = GetKeyGenKey(request, client)
						fmt.Println(err.Error())
						return
					}

					if intMax > IntCurrentVal {
						SetKeyGenKey(request, client, max)
					}

					UnlockKeyGen(request, client)
					key = max
				}
			} else {
				key = GetKeyGenKey(request, client)
			}
		}

	} else {
		//Key Not Available in Database
		LockKeyGen(request, client)
		max := VerifyMaxFromDB(request, repository, 0, true)
		SetKeyGenKey(request, client, max)
		UnlockKeyGen(request, client)
		SetKeyGenTime(request, client)
		key = max
	}
	return
}

func VerifyMaxFromDB(request *messaging.ObjectRequest, repository string, count int, verifySchema bool) (max string) {
	fmt.Println("Readying to Update DomainClassAttributes class....")
	switch repository {
	case "CLOUDSQL":
		var sqlDriver drivers.CloudSql
		max = sqlDriver.VerifyMaxValueDB(request, count, verifySchema)
		break
	case "ELASTIC":
		break
	default:
		fmt.Println("Error! No such Repository : " + repository + " exists!")
		break
	}
	return
}

var RedisConnection *goredis.Redis

func GetConnection(request *messaging.ObjectRequest) (client *goredis.Redis, err error) {
	if RedisConnection == nil {
		client, err = goredis.DialURL("tcp://@" + request.Configuration.ServerConfiguration["REDIS"]["Host"] + ":" + request.Configuration.ServerConfiguration["REDIS"]["Port"] + "/0?timeout=1s&maxidle=1")
		if err != nil {
			return nil, err
		}
		if client == nil {
			return nil, errors.New("Connection to REDIS Failed!")
		}
	} else {
		if err = RedisConnection.Ping(); err != nil {
			RedisConnection = nil
			client, err = goredis.DialURL("tcp://@" + request.Configuration.ServerConfiguration["REDIS"]["Host"] + ":" + request.Configuration.ServerConfiguration["REDIS"]["Port"] + "/0?timeout=1s&maxidle=1")
			if err != nil {
				return nil, err
			}
			if client == nil {
				return nil, errors.New("Connection to REDIS Failed!")
			}
		} else {
			client = RedisConnection
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
		fmt.Println(err.Error())
	}
	return
}

func CheckKeyGenLock(request *messaging.ObjectRequest, client *goredis.Redis) (status bool) {
	status = false
	key := "KeyGenLock." + request.Controls.Namespace + "." + request.Controls.Class
	val, err := client.Get(key)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if val == nil {
		_ = client.Set(key, "false", 0, 0, false, false)
		return
	}

	err = json.Unmarshal(val, &status)
	if err != nil {
		fmt.Println(err.Error())
	}

	return

}

func LockKeyGen(request *messaging.ObjectRequest, client *goredis.Redis) {
	key := "KeyGenLock." + request.Controls.Namespace + "." + request.Controls.Class
	err := client.Set(key, "false", 0, 0, false, false)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func UnlockKeyGen(request *messaging.ObjectRequest, client *goredis.Redis) {
	key := "KeyGenLock." + request.Controls.Namespace + "." + request.Controls.Class
	err := client.Set(key, "false", 0, 0, false, false)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func SetKeyGenTime(request *messaging.ObjectRequest, client *goredis.Redis) {
	key := "KeyGenTime." + request.Controls.Namespace + "." + request.Controls.Class
	nowTime := time.Now().UTC().Format("2006-01-02 15:04:05")
	err := client.Set(key, nowTime, 0, 0, false, false)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func CheckIfTimeToUpdateDB(request *messaging.ObjectRequest, client *goredis.Redis, timeInMinutes float64) (status bool) {
	status = false
	timeKey := "KeyGenTime." + request.Controls.Namespace + "." + request.Controls.Class

	val, err := client.Get(timeKey)
	if err != nil {
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
		fmt.Println(err.Error())
	}
}

func GetKeyGenKey(request *messaging.ObjectRequest, client *goredis.Redis) (value string) {
	key := "KeyGenKey." + request.Controls.Namespace + "." + request.Controls.Class
	val, err := client.Incr(key)
	if err != nil {
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
