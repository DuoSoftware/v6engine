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
			request.Extras["IgnoreCacheRead"] = true
			response = repository.InsertSingle(request)
			if response.IsSuccess {
				request.Extras["IgnoreCacheRead"] = false
				cache.ResetSearchResults(request, cache.Data)
				//PushSingleMapToCache(request, request.Body.Object)
			}
		} else {
			request.Extras["IgnoreCacheRead"] = true
			response = repository.InsertMultiple(request)
			if response.IsSuccess {
				request.Extras["IgnoreCacheRead"] = false
				cache.ResetSearchResults(request, cache.Data)
				//PushMultipleMapToCache(request, request.Body.Objects)
			}
		}

	case "read-all":
		//check cache
		result := cache.Search(request, cache.Data)
		if result == nil {
			request.Extras["IgnoreCacheRead"] = true
			request.Log("Debug : Not Available in Cache.. Reading from Repositories...")
			response = repository.GetAll(request)
			if response.IsSuccess && !checkEmptyByteArray(response.Body) {
				request.Extras["IgnoreCacheRead"] = false
				var data []map[string]interface{}
				_ = json.Unmarshal(response.Body, &data)
				if errCache := cache.StoreResult(request, data, cache.Data); errCache != nil {
					request.Log("Error : " + errCache.Error())
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
			request.Extras["IgnoreCacheRead"] = true
			request.Log("Debug : Not Available in Cache.. Reading from Repositories...")
			response = repository.GetByKey(request)

			if response.IsSuccess && !checkEmptyByteArray(response.Body) {
				request.Extras["IgnoreCacheRead"] = false
				var data map[string]interface{}
				data = make(map[string]interface{})
				_ = json.Unmarshal(response.Body, &data)

				if errCache := cache.StoreOne(request, data, cache.Data); errCache != nil {
					request.Log("Error : " + errCache.Error())
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
			request.Extras["IgnoreCacheRead"] = true
			term.Write("Not Available in Cache.. Reading from Repositories...", term.Debug)
			response = repository.GetSearch(request)
			if response.IsSuccess && !checkEmptyByteArray(response.Body) {
				request.Extras["IgnoreCacheRead"] = false
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
		/*result := cache.Query(request, cache.Data)
		if result == nil {
			request.Extras["IgnoreCacheRead"] = true
			term.Write("Not Available in Cache.. Reading from Repositories...", term.Debug)
			response = repository.GetQuery(request)
			if response.IsSuccess && !checkEmptyByteArray(response.Body) {
				request.Extras["IgnoreCacheRead"] = false
				var data []map[string]interface{}
				_ = json.Unmarshal(response.Body, &data)
				if errCache := cache.StoreQuery(request, data, cache.Data); errCache != nil {
					term.Write(errCache.Error(), term.Debug)
				}
			}
		} else {
			response.IsSuccess = true
			response.Body = result

		}*/
		response = repository.GetQuery(request)
	case "update":
		if request.Controls.Multiplicity == "single" {
			response = repository.UpdateSingle(request)
			if response.IsSuccess {
				cache.ResetSearchResults(request, cache.Data)
				//PushSingleMapToCache(request, request.Body.Object)
			}
		} else {
			response = repository.UpdateMultiple(request)
			if response.IsSuccess {
				cache.ResetSearchResults(request, cache.Data)
				//PushMultipleMapToCache(request, request.Body.Objects)
			}
		}
	case "delete":
		request.Extras["IgnoreCacheRead"] = true
		if request.Controls.Multiplicity == "single" {
			response = repository.DeleteSingle(request)
			if response.IsSuccess {
				if errCache := cache.DeleteOne(request, request.Body.Object, cache.Data); errCache != nil {
					term.Write(errCache.Error(), term.Debug)
				}
			}
			cache.ResetSearchResults(request, cache.Data)
		} else {
			response = repository.DeleteMultiple(request)
			if response.IsSuccess {
				if errCache := cache.DeleteMany(request, request.Body.Objects, cache.Data); errCache != nil {
					term.Write(errCache.Error(), term.Debug)
				}
			}
			cache.ResetSearchResults(request, cache.Data)
		}
	case "special":
		request.Extras["IgnoreCacheRead"] = true
		cache.ResetSearchResults(request, cache.Data)
		response = repository.Special(request)
	}

	return
}

func CheckCacheable(request *messaging.ObjectRequest) (status bool) {
	status = false
	if request.Extras["IsCached"] == nil {
		status = true
	}
	return
}

func RevertCacheable(request *messaging.ObjectRequest) {
	request.Extras["IsCached"] = true
}

func PushSingleMapToCache(request *messaging.ObjectRequest, obj map[string]interface{}) {
	go cache.ResetSearchResults(request, cache.Data)
	return //execute nothing under here
	if CheckCacheable(request) {
		RevertCacheable(request)
		if cache.GetIncrValue(request, ("RequestCounter."+request.Controls.Namespace+"."+request.Controls.Class), cache.RequestCounter) < 1000 {
			go cache.StoreOne(request, obj, cache.Data)
		}
	}
}

func PushMultipleMapToCache(request *messaging.ObjectRequest, objs []map[string]interface{}) {
	go cache.ResetSearchResults(request, cache.Data)
	return //execute nothing under here
	if len(objs) == 1 {
		return //execute nothing under here
		if CheckCacheable(request) {
			RevertCacheable(request)
			if cache.GetIncrValue(request, ("RequestCounter."+request.Controls.Namespace+"."+request.Controls.Class), cache.RequestCounter) < 1000 {
				go cache.StoreMany(request, objs, cache.Data)
			}
		}
	}
}
