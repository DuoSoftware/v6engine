package repositories

import (
	"duov6.com/objectstore/messaging"
	"github.com/couchbaselabs/go-couchbase"
)

type CouchRepository struct {
}

func (repository CouchRepository) GetRepositoryName() string {
	return "Couchbase"
}

func (repository CouchRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("GetAll not implemented in Couchbase repository")
	return getDefaultNotImplemented()
}

func (repository CouchRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("GetSearch not implemented in Couchbase repository")
	return getDefaultNotImplemented()
}

func (repository CouchRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("GetQuery not implemented in Couchbase repository")
	return getDefaultNotImplemented()
}

func (repository CouchRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {

	response := RepositoryResponse{}
	bucket, errorMessage, isError := getCouchBucket()(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id
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
	request.Log("InsertMultiple not implemented in Couchbase repository")
	return getDefaultNotImplemented()
}

func (repository CouchRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	response := setOne(request)
	return response
}

func (repository CouchRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("UpdateMultiple not implemented in Couchbase repository")
	return getDefaultNotImplemented()
}

func (repository CouchRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	response := setOne(request)
	return response
}

func (repository CouchRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("DeleteMultiple not implemented in Couchbase repository")
	return getDefaultNotImplemented()
}

func (repository CouchRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	bucket, errorMessage, isError := getCouchBucket()(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id
		request.Log("Deleting object from couchbase : " + key)
		err := bucket.Delete(key)
		if err != nil {
			response.IsSuccess = false
			request.Log("Error deleting object from couchbase : " + key + ", " + err.Error())
			response.GetErrorResponse("Error deleting one object in Couchbase" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully deleted one object in Coucahbase"
			request.Log(response.Message)
		}

	}

	return response
}

func (repository CouchRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Special not implemented in Couchbase repository")
	return getDefaultNotImplemented()
}

func (repository CouchRepository) Test(request *messaging.ObjectRequest) {

}

func setOne(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	bucket, errorMessage, isError := getCouchBucket()(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id
		request.Log("Inserting/Updating object in Couchbase : " + key)
		err := bucket.Set(key, 0, request.Body.Object)
		if err != nil {
			response.IsSuccess = false
			request.Log("Error inserting/updating object in Couchbase : " + key + ", " + err.Error())
			response.GetErrorResponse("Error inserting/updating one object in Couchbase" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserted/updated one object in Coucahbase"
			request.Log(response.Message)
		}

	}

	return response
}

func setMany(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}
	bucket, errorMessage, isError := getCouchBucket()(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id
		err := bucket.Set(key, 0, request.Body.Object)
		if err != nil {
			response.IsSuccess = false
			request.Log("Error inserting/updating multiple objects in Couchbase : " + key + ", " + err.Error())
			response.GetErrorResponse("Error inserting/updating one object in Couchbase" + err.Error())
		} else {
			response.IsSuccess = true
			response.Message = "Successfully inserting/updating one object in Coucahbase"
			request.Log(response.Message)
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
			request.Log("Getting store configuration settings for Couchbase")

			setting_host := request.Configuration.ServerConfiguration["COUCH"]["Url"]
			setting_bucket := request.Configuration.ServerConfiguration["COUCH"]["Bucket"]
			//setting_userName := request.StoreConfiguration.ServerConfiguration["COUCH"]["UserName"]
			//setting_password := request.StoreConfiguration.ServerConfiguration["COUCH"]["Password"]
			request.Log("Store configuration settings recieved for Couchbase Host : " + setting_host + " , Bucket : " + setting_bucket)

			c, err := couchbase.Connect(setting_host)
			if err != nil {
				isError = true
				errorMessage = "Error connecting Couchbase to :  " + setting_host
				request.Log(errorMessage)
			}

			pool, err := c.GetPool("default")
			if err != nil {
				isError = true
				errorMessage = "Error getting pool: "
				request.Log(errorMessage)
			}

			returnBucket, err := pool.GetBucket(setting_bucket)

			if err != nil {
				isError = true
				errorMessage = "Error getting Couchbase bucket: " + setting_bucket
				request.Log(errorMessage)
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
