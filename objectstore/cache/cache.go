package cache

import (
	"duov6.com/objectstore/cache/repositories"
	"duov6.com/objectstore/messaging"
	"fmt"
	"strings"
)

func DeleteOne(request *messaging.ObjectRequest, data map[string]interface{}) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.RemoveOneRedis(request, data)
		if err != nil {
			fmt.Println("Error storing to Cache : " + err.Error())
		}
	}
	return
}

func DeleteMany(request *messaging.ObjectRequest, data []map[string]interface{}) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.RemoveManyRedis(request, data)
		if err != nil {
			fmt.Println("Error storing to Cache : " + err.Error())
		}
	}
	return
}

func Search(request *messaging.ObjectRequest) (body []byte) {
	if CheckCacheAvailability(request) {
		body = repositories.GetSearch(request)
	}
	return
}

func Query(request *messaging.ObjectRequest) (body []byte) {
	if CheckCacheAvailability(request) {
		body = repositories.GetQuery(request)
	}
	return
}

func GetByKey(request *messaging.ObjectRequest) (body []byte) {
	if CheckCacheAvailability(request) {
		body = repositories.GetByKey(request)
	}
	return
}

func StoreOne(request *messaging.ObjectRequest, data map[string]interface{}) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.SetOneRedis(request, data)
		if err != nil {
			fmt.Println("Error storing to Cache : " + err.Error())
		}
	}
	return
}

func StoreMany(request *messaging.ObjectRequest, data []map[string]interface{}) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.SetManyRedis(request, data)
		if err != nil {
			fmt.Println("Error storing to Cache : " + err.Error())
		}
	}
	return
}

func StoreResult(request *messaging.ObjectRequest, data interface{}) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.SetResultRedis(request, data)
		if err != nil {
			fmt.Println("Error storing to Cache : " + err.Error())
		}
	}
	return
}

func StoreQuery(request *messaging.ObjectRequest, data interface{}) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.SetQueryRedis(request, data)
		if err != nil {
			fmt.Println("Error storing to Cache : " + err.Error())
		}
	}
	return
}

func CheckCacheAvailability(request *messaging.ObjectRequest) (status bool) {
	status = true
	if request.Configuration.ServerConfiguration["REDIS"] == nil {
		fmt.Println("Cache Config/Server Not Found!")
		status = false
	} else if !checkValidTenentClass(request) {
		status = false
	}

	return
}

func checkValidTenentClass(request *messaging.ObjectRequest) (status bool) {
	namespaces := [...]string{}
	classes := [...]string{"domainclassattributes"}

	status = true

	for _, namespace := range namespaces {
		if strings.ToLower(request.Controls.Namespace) == strings.ToLower(namespace) {
			fmt.Println("Invalid Namespace. Wouldn't be saved on Cache")
			status = false
			return
		}
	}
	for _, class := range classes {
		if strings.ToLower(request.Controls.Class) == strings.ToLower(class) {
			fmt.Println("Invalid Class. Wouldn't be saved on Cache!")
			status = false
			return
		}
	}

	return
}
