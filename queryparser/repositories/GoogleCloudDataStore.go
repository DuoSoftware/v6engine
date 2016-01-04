package repositories

import (
	"duov6.com/queryparser/structs"
	"google.golang.org/cloud/datastore"
)

type GoogleCloudDataStore struct {
}

func (repository GoogleCloudDataStore) GetName(request structs.RepoRequest) string {
	return "GoogleCloudDataStore"
}

func (repository GoogleCloudDataStore) GetQuery(request structs.RepoRequest) structs.RepoResponse {
	response := structs.RepoResponse{}
	response.Query = datastore.NewQuery(request.Query)
	return response
}
