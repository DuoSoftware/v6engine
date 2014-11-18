package repositories

import (
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"fmt"
	"github.com/mattbaird/elastigo/lib"
	"time"
)

type ElasticRepository struct {
}

func (repository ElasticRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository ElasticRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	return search(request, request.Body.Query.Parameters)
}

func search(request *messaging.ObjectRequest, searchStr string) RepositoryResponse {
	fmt.Println("Elastic Search Get By Key : " + request.Controls.Id)
	response := RepositoryResponse{}
	conn := getConnection()(request)

	query := "{\"query\":{\"query_string\" : {\"query\" : \"" + searchStr + "\"}}}"

	data, err := conn.Search(request.Controls.Class, request.Controls.Class, nil, query)

	if err != nil {
		fmt.Println(err.Error())
		response.GetErrorResponse("Error retrieving object from elastic search : " + err.Error())
	} else {
		var allMaps []map[string]interface{}
		allMaps = make([]map[string]interface{}, data.Hits.Len())

		for index, hit := range data.Hits.Hits {
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})

			byteData, _ := hit.Source.MarshalJSON()

			json.Unmarshal(byteData, &currentMap)

			allMaps[index] = currentMap
		}

		finalBytes, _ := json.Marshal(allMaps)
		response.GetResponseWithBody(finalBytes)
	}

	return response
}

func (repository ElasticRepository) GetQuery(request *messaging.ObjectRequest) RepositoryResponse {
	return search(request, request.Body.Query.Parameters)
}

func (repository ElasticRepository) GetByKey(request *messaging.ObjectRequest) RepositoryResponse {
	fmt.Println("Elastic Search Get By Key : " + request.Controls.Id)
	response := RepositoryResponse{}
	conn := getConnection()(request)

	data, err := conn.Get(request.Controls.Class, request.Controls.Class, getNoSqlKey(request), nil)

	if err != nil {
		response.GetErrorResponse("Error retrieving object from elastic search : " + err.Error())
	} else {
		bytes, err := data.Source.MarshalJSON()
		if err != nil {
			response.GetErrorResponse("Elastic search JSON marshal error : " + err.Error())
		} else {
			response.GetResponseWithBody(bytes)
		}

	}

	return response
}

func (repository ElasticRepository) InsertMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return setManyElastic(request)
}

func (repository ElasticRepository) InsertSingle(request *messaging.ObjectRequest) RepositoryResponse {
	return setOneElastic(request)
}

func setOneElastic(request *messaging.ObjectRequest) RepositoryResponse {
	fmt.Println("elastic search single insert")
	response := RepositoryResponse{}

	conn := getConnection()(request)

	_, err := conn.Index(request.Controls.Class, request.Controls.Class, getNoSqlKey(request), nil, request.Body.Object)

	if err != nil {
		response.GetErrorResponse("Elastic Search Single Insert Error : " + err.Error())
	} else {
		response.IsSuccess = true
		response.Message = "Successfully inserted one to elastic search"
	}

	return response
}

func setManyElastic(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	conn := getConnection()(request)

	indexer := conn.NewBulkIndexer(100)

	nowTime := time.Now()
	for _, obj := range request.Body.Objects {
		nosqlid := getNoSqlKeyById(request, obj)
		fmt.Println(nosqlid)
		indexer.Index(request.Controls.Class, request.Controls.Class, nosqlid, "10", &nowTime, obj, false)
	}

	indexer.Start()
	numerrors := indexer.NumErrors()

	if numerrors != 0 {
		response.GetErrorResponse("Elastic Search bulk insert error")
	} else {
		response.IsSuccess = true
		response.Message = "Successfully inserted bulk to Elastic Search"
	}

	return response
}

func (repository ElasticRepository) UpdateMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return setManyElastic(request)
}

func (repository ElasticRepository) UpdateSingle(request *messaging.ObjectRequest) RepositoryResponse {
	return setOneElastic(request)
}

func (repository ElasticRepository) DeleteMultiple(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository ElasticRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	fmt.Println("elastic search single insert")
	response := RepositoryResponse{}

	conn := getConnection()(request)

	_, err := conn.Delete(request.Controls.Class, request.Controls.Class, getNoSqlKey(request), nil)

	if err != nil {
		response.GetErrorResponse("Elastic Search single delete error : " + err.Error())
	} else {
		response.IsSuccess = true
		response.Message = "Successfully deleted one in elastic search"
	}

	return response
}

func (repository ElasticRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	return getDefaultNotImplemented()
}

func (repository ElasticRepository) Test(request *messaging.ObjectRequest) {

}

func getConnection() func(request *messaging.ObjectRequest) *elastigo.Conn {

	var connection *elastigo.Conn

	return func(request *messaging.ObjectRequest) *elastigo.Conn {
		if connection == nil {
			conn := elastigo.NewConn()
			conn.SetHosts([]string{"localhost"})
			conn.Port = "9200"
			connection = conn
		}

		return connection
	}
}
