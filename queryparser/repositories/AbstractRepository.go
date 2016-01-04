package repositories

import (
	"duov6.com/queryparser/structs"
)

type AbstractRepository interface {
	GetName(request structs.RepoRequest) string
	GetQuery(request structs.RepoRequest) structs.RepoResponse
}
