package repositories

import (
	"duov6.com/objectstore/messaging"
	//"encoding/json"
	//"fmt"
	"github.com/twinj/uuid"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/datastore"
	"io/ioutil"
	//"time"
	//"strconv"
	//"strings"
	"duov6.com/term"
)

type GoogleDataStoreRepository struct {
}

func (repository GoogleDataStoreRepository) GetRepositoryName() string {
	return "GoogleDataStore"
}

func (repository GoogleDataStoreRepository) getConnection(request *messaging.ObjectRequest) (client *datastore.Client, err error, projectID string) {
	dataStoreConfig := request.Configuration.ServerConfiguration["GoogleDataStore"]

	keyFile := dataStoreConfig["KeyFile"]
	projectID = dataStoreConfig["ProjectID"]

	jsonKey, err := ioutil.ReadFile(keyFile)
	if err != nil {
		term.Write(err.Error(), 1)
	} else {
		conf, err := google.JWTConfigFromJSON(
			jsonKey,
			datastore.ScopeDatastore,
			datastore.ScopeUserEmail,
		)
		if err != nil {
			term.Write(err.Error(), 1)
		} else {
			ctx := context.Background()
			client, err = datastore.NewClient(ctx, projectID, cloud.WithTokenSource(conf.TokenSource(ctx)))
			if err != nil {
				term.Write(err.Error(), 1)
			}
		}
	}

	return
}

func (repository GoogleDataStoreRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-ALL")
	response := RepositoryResponse{}

	return response
}

func (repository GoogleDataStoreRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("GetSearch not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository GoogleDataStoreRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-QUERY!")
	response := RepositoryResponse{}
	queryType := request.Body.Query.Type

	switch queryType {
	case "Query":
		if request.Body.Query.Parameters != "*" {
			request.Log("Support for SQL Query not implemented in Cassandra Db repository")
			return getDefaultNotImplemented()
		} else {
			return repository.GetAll(request)
		}
	default:
		request.Log(queryType + " not implemented in Redis Db repository")
		return getDefaultNotImplemented()
	}

	return response
}

func (repository GoogleDataStoreRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting GET-BY-KEY")
	response := RepositoryResponse{}

	return response
}

func (repository GoogleDataStoreRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-MULTIPLE")
	return setManyRedis(request)
}

func (repository GoogleDataStoreRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting INSERT-SINGLE")
	return setOneRedis(request)
}

func (repository GoogleDataStoreRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-MULTIPLE")
	return setManyRedis(request)
}

func (repository GoogleDataStoreRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting UPDATE-SINGLE")
	return setOneRedis(request)
}

func (repository GoogleDataStoreRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-MULTIPLE")
	response := RepositoryResponse{}
	return response
}

func (repository GoogleDataStoreRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Starting DELETE-SINGLE")
	response := RepositoryResponse{}

	return response
}

func (repository GoogleDataStoreRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Special not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository GoogleDataStoreRepository) Test(request *messaging.ObjectRequest) {

}

func (repository GoogleDataStoreRepository) getRecordID(request *messaging.ObjectRequest, obj map[string]interface{}) (returnID string) {
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
		returnID = uuid.NewV1().String()
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
