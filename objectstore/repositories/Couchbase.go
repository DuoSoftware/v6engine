package repositories
//update servers to 15.04 so gocb is supported
/*
import (
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"fmt"
	"github.com/couchbaselabs/go-couchbase"
	"github.com/couchbaselabs/gocb"
	"github.com/twinj/uuid"
	"strconv"
	"strings"
	"time"
)
*/

type CouchRepository struct {
}
/*
func (repository CouchRepository) GetRepositoryName() string {
	return "Couchbase"
}

func (repository CouchRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-ALL")
	response := RepositoryResponse{}
	bucket, errorMessage, isError := getCouchBucket(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		//Get all IDs
		viewResult, err := bucket.View(("dev_" + getSQLnamespace(request)), ("dev_" + getSQLnamespace(request)), nil)

		//Iterate check for pattern and choose only desired IDs
		pattern := request.Controls.Namespace + "." + request.Controls.Class + "."

		var idArray []string
		for _, row := range viewResult.Rows {
			if strings.Contains((row.Key).(string), pattern) {
				idArray = append(idArray, (row.Key).(string))
			}
		}

		//read all data for fetched days and return
		var returnDataMap []map[string]interface{}

		data, err := bucket.GetBulkRaw(idArray)

		for key, _ := range data {
			var tempData map[string]interface{}
			tempData = make(map[string]interface{})
			json.Unmarshal(data[key], &tempData)

			if request.Controls.SendMetaData == "false" {
				delete(tempData, "__osHeaders")
			}

			returnDataMap = append(returnDataMap, tempData)
		}

		if len(returnDataMap) == 0 {
			response.IsSuccess = true
			response.Message = "No objects found in Couchbase"
			var emptyMap map[string]interface{}
			emptyMap = make(map[string]interface{})
			byte, _ := json.Marshal(emptyMap)
			response.GetResponseWithBody(byte)
		}

		rawBytes, err := json.Marshal(returnDataMap)
		if err != nil {
			fmt.Println(err.Error())
			response.GetErrorResponse("Error retrieving object from couchbase : " + err.Error())
		} else {
			response.GetResponseWithBody(rawBytes)
		}
	}

	return response
}

func (repository CouchRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("GetSearch not implemented in Couchbase repository")
	return getDefaultNotImplemented()
}

func (repository CouchRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-QUERY!")
	response := RepositoryResponse{}
	queryType := request.Body.Query.Type

	switch queryType {
	case "Query":
		if request.Body.Query.Parameters != "*" {
			request.Log("Support for SQL Query not implemented in CouchBase Db repository")
			return getDefaultNotImplemented()
		} else {
			return repository.GetAll(request)
		}
	default:
		request.Log(queryType + " not implemented in CouchBase Db repository")
		return getDefaultNotImplemented()
	}

	return response
}

func (repository CouchRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-BY-KEY")
	response := RepositoryResponse{}

	bucket, errorMessage, isError := getCouchBucket(request)

	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id
		rawBytes, err := bucket.GetRaw(key)

		var returnData map[string]interface{}
		returnData = make(map[string]interface{})

		json.Unmarshal(rawBytes, &returnData)

		if request.Controls.SendMetaData == "false" {
			delete(returnData, "__osHeaders")
		}

		if len(returnData) == 0 {
			response.IsSuccess = true
			response.Message = "No objects found in Couchbase"
			var emptyMap map[string]interface{}
			emptyMap = make(map[string]interface{})
			byte, _ := json.Marshal(emptyMap)
			response.GetResponseWithBody(byte)
		}

		rawBytes, err = json.Marshal(returnData)

		if err != nil {
			response.GetErrorResponse("Error retrieving object from couchbase")
		} else {
			response.GetResponseWithBody(rawBytes)
		}
	}

	return response
}

func (repository CouchRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-MULTIPLE")
	response := setMany(request)
	return response
}

func (repository CouchRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-SINGLE")
	response := setOne(request)
	return response
}

func (repository CouchRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-MULTIPLE")
	response := setMany(request)
	return response
}

func (repository CouchRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-SINGLE")
	response := setOne(request)
	return response
}

func (repository CouchRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE_MULTIPLE")
	response := RepositoryResponse{}

	bucket, errorMessage, isError := getCouchBucket(request)
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		for _, obj := range request.Body.Objects {
			key := getNoSqlKeyById(request, obj)
			request.Log("Deleting object from couchbase : " + key)
			err := bucket.Delete(key)
			if err != nil {
				response.IsSuccess = false
				request.Log("Error deleting object from couchbase : " + key + ", " + err.Error())
				response.GetErrorResponse("Error deleting Multiple objects in Couchbase" + err.Error())
			} else {
				response.IsSuccess = true
				response.Message = "Successfully deleted Multiple objects in Coucahbase"
				request.Log(response.Message)
			}
		}

	}

	return response
}

func (repository CouchRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-SINGLE")
	response := RepositoryResponse{}

	bucket, errorMessage, isError := getCouchBucket(request)
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
	response := RepositoryResponse{}
	request.Log("Starting SPECIAL!")
	queryType := request.Body.Special.Type

	switch queryType {
	case "getFields":
		request.Log("Starting GET-FIELDS sub routine!")
		fieldsInByte := executeCouchbaseGetFields(request)
		if fieldsInByte != nil {
			response.IsSuccess = true
			response.Message = "Successfully Retrieved Fileds on Class : " + request.Controls.Class
			response.GetResponseWithBody(fieldsInByte)
		} else {
			response.IsSuccess = false
			response.Message = "Aborted! Unsuccessful Retrieving Fileds on Class : " + request.Controls.Class
			errorMessage := response.Message
			response.GetErrorResponse(errorMessage)
		}
	case "getClasses":
		request.Log("Starting GET-CLASSES sub routine")
		fieldsInByte := executeCouchbaseGetClasses(request)
		if fieldsInByte != nil {
			response.IsSuccess = true
			response.Message = "Successfully Retrieved Fileds on Class : " + request.Controls.Class
			response.GetResponseWithBody(fieldsInByte)
		} else {
			response.IsSuccess = false
			response.Message = "Aborted! Unsuccessful Retrieving Fileds on Class : " + request.Controls.Class
			errorMessage := response.Message
			response.GetErrorResponse(errorMessage)
		}
	case "getNamespaces":
		request.Log("Starting GET-NAMESPACES sub routine")
		fieldsInByte := executeCouchbaseGetNamespaces(request)
		if fieldsInByte != nil {
			response.IsSuccess = true
			response.Message = "Successfully Retrieved All Namespaces"
			response.GetResponseWithBody(fieldsInByte)
		} else {
			response.IsSuccess = false
			response.Message = "Aborted! Unsuccessful Retrieving All Namespaces"
			errorMessage := response.Message
			response.GetErrorResponse(errorMessage)
		}
	case "getSelected":
		request.Log("Starting GET-SELECTED_FIELDS sub routine")
		fieldsInByte := executeCouchbaseGetSelectedFields(request)
		if fieldsInByte != nil {
			response.IsSuccess = true
			response.Message = "Successfully Retrieved All selected Field data"
			response.GetResponseWithBody(fieldsInByte)
		} else {
			response.IsSuccess = false
			response.Message = "Aborted! Unsuccessful Retrieving All selected field data"
			errorMessage := response.Message
			response.GetErrorResponse(errorMessage)
		}
	}

	return response
}

func (repository CouchRepository) Test(request *messaging.ObjectRequest) {

}

func setOne(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	bucket, errorMessage, isError := getCouchBucket(request)
	keyValue := getCouchBaseRecordID(request, nil)
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else if makeCouchBaseIndexing(request) {

		key := request.Controls.Namespace + "." + request.Controls.Class + "." + keyValue
		request.Body.Object[request.Body.Parameters.KeyProperty] = keyValue
		request.Log("Inserting/Updating object in Couchbase : " + key)
		request.Log(key)
		fmt.Println(request.Body.Object)
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
	//Update Response
	var Data []map[string]interface{}
	Data = make([]map[string]interface{}, 1)
	var actualData map[string]interface{}
	actualData = make(map[string]interface{})
	actualData["ID"] = keyValue
	Data[0] = actualData
	response.Data = Data
	return response
}

func setMany(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	bucket, errorMessage, isError := getCouchBucket(request)
	var idData map[string]interface{}
	idData = make(map[string]interface{})
	if isError == true {
		response.GetErrorResponse(errorMessage)
	} else {
		for index, obj := range request.Body.Objects {
			keyValue := getCouchBaseRecordID(request, obj)
			key := request.Controls.Namespace + "." + request.Controls.Class + "." + keyValue
			obj[request.Body.Parameters.KeyProperty] = keyValue
			idData[strconv.Itoa(index)] = keyValue
			err := bucket.Set(key, 0, obj)
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

	}
	//Update Response
	var DataMap []map[string]interface{}
	DataMap = make([]map[string]interface{}, 1)
	var actualInput map[string]interface{}
	actualInput = make(map[string]interface{})
	actualInput["ID"] = idData
	DataMap[0] = actualInput
	response.Data = DataMap
	return response
}

func getCouchBucket(request *messaging.ObjectRequest) (bucket *couchbase.Bucket, errorMessage string, isError bool) {

	isError = false
	request.Log("Getting store configuration settings for Couchbase")

	setting_host := request.Configuration.ServerConfiguration["COUCH"]["Url"]
	setting_bucket := request.Configuration.ServerConfiguration["COUCH"]["Bucket"]
	setting_bucket = getSQLnamespace(request)
	request.Log("Store configuration settings recieved for Couchbase Host : " + setting_host + " , Bucket : " + setting_bucket)

	c, err := couchbase.Connect("http://" + setting_host + "/")
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
		
		err1 := createCouchbaseBucket(setting_host, setting_bucket)
		if !err1 {
			isError = true
			errorMessage = "Error getting/creating Couchbase bucket: " + setting_bucket
			request.Log(errorMessage)
			return bucket, errorMessage, isError
		} else {
			time.Sleep(2 * time.Second)
			//Insert Document Wait for another second to refresh
			status := uploadDesignDocument(setting_host, setting_bucket, getSQLnamespace(request))
			fmt.Println(status)
			//reconnect
			bucket = reconnect(setting_host, setting_bucket)
		}

	} else {
		request.Log("Successfully recieved Couchbase bucket")
		bucket = returnBucket
	}

	return
}

func reconnect(url string, bucketName string) (bucket *couchbase.Bucket) {
	c, err := couchbase.Connect("http://" + url + "/")
	if err != nil {
		fmt.Println(err.Error())
	}

	pool, err := c.GetPool("default")
	if err != nil {
		fmt.Println(err.Error())
	}

	bucket, err = pool.GetBucket(bucketName)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Successfully recieved Couchbase bucket")
	}
	return bucket
}

func createCouchbaseBucket(url string, bucketName string) (status bool) {
	tempUrl := strings.Split(url, ":")
	cluster, err := gocb.Connect("couchbase://" + tempUrl[0] + "")
	clustermgr := cluster.Manager("Administrator", "123456")

	bucketSettings := gocb.BucketSettings{}
	bucketSettings.FlushEnabled = true
	bucketSettings.IndexReplicas = false
	bucketSettings.Name = bucketName
	bucketSettings.Password = ""
	bucketSettings.Quota = 100
	bucketSettings.Replicas = 0
	bucketSettings.Type = 0
	q := &bucketSettings

	err = clustermgr.InsertBucket(q)
	if err != nil {
		fmt.Println("Error creating the new Bucket! : " + err.Error())
		status = false
	} else {
		fmt.Println("Successfully created new Bucket!")
		status = true
	}
	return status
}

func uploadDesignDocument(url string, bucketname string, namespace string) (status bool) {
	designDocumentName := "dev_" + namespace
	tempUrl := strings.Split(url, ":")
	cluster, _ := gocb.Connect("couchbase://" + tempUrl[0] + "")
	buckett, _ := cluster.OpenBucket(bucketname, "")

	bucketMgr := buckett.Manager("Administrator", "123456")

	ddoc := gocb.DesignDocument{}
	ddoc.Name = designDocumentName

	vview := gocb.View{}
	vview.Map = "function(doc,meta){emit(meta.id);}"

	var myView map[string]gocb.View
	myView = make(map[string]gocb.View)

	myView[designDocumentName] = vview

	ddoc.Views = myView

	p := &ddoc

	err := bucketMgr.UpsertDesignDocument(p)
	if err != nil {
		fmt.Println(err.Error())
		status = false
	} else {
		status = true
	}

	return

}

// Helper Functions

func getCouchBaseRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
	isGUIDKey := false
	isAutoIncrementId := false //else MANUAL key from the user

	if obj == nil {
		//single request
		if (request.Controls.Id == "-999") || (request.Body.Parameters.AutoIncrement == true) {
			isAutoIncrementId = true
		}

		if (request.Controls.Id == "-888") || (request.Body.Parameters.GUIDKey == true) {
			isGUIDKey = true
		}

	} else {
		//multiple requests
		if (obj[request.Body.Parameters.KeyProperty].(string) == "-999") || (request.Body.Parameters.AutoIncrement == true) {
			isAutoIncrementId = true
		}

		if (obj[request.Body.Parameters.KeyProperty].(string) == "-888") || (request.Body.Parameters.GUIDKey == true) {
			isGUIDKey = true
		}

	}

	if isGUIDKey {
		request.Log("GUID Key generation requested!")
		returnID = uuid.NewV1().String()
	} else if isAutoIncrementId {
		request.Log("Automatic Increment Key generation requested!")
		bucket, _, isError := getCouchBucket(request)
		if isError {
			returnID = ""
			request.Log("Connecting to Couchbase Failed!")
		} else {
			//read Attributes table

			key := request.Controls.Namespace + "." + request.Controls.Class + "#domainClassAttributes"
			rawBytes, err := bucket.GetRaw(key)

			if err != nil {
				request.Log("This is a freshly created Class. Inserting new Class record.")
				var ObjectBody map[string]interface{}
				ObjectBody = make(map[string]interface{})
				ObjectBody["maxCount"] = "1"
				ObjectBody["version"] = uuid.NewV1().String()
				err = bucket.Set(key, 0, ObjectBody)
				if err != nil {
					request.Log("Update of maxCount Failed")
					returnID = ""
				} else {
					returnID = "1"
				}
			} else {
				var UpdatedCount int
				var returnData map[string]interface{}
				returnData = make(map[string]interface{})

				json.Unmarshal(rawBytes, &returnData)

				for fieldName, fieldvalue := range returnData {
					if strings.ToLower(fieldName) == "maxcount" {
						UpdatedCount, _ = strconv.Atoi(fieldvalue.(string))
						UpdatedCount++
						returnID = strconv.Itoa(UpdatedCount)
						break
					}
				}

				//update the table
				//save to attributes table
				returnData["maxCount"] = returnID
				returnData["version"] = uuid.NewV1().String()
				err = bucket.Set(key, 0, returnData)
				if err != nil {
					request.Log("Update of maxCount Failed")
					returnID = ""
				}

			}

		}
	} else {
		request.Log("Manual Key requested!")
		if obj == nil {
			returnID = request.Controls.Id
		} else {
			returnID = obj[request.Body.Parameters.KeyProperty].(string)
		}
	}

	return
}

func makeCouchBaseIndexing(request *messaging.ObjectRequest) (status bool) {

	bucket, _, isError := getCouchBucket(request)

	if isError == true {
		status = false
		return
	} else {
		//check if new namespace
		//Add if new
		//Ignore if already available
		key := "CouchBase#DomainList"
		rawBytes, err := bucket.GetRaw(key)
		if err != nil {
			request.Log("There is no record.. Creating New!")

			var newData map[string]interface{}
			newData = make(map[string]interface{})

			newData["1"] = request.Controls.Namespace

			err = bucket.Set(key, 0, newData)

			if err != nil {
				request.Log("Error storing data in Couchbase! : " + err.Error())
				status = false
				return
			} else {
				request.Log("Successfully created new CouchBase#DomainList attribute!")
				status = true
			}
		} else {
			request.Log("Record is there... Check if already available")

			var returnData map[string]interface{}
			returnData = make(map[string]interface{})

			json.Unmarshal(rawBytes, &returnData)

			isFound := false
			for _, namespace := range returnData {
				if namespace == request.Controls.Namespace {
					request.Log("Domain already available... Nothing to do")
					isFound = true
					status = true
				}
			}

			//if not found.....
			if !isFound {
				returnData[strconv.Itoa(len(returnData)+1)] = request.Controls.Namespace
				err = bucket.Set(key, 0, returnData)

				if err != nil {
					request.Log("Error storing data in Couchbase. Connection Resetted!")
					status = false
					return
				} else {
					request.Log("Successfully Updated CouchBase#DomainList attribute!")
					status = true
				}
			}

		}

		//check if class is in namespace#DomainClasses is there
		//If not add new record
		//if there.. just ignore

		Classkey := request.Controls.Namespace + "#DomainClasses"
		rawBytes, err2 := bucket.GetRaw(Classkey)
		if err2 != nil {
			request.Log("There is no Class record.. Creating New!")

			var newData2 map[string]interface{}
			newData2 = make(map[string]interface{})

			newData2["1"] = request.Controls.Class

			err2 = bucket.Set(Classkey, 0, newData2)

			if err2 != nil {
				request.Log("Error storing data in Couchbase. Connection Resetted!")
				status = false
				return
			} else {
				request.Log("Successfully created new " + Classkey + " attribute!")
				status = true
			}
		} else {
			request.Log("Record is there... Check if already available")

			var returnData2 map[string]interface{}
			returnData2 = make(map[string]interface{})

			json.Unmarshal(rawBytes, &returnData2)

			isFound := false
			for _, class := range returnData2 {
				if class == request.Controls.Class {
					request.Log("Domain already available... Nothing to do")
					isFound = true
					status = true
				}
			}

			//if not found.....
			if !isFound {
				returnData2[strconv.Itoa(len(returnData2)+1)] = request.Controls.Class
				err2 = bucket.Set(Classkey, 0, returnData2)

				if err2 != nil {
					request.Log("Error storing data in Couchbase. Connection Resetted!")
					status = false
					return
				} else {
					request.Log("Successfully Updated " + Classkey + " attribute!")
					status = true
				}
			}

		}

	}

	return

}

// Sub Routine Functions

func executeCouchbaseGetFields(request *messaging.ObjectRequest) (returnByte []byte) {

	bucket, _, isError := getCouchBucket(request)

	if isError == true {
		returnByte = nil
	} else {
		//Get all IDs
		viewResult, err := bucket.View(("dev_" + getSQLnamespace(request)), ("dev_" + getSQLnamespace(request)), nil)
		if err != nil {
			request.Log("Error fetching result from View")
			returnByte = nil
			return
		}

		//Iterate check for pattern and choose only desired IDs
		pattern := request.Controls.Namespace + "." + request.Controls.Class + "."

		var key string
		for _, row := range viewResult.Rows {
			if strings.Contains((row.Key).(string), pattern) {
				key = (row.Key).(string)
				break
			}
		}

		//read all data for fetched days and return
		rawBytes, err := bucket.GetRaw(key)
		if err != nil {
			request.Log("Error fetching field names from the class")
			returnByte = nil
			return
		}

		var returnData map[string]interface{}
		returnData = make(map[string]interface{})

		var returnFields []string

		json.Unmarshal(rawBytes, &returnData)

		for fieldName, _ := range returnData {
			returnFields = append(returnFields, fieldName)
		}

		returnByte, err = json.Marshal(returnFields)

		if err != nil {
			request.Log("Error fetching field names from the class")
			returnByte = nil
		}
	}
	return
}

func executeCouchbaseGetSelectedFields(request *messaging.ObjectRequest) (returnByte []byte) {

	bucket, _, isError := getCouchBucket(request)

	if isError == true {
		returnByte = nil
	} else {
		//Get all IDs
		viewResult, err := bucket.View(("dev_" + getSQLnamespace(request)), ("dev_" + getSQLnamespace(request)), nil)
		if err != nil {
			returnByte = nil
			request.Log("Couldn't fetch results from the View")
			return
		}

		//Iterate check for pattern and choose only desired IDs
		pattern := request.Controls.Namespace + "." + request.Controls.Class + "."

		var idArray []string
		for _, row := range viewResult.Rows {
			if strings.Contains((row.Key).(string), pattern) {
				idArray = append(idArray, (row.Key).(string))
			}
		}

		//read all data for fetched days and return
		var returnDataMap []map[string]interface{}

		data, err := bucket.GetBulkRaw(idArray)

		for key, _ := range data {
			var tempData map[string]interface{}
			tempData = make(map[string]interface{})
			json.Unmarshal(data[key], &tempData)

			if request.Controls.SendMetaData == "false" {
				delete(tempData, "__osHeaders")
			}

			var addMap map[string]interface{}
			addMap = make(map[string]interface{})

			request.Log("Requested Field List : " + request.Body.Special.Parameters)
			if request.Body.Special.Parameters != "*" {
				requestedFields := strings.Split(request.Body.Special.Parameters, " ")
				for _, fieldName := range requestedFields {
					addMap[fieldName] = tempData[fieldName]
				}
			}

			returnDataMap = append(returnDataMap, addMap)
		}

		returnByte, err = json.Marshal(returnDataMap)

		if err != nil {
			request.Log("Error fetching field names from the class")
			returnByte = nil
		}
	}
	return
}

func executeCouchbaseGetNamespaces(request *messaging.ObjectRequest) (returnByte []byte) {

	bucket, _, isError := getCouchBucket(request)

	if isError == true {
		returnByte = nil
	} else {
		//Get Namespace Details

		key := "CouchBase#DomainList"
		rawBytes, err := bucket.GetRaw(key)
		if err != nil {
			request.Log("Error connection reset. Try again!")
			returnByte = nil
		} else {
			returnByte = rawBytes
		}
	}
	return
}

func executeCouchbaseGetClasses(request *messaging.ObjectRequest) (returnByte []byte) {

	bucket, _, isError := getCouchBucket(request)

	if isError == true {
		returnByte = nil
	} else {
		//Get Classes Details
		key := request.Controls.Namespace + "#DomainClasses"
		rawBytes, err := bucket.GetRaw(key)
		if err != nil {
			request.Log("Error connection reset. Try again!")
			returnByte = nil
		} else {
			returnByte = rawBytes
		}
	}
	return
}
*/
