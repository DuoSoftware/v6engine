package keygenerator

import (
	"duov6.com/common"
	"duov6.com/objectstore/messaging"
	//"duov6.com/objectstore/repositories"
	"errors"
	"github.com/xuyu/goredis"
	//"strings"
	"fmt"
)

func GetIncrementID(request *messaging.ObjectRequest, repository string) (key string) {

	ifShouldVerifyList, err := VerifyListRefresh(request)
	if err != nil {
		key = common.GetGUID()
	}
	fmt.Println(ifShouldVerifyList)
	key = common.GetGUID()
	// if ifShouldVerifyList {
	// 	var repository repositories.AbstractRepository
	// 	if strings.EqualFold(repository, "CloudSQL") {
	// 		repository = repositories.Create("CLOUDSQL")
	// 	} else if strings.EqualFold(repository, "ELASTIC") {
	// 		repository = repositories.Create("ELASTIC")
	// 	}
	// 	go repository.IncrementDomainClassAttributes(request, 1950)
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
