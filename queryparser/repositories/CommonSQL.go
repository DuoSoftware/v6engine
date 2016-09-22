package repositories

import (
	"duov6.com/queryparser/structs"
)

type CommonSQL struct {
}

func (repository CommonSQL) GetName(request structs.RepoRequest) string {
	return ("CommonSQL : " + request.Repository)
}

func (repository CommonSQL) GetQuery(request structs.RepoRequest) structs.RepoResponse {
	response := structs.RepoResponse{}
	response.Query = request.Query
	return response
}
