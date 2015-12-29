package repositories

import (
	"duov6.com/queryparser/structs"
)

type CommonSQL struct {
}

func (repository CommonSQL) GetName(request structs.RepoRequest) string {
	return ("CommonSQL : " + request.Repository)
}

func (repository CommonSQL) GetQuery(request structs.RepoRequest) interface{} {
	var result interface{}
	result = request.Query
	return result
}
