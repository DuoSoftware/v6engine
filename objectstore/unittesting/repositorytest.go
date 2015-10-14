package unittesting

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
	"fmt"
)

func getRepository() repositories.AbstractRepository {
	//commented bcs COUCH no longer implements AbstractRepository methods
	//var repository repositories.AbstractRepository = repositories.CouchRepository{}
	var repository repositories.AbstractRepository = repositories.ElasticRepository{}
	return repository
}

func RepositoryTest() {
	repository := getRepository()

	//repositoryResponse := repository.GetAll(messaging.ObjectRequest{})
	//fmt.Println(repositoryResponse.Message)

	repository.Test(&messaging.ObjectRequest{})
	fmt.Println("Test Completed!!!")
}
