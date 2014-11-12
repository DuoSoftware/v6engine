package storageengines

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
)

type SingleStorageEngine struct {
}

func (r SingleStorageEngine) Store(request *messaging.ObjectRequest) (response repositories.RepositoryResponse) {
	x := repositories.RepositoryResponse{}
	return x
}
