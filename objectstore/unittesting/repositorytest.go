package unittesting

import (
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/repositories"
	"fmt"
)

func getRepository() repositories.AbstractRepository {
	var repository repositories.AbstractRepository = repositories.CouchRepository{}
	return repository
}

func RepositoryTest() {
	repository := getRepository()

	//repositoryResponse := repository.GetAll(messaging.ObjectRequest{})
	//fmt.Println(repositoryResponse.Message)

	repository.Test(&messaging.ObjectRequest{})
	fmt.Println("Test Completed!!!")
}
