package repositories

import (
	"duov6.com/queryparser/structs"
	//"fmt"
)

func Execute(request structs.RepoRequest) structs.RepoResponse {
	result := structs.RepoResponse{}
	var repository AbstractRepository
	repository = Create(request.Repository)
	//fmt.Println("Executing : " + repository.GetName(request))
	result = repository.GetQuery(request)
	return result
}
