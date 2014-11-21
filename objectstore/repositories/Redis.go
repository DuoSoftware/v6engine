package repositories

import (
	"duov6.com/objectstore/messaging"
)

type RedisRepository struct {
}

func (repository RedisRepository) GetRepositoryName() string {
	return "Redis"
}

func (repository RedisRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("GetAll not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("GetSearch not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("GetQuery not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("GetByKey not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("InsertMultiple not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("InsertSingle not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("UpdateMultiple not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("UpdateSingle not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("DeleteMultiple not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("DeleteSingle not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Special not implemented in Redis repository")
	return getDefaultNotImplemented()
}

func (repository RedisRepository) Test(request *messaging.ObjectRequest) {

}
