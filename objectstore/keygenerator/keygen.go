package keygenerator

import (
	"duov6.com/common"
	//"duov6.com/objectstore/keygenerator/drivers"
	"duov6.com/objectstore/messaging"
	"errors"
	"fmt"
	"github.com/xuyu/goredis"
	//"strings"
)

func GetIncrementID(request *messaging.ObjectRequest, repository string) (key string) {

	ifShouldVerifyList, err := VerifyListRefresh(request)
	if err != nil {
		key = common.GetGUID()
	}
	fmt.Println(ifShouldVerifyList)
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
