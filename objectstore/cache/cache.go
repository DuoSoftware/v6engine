package cache

import (
	"duov6.com/objectstore/cache/repositories"
	"duov6.com/objectstore/messaging"
	"fmt"
)

func Delete(request *messaging.ObjectRequest) {

}

func GetByKey(request *messaging.ObjectRequest) (body []byte) {
	if CheckCacheAvailability(request) {
		body = repositories.GetByKey(request)
	} else {
		fmt.Println("Redis not available to host Cache!")
	}

	return
}

func StoreOne(request *messaging.ObjectRequest, data map[string]interface{}) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.SetOneRedis(request, data)
		if err != nil {
			fmt.Println("Error storing to Cache : " + err.Error())
		}
	} else {
		fmt.Println("Redis not available to host Cache!")
	}
	return
}

func StoreMany(request *messaging.ObjectRequest, data []map[string]interface{}) (err error) {
	if CheckCacheAvailability(request) {
		err = repositories.SetManyRedis(request, data)
		if err != nil {
			fmt.Println("Error storing to Cache : " + err.Error())
		}
	} else {
		fmt.Println("Redis not available to host Cache!")
	}
	return
}

func CheckCacheAvailability(request *messaging.ObjectRequest) (status bool) {
	status = true
	if request.Configuration.ServerConfiguration["REDIS"] == nil {
		fmt.Println("No Config Found")
		status = false
	}
	return
}
