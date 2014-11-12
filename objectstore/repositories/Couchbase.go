package repositories

import (
	"duov6.com/objectstore/messaging"
	"fmt"
	"github.com/couchbaselabs/go-couchbase"
)

type CouchRepository struct {
}

func (repository CouchRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository CouchRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository CouchRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository CouchRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {

	response := RepositoryResponse{}
	bucket, errorMessage, isError := getCouchBucket()(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id
		fmt.Println("Couchbase Get By Key" + key)
		rawBytes, err := bucket.GetRaw(key)
		if err != nil {
			response.GetErrorResponse("Error retrieving object from couchbase")
		} else {
			response.GetResponseWithBody(rawBytes)
		}

	}

	return response
}

func (repository CouchRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository CouchRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	fmt.Println("Couchbase Single Insert")
	response := setOne(request)
	return response
}

func (repository CouchRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository CouchRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	response := setOne(request)
	return response
}

func (repository CouchRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository CouchRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	bucket, errorMessage, isError := getCouchBucket()(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id
		err := bucket.Delete(key)
		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error deleting one object in Couchbase" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully deleted one object in Coucahbase"
		}

	}

	return response
}

func (repository CouchRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository CouchRepository) Test(request *messaging.ObjectRequest) {

}

func setOne(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	bucket, errorMessage, isError := getCouchBucket()(request)

	if isError == true {
		fmt.Println(errorMessage)
		response.GetErrorResponse(errorMessage)
	} else {
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id
		err := bucket.Set(key, 0, request.Body.Object)
		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error inserting/updating one object in Couchbase" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted/updated one object in Coucahbase"
		}

	}

	return response
}

func setMany(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	bucket, errorMessage, isError := getCouchBucket()(request)

	if isError == true {
		fmt.Println(errorMessage)
		response.GetErrorResponse(errorMessage)
	} else {
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id
		err := bucket.Set(key, 0, request.Body.Object)
		if err != nil {
			response.IsSuccess = false
			response.GetErrorResponse("Error inserting/updating one object in Couchbase" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserting/updating one object in Coucahbase"
		}

	}

	return response
}

//func getCouchBucket(request *messaging.ObjectRequest) (bucket *couchbase.Bucket, errorMessage string, isError bool) {

//}

func getCouchBucket() func(request *messaging.ObjectRequest) (bucket *couchbase.Bucket, errorMessage string, isError bool) {

	var createdBucket *couchbase.Bucket

	return func(request *messaging.ObjectRequest) (bucket *couchbase.Bucket, errorMessage string, isError bool) {
		isError = false

		if createdBucket == nil {
			setting_host := request.Configuration.ServerConfiguration["COUCH"]["Url"]
			setting_bucket := request.Configuration.ServerConfiguration["COUCH"]["Bucket"]
			//setting_userName := request.StoreConfiguration.ServerConfiguration["COUCH"]["UserName"]
			//setting_password := request.StoreConfiguration.ServerConfiguration["COUCH"]["Password"]

			c, err := couchbase.Connect(setting_host)
			if err != nil {
				isError = true
				errorMessage = "Error connecting Couchbase to :  " + setting_host
			}

			pool, err := c.GetPool("default")
			if err != nil {
				isError = true
				errorMessage = "Error getting pool: "
			}

			returnBucket, err := pool.GetBucket(setting_bucket)

			if err != nil {
				isError = true
				errorMessage = "Error getting Couchbase bucket: " + setting_bucket
			} else {
				createdBucket = returnBucket
			}

			return

		} else {
			bucket = createdBucket
			return
		}
	}
}
