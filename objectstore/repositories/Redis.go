package repositories

import (
	"duov6.com/objectstore/messaging"
)

type RedisRepository struct {
}

func (repository RedisRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository RedisRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository RedisRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository RedisRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository RedisRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository RedisRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository RedisRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository RedisRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository RedisRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository RedisRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository RedisRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository RedisRepository) Test(request *messaging.ObjectRequest) {

}
