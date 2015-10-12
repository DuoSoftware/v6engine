package storageengines

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
)

type AbstractStorageEngine interface {
	Store(request *messaging.ObjectRequest) (response repositories.RepositoryResponse)
}
