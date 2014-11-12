package repositories

import (
	"duov6.com/objectstore/messaging"
)

type MongoRepository struct {
}

func (repository MongoRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository MongoRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository MongoRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository MongoRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository MongoRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository MongoRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository MongoRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository MongoRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository MongoRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository MongoRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository MongoRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository MongoRepository) Test(request *messaging.ObjectRequest) {

}
