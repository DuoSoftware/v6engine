package repositories

import (
	"duov6.com/queryparser/structs"
)

type GoogleCloudDataStore struct {
}

func (repository GoogleCloudDataStore) GetName(request structs.RepoRequest) string {
	return "GoogleCloudDataStore"
}

func (repository GoogleCloudDataStore) GetQuery(request structs.RepoRequest) interface{} {
	var result interface{}
	return result
}
