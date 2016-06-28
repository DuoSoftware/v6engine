package repositories

import (
	"duov6.com/objectstore/cache"
	"duov6.com/objectstore/messaging"
	"duov6.com/term"
	"encoding/json"
)

func Execute(request *messaging.ObjectRequest, repository AbstractRepository) (response RepositoryResponse) {

	switch request.Controls.Operation { //CREATE, READ, UPDATE, DELETE, SPECIAL
	case "insert":
		if request.Controls.Multiplicity == "single" {
			response = repository.InsertSingle(request)
			if response.IsSuccess {
				// if errCache := cache.StoreOne(request, request.Body.Object, cache.Data); errCache != nil {
				// 	term.Write(errCache.Error(), term.Debug)
				// }
				go cache.StoreOne(request, request.Body.Object, cache.Data)
			}
		} else {
			response = repository.InsertMultiple(request)
			if response.IsSuccess {
				// if errCache := cache.StoreMany(request, request.Body.Objects, cache.Data); errCache != nil {
				// 	term.Write(errCache.Error(), term.Debug)
				// }
				go cache.StoreMany(request, request.Body.Objects, cache.Data)
			}
		}

	case "read-all":
		//check cache
		result := cache.Search(request, cache.Data)
		if result == nil {
			term.Write("Not Available in Cache.. Reading from Repositories...", term.Debug)
			response = repository.GetAll(request)

			if response.IsSuccess && !checkEmptyByteArray(response.Body) {
				var data []map[string]interface{}
				_ = json.Unmarshal(response.Body, &data)
				if errCache := cache.StoreResult(request, data, cache.Data); errCache != nil {
					term.Write(errCache.Error(), term.Debug)
				}
			}
		} else {
			response.IsSuccess = true
			response.Body = result

		}
		//response = repository.GetAll(request)
	case "read-key":
		//check cache
		result := cache.GetByKey(request, cache.Data)
		if result == nil {
			term.Write("Not Available in Cache.. Reading from Repositories...", term.Debug)
			response = repository.GetByKey(request)

			if response.IsSuccess && !checkEmptyByteArray(response.Body) {
				var data map[string]interface{}
				data = make(map[string]interface{})
				_ = json.Unmarshal(response.Body, &data)

				if errCache := cache.StoreOne(request, data, cache.Data); errCache != nil {
					term.Write(errCache.Error(), term.Debug)
				}
			}
		} else {
			response.IsSuccess = true
			response.Body = result
		}
		//response = repository.GetByKey(request)
	case "read-keyword":
		//check cache
		result := cache.Search(request, cache.Data)
		if result == nil {
			term.Write("Not Available in Cache.. Reading from Repositories...", term.Debug)
			response = repository.GetSearch(request)
			if response.IsSuccess && !checkEmptyByteArray(response.Body) {
				var data []map[string]interface{}
				_ = json.Unmarshal(response.Body, &data)
				if errCache := cache.StoreResult(request, data, cache.Data); errCache != nil {
					term.Write(errCache.Error(), term.Debug)
				}
			}
		} else {
			response.IsSuccess = true
			response.Body = result

		}
		//response = repository.GetSearch(request)
	case "read-filter":
		//check cache
		result := cache.Query(request, cache.Data)
		if result == nil {
			term.Write("Not Available in Cache.. Reading from Repositories...", term.Debug)
			response = repository.GetQuery(request)
			if response.IsSuccess && !checkEmptyByteArray(response.Body) {
				var data []map[string]interface{}
				_ = json.Unmarshal(response.Body, &data)
				if errCache := cache.StoreQuery(request, data, cache.Data); errCache != nil {
					term.Write(errCache.Error(), term.Debug)
				}
			}
		} else {
			response.IsSuccess = true
			response.Body = result

		}
		//response = repository.GetQuery(request)
	case "update":
		if request.Controls.Multiplicity == "single" {
			response = repository.UpdateSingle(request)
			if response.IsSuccess {
				// if errCache := cache.StoreOne(request, request.Body.Object, cache.Data); errCache != nil {
				// 	term.Write(errCache.Error(), term.Debug)
				// }
				go cache.StoreOne(request, request.Body.Object, cache.Data)
			}
		} else {
			response = repository.UpdateMultiple(request)
			if response.IsSuccess {
				// if errCache := cache.StoreMany(request, request.Body.Objects, cache.Data); errCache != nil {
				// 	term.Write(errCache.Error(), term.Debug)
				// }
				cache.StoreMany(request, request.Body.Objects, cache.Data)
			}
		}
	case "delete":
		if request.Controls.Multiplicity == "single" {
			response = repository.DeleteSingle(request)
			if response.IsSuccess {
				if errCache := cache.DeleteOne(request, request.Body.Object, cache.Data); errCache != nil {
					term.Write(errCache.Error(), term.Debug)
				}
			}
		} else {
			response = repository.DeleteMultiple(request)
			if response.IsSuccess {
				if errCache := cache.DeleteMany(request, request.Body.Objects, cache.Data); errCache != nil {
					term.Write(errCache.Error(), term.Debug)
				}
			}
		}
	case "special":
		response = repository.Special(request)
	}

	return
}
