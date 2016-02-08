package repositories

import (
	"duov6.com/objectstore/cache"
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"fmt"
)

func Execute(request *messaging.ObjectRequest, repository AbstractRepository) (response RepositoryResponse) {

	switch request.Controls.Operation { //CREATE, READ, UPDATE, DELETE, SPECIAL
	case "insert":
		if request.Controls.Multiplicity == "single" {
			response = repository.InsertSingle(request)
		} else {
			response = repository.InsertMultiple(request)
		}

	case "read-all":
		//check cache
		result := cache.Search(request)
		if result == nil {
			fmt.Println("Not Available in Cache.. Reading from Repositories...")
			response = repository.GetAll(request)
		} else {
			response.IsSuccess = true
			response.Body = result

			if response.IsSuccess && !checkEmptyByteArray(response.Body) {
				var data interface{}
				_ = json.Unmarshal(response.Body, &data)
				if errCache := cache.StoreResult(request, data); errCache != nil {
					fmt.Println(errCache.Error())
				}
			}

		}
		//response = repository.GetAll(request)
	case "read-key":
		//check cache
		result := cache.GetByKey(request)
		if result == nil {
			fmt.Println("Not Available in Cache.. Reading from Repositories...")
			response = repository.GetByKey(request)
		} else {
			response.IsSuccess = true
			response.Body = result

			if response.IsSuccess && !checkEmptyByteArray(response.Body) {
				var data map[string]interface{}
				data = make(map[string]interface{})
				_ = json.Unmarshal(response.Body, &data)

				if errCache := cache.StoreOne(request, data); errCache != nil {
					fmt.Println(errCache.Error())
				}
			}
		}
		//response = repository.GetByKey(request)
	case "read-keyword":
		//check cache
		result := cache.Search(request)
		if result == nil {
			fmt.Println("Not Available in Cache.. Reading from Repositories...")
			response = repository.GetSearch(request)
		} else {
			response.IsSuccess = true
			response.Body = result

			if response.IsSuccess && !checkEmptyByteArray(response.Body) {
				var data interface{}
				_ = json.Unmarshal(response.Body, &data)
				if errCache := cache.StoreResult(request, data); errCache != nil {
					fmt.Println(errCache.Error())
				}
			}
		}
		//response = repository.GetSearch(request)
	case "read-filter":
		response = repository.GetQuery(request)
	case "update":
		if request.Controls.Multiplicity == "single" {
			response = repository.UpdateSingle(request)
		} else {
			response = repository.UpdateMultiple(request)
		}
	case "delete":
		if request.Controls.Multiplicity == "single" {
			response = repository.DeleteSingle(request)
		} else {
			response = repository.DeleteMultiple(request)
		}
	case "special":
		response = repository.Special(request)
	}

	return
}
