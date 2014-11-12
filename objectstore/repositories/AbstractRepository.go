package repositories

import (
	"duov6.com/objectstore/messaging"
)

type AbstractRepository interface {
	GetAll(request *messaging.ObjectRequest) RepositoryResponse
	GetSearch(request *messaging.ObjectRequest) RepositoryResponse
	GetQuery(request *messaging.ObjectRequest) RepositoryResponse
	GetByKey(request *messaging.ObjectRequest) RepositoryResponse
	InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse
	InsertSingle(request *messaging.ObjectRequest) RepositoryResponse
	UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse
	UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse
	DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse
	DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse
	Special(request *messaging.ObjectRequest) RepositoryResponse
	Test(request *messaging.ObjectRequest)
}
