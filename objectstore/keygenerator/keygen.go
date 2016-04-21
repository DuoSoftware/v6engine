package keygenerator

import (
	"duov6.com/common"
	//"duov6.com/objectstore/keygenerator/drivers"
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

	// ifShouldVerifyList, err := VerifyListRefresh(request)
	// if err != nil {
	// 	key = common.GetGUID()
	// }
	//fmt.Println(ifShouldVerifyList)
	key = common.GetGUID()

	// if ifShouldVerifyList {
	// 	if strings.EqualFold(repository, "CloudSQL") {
	// 		//check if key is available
	// 		listKey := "KeyGenList." + request.Controls.Namespace + "." + request.Controls.Class
	// 		lockKey := "KeyGenLock." + request.Controls.Namespace + "." + request.Controls.Class
	// 		var sqlDriver drivers.CloudSql
	// 		go sqlDriver.UpdateCloudSqlRecordID(request, 1450)
	// 	} else if strings.EqualFold(repository, "ELASTIC") {
	// 		//go drivers.UpdateElasticRecordID(request, 1450)
	// 	}
	// }

	return
}

func VerifyListRefresh(request *messaging.ObjectRequest) (status bool, err error) {
	status = false

	client, err := GetConnection(request)
	if err != nil {
		return
	}

	listKey := "KeyGenList." + request.Controls.Namespace + "." + request.Controls.Class

	length, err := client.LLen(listKey)
	if err != nil {
		return
	}

	if length < 550 {
		status = true
	}
	client.ClosePool()
	return
}

func GetConnection(request *messaging.ObjectRequest) (client *goredis.Redis, err error) {
	client, err = goredis.DialURL("tcp://@" + request.Configuration.ServerConfiguration["REDIS"]["Host"] + ":" + request.Configuration.ServerConfiguration["REDIS"]["Port"] + "/0?timeout=1s&maxidle=1")
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("Connection to REDIS Failed!")
	}
	return
}

// func CheckListAvailability(key string) (status bool) {
// 	status = false
// 	client, err := GetConnection(request)
// 	if err != nil {
// 		return
// 	}

// 	status, err = client.Exists(key)
// 	if err != nil {
// 		return
// 	}
// 	client.ClosePool()
// 	return
// }

// func CheckIfKeysAvailable(listkey string, lockkey string) (status bool) {
// 	status = CheckListAvailability(lockkey)
// 	if !status {
// 		err = client.Set(lockkey, "false", 0, 0, false, false)
// 		if err != nil {
// 			return
// 		}
// 	}
// 	status = CheckListAvailability(listkey)
// }

// func SetListItems(value int, listName string) {
// 	client, err := GetConnection(request)
// 	if err != nil {
// 		return
// 	}

// 	client.ClosePool()
// }

//-----------------------------------------------------------------------------------------------------------

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
	nowTime := time.Now().Format("2006-01-02 15:04:05")
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
			fmt.Println("Readying to Update DomainClassAttributes class....")
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
	value = strconv.FormatInt(val, 16)
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
		fmt.Println(err.Error())
	}
	return
}
