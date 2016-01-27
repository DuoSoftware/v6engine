package repositories

import (
	"duov6.com/objectstore/messaging"
	//"fmt"
)

func Execute(request *messaging.ObjectRequest, repository AbstractRepository) (response RepositoryResponse) {

	switch request.Controls.Operation { //CREATE, READ, UPDATE, DELETE, SPECIAL
	case "insert":
		//fmt.Println("Insert")
		if request.Controls.Multiplicity == "single" {
			//fmt.Println("Single")
			response = repository.InsertSingle(request)
		} else {
			response = repository.InsertMultiple(request)
		}

	case "read-all":
		response = repository.GetAll(request)
	case "read-key":
		response = repository.GetByKey(request)
	case "read-keyword":
		response = repository.GetSearch(request)
	case "read-filter":
		response = repository.GetQuery(request)

	case "update":
		if request.Controls.Multiplicity == "single" {
			response = repository.UpdateSingle(request)
		} else {
			response = repository.UpdateMultiple(request)
		}
	case "delete":
		if request.Controls.Multiplicity == "single" {
			response = repository.DeleteSingle(request)
		} else {
			response = repository.DeleteMultiple(request)
		}
	case "special":
		response = repository.Special(request)
	}

	return
}
