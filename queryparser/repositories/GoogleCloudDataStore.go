package repositories

import (
	"duov6.com/queryparser/structs"
	"google.golang.org/cloud/datastore"
	"reflect"
	"strings"
)

type GoogleCloudDataStore struct {
}

func (repository GoogleCloudDataStore) GetName(request structs.RepoRequest) string {
	return "GoogleCloudDataStore"
}

func (repository GoogleCloudDataStore) GetQuery(request structs.RepoRequest) structs.RepoResponse {
	response := structs.RepoResponse{}
	var query *datastore.Query
	query = datastore.NewQuery(request.Queryobject.Table)
	query = repository.GetFields(request, query)
	query = repository.GetOrderBy(request, query)
	query = repository.GetWhereClauses(request, query)
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

	if len(request.Queryobject.SelectedFields) == 1 && request.Queryobject.SelectedFields[0] == "*" {
		queryOut = queryIn
	} else {
		projectArgs := make([]reflect.Value, len(request.Queryobject.SelectedFields))

		for key, value := range request.Queryobject.SelectedFields {
			projectArgs[key] = reflect.ValueOf(value)
		}

		project_return_values := reflect.ValueOf(queryIn).MethodByName("Project").Call(projectArgs)
		queryOut = project_return_values[0].Interface().(*datastore.Query)
	}
	return queryOut
}

func (repository GoogleCloudDataStore) GetWhereClauses(request structs.RepoRequest, queryIn *datastore.Query) (queryOut *datastore.Query) {
	queryOut = queryIn
	if len(request.Queryobject.Where) != 0 {

		for x := 0; x < len(request.Queryobject.Where); x++ {
			if status := repository.checkIfCombiner(request.Queryobject.Where[x]); status {
				//ignore since AND condition is DEFAULT combiner in Cloud Datastore
			} else {
				//check if complex operator... if not process as simple if condition
				complexPattern := "BETWEEN-NOTBETWEEN-NOTIN"
				if strings.Contains(complexPattern, request.Queryobject.Where[x][1]) {
					queryOut = repository.processComplexOperatorString(request.Queryobject.Where[x], queryOut)
				} else {
					queryOut = repository.processNonComplexOperatorString(request.Queryobject.Where[x], queryOut)
				}
			}
		}
	} else {
		queryOut = queryIn
	}

	return queryOut
}

func (repository GoogleCloudDataStore) checkIfCombiner(arr []string) (status bool) {
	//checks if this is logic combine..  AND / OR
	status = false
	if len(arr) == 1 {
		if strings.EqualFold(arr[0], "AND") {
			status = true
		}
	}
	return status
}

func (repository GoogleCloudDataStore) getNormalizedArray(input []string) (output []string) {
	output = make([]string, len(input))

	for x := 0; x < len(input); x++ {
		output[x] = strings.Replace(input[x], "'", "", -1)
	}

	return output
}

func (repository GoogleCloudDataStore) processComplexOperatorString(arr []string, input *datastore.Query) (output *datastore.Query) {
	output = input
	arr = repository.getNormalizedArray(arr)
	operator := arr[1]
	switch operator {
	case "BETWEEN":
		output = output.Filter((arr[0]+" >="), arr[2]).Filter((arr[0] + " <="), arr[4])
		break
	case "NOTBETWEEN":
		output = output.Filter((arr[0]+" <"), arr[2]).Filter((arr[0] + " >"), arr[4])
		break
	case "NOTIN":
		for x := 2; x < len(arr); x++ {
			value := strings.Replace(arr[x], "'", "", -1)
			output = output.Filter((arr[0]+" >"), value).Filter((arr[0] + " <"), value)
		}
		break
	default:
		//do nothing
		output = input
		break
	}
	return output
}

func (repository GoogleCloudDataStore) processNonComplexOperatorString(arr []string, input *datastore.Query) (output *datastore.Query) {
	arr = repository.getNormalizedArray(arr)
	operator := arr[1]
	switch operator {
	case "=":
		output = input.Filter((arr[0] + " ="), arr[2])
		break
	case "!=":
		output = input.Filter((arr[0]+" >"), arr[2]).Filter((arr[0] + " <"), arr[2])
		break
	case ">":
		output = input.Filter((arr[0] + " >"), arr[2])
		break
	case ">=":
		output = input.Filter((arr[0] + " >="), arr[2])
		break
	case "<":
		output = input.Filter((arr[0] + " <"), arr[2])
		break
	case "<=":
		output = input.Filter((arr[0] + " <="), arr[2])
		break
	default:
		//do nothing..
		output = input
		break
	}

	return output
}
