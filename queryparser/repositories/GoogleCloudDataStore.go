package repositories

import (
	"duov6.com/queryparser/structs"
	"fmt"
	"google.golang.org/cloud/datastore"
	"reflect"
)

type GoogleCloudDataStore struct {
}

func (repository GoogleCloudDataStore) GetName(request structs.RepoRequest) string {
	return "GoogleCloudDataStore"
}

func (repository GoogleCloudDataStore) GetQuery(request structs.RepoRequest) structs.RepoResponse {
	response := structs.RepoResponse{}
	var query *datastore.Query
	query = datastore.NewQuery("huehuehue")
	query = repository.GetFields(request, query)
	response.Query = query
	return response
}

func (repository GoogleCloudDataStore) GetOrderBy(request structs.RepoRequest, queryIn *datastore.Query) (queryOut *datastore.Query) {
	queryOut = queryIn
	orderBy := request.Queryobject.Orderby

	for order, field := range orderBy {
		if order == "ASC" {
			queryOut = queryOut.Order(field)
		} else {
			queryOut = queryOut.Order(("-" + field))
		}
	}
	return
}

func (repository GoogleCloudDataStore) GetFields(request structs.RepoRequest, queryIn *datastore.Query) (queryOut *datastore.Query) {
	queryOut = queryIn
	//m := make([]string, 2)
	//m[0] = "111"
	m := "222"
	in := []reflect.Value{reflect.ValueOf(m)}
	fmt.Println(reflect.ValueOf(queryIn))
	return_values := reflect.ValueOf(queryIn).MethodByName("Project").Call(in)
	fmt.Println(return_values[0].Kind())
	return queryOut
}
