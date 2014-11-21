package repositories

import (
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"github.com/mattbaird/elastigo/lib"
	"time"
)

type ElasticRepository struct {
}

func (repository ElasticRepository) GetRepositoryName() string {
	return "Elastic Search"
}

func (repository ElasticRepository) GetAll(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("GetAll not implemented in Elastic Search repository")
	return getDefaultNotImplemented()
}

func (repository ElasticRepository) GetSearch(request *messaging.ObjectRequest) RepositoryResponse {
	return search(request, request.Body.Query.Parameters)
}

func search(request *messaging.ObjectRequest, searchStr string) RepositoryResponse {
	response := RepositoryResponse{}
	conn := getConnection()(request)

	query := "{\"query\":{\"query_string\" : {\"query\" : \"" + searchStr + "\"}}}"

	data, err := conn.Search(request.Controls.Class, request.Controls.Class, nil, query)

	if err != nil {
		errorMessage := "Error retrieving object from elastic search : " + err.Error()
		request.Log(errorMessage)
		request.Log("Error Query : " + query)
		response.GetErrorResponse(errorMessage)
	} else {
		request.Log("Successfully retrieved object from Elastic Search")

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

	response := RepositoryResponse{}
	conn := getConnection()(request)

	key := getNoSqlKey(request)
	request.Log("Elastic Search Get By Key : " + key)
	data, err := conn.Get(request.Controls.Class, request.Controls.Class, key, nil)

	if err != nil {
		errorMessage := "Error retrieving object from elastic search : " + err.Error()
		request.Log(errorMessage)
		response.GetErrorResponse(errorMessage)
	} else {
		request.Log("Successfully retrieved object from Elastic Search")
		bytes, err := data.Source.MarshalJSON()
		if err != nil {
			errorMessage := "Elastic search JSON marshal error : " + err.Error()
			request.Log(errorMessage)
			response.GetErrorResponse(errorMessage)
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
	response := RepositoryResponse{}

	conn := getConnection()(request)

	key := getNoSqlKey(request)
	request.Log("Inserting single object to Elastic Search : " + key)
	_, err := conn.Index(request.Controls.Class, request.Controls.Class, key, nil, request.Body.Object)

	if err != nil {
		errorMessage := "Elastic Search Single Insert Error : " + err.Error()
		request.Log(errorMessage)
		response.GetErrorResponse(errorMessage)
	} else {
		response.IsSuccess = true
		response.Message = "Successfully inserted one to elastic search"
		request.Log(response.Message)
	}

	return response
}

func setManyElastic(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	conn := getConnection()(request)

	request.Log("Starting Elastic Search bulk insert")
	indexer := conn.NewBulkIndexer(100)

	nowTime := time.Now()
	for _, obj := range request.Body.Objects {
		nosqlid := getNoSqlKeyById(request, obj)
		indexer.Index(request.Controls.Class, request.Controls.Class, nosqlid, "10", &nowTime, obj, false)
	}

	indexer.Start()
	numerrors := indexer.NumErrors()

	if numerrors != 0 {
		request.Log("Elastic Search bulk insert error")
		response.GetErrorResponse("Elastic Search bulk insert error")
	} else {
		response.IsSuccess = true
		response.Message = "Successfully inserted bulk to Elastic Search"
		request.Log(response.Message)
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
	request.Log("DeleteMultiple not implemented in Elastic Search repository")
	return getDefaultNotImplemented()
}

func (repository ElasticRepository) DeleteSingle(request *messaging.ObjectRequest) RepositoryResponse {
	response := RepositoryResponse{}

	conn := getConnection()(request)

	key := getNoSqlKey(request)
	request.Log("Deleting single object from Elastic Search : " + key)
	_, err := conn.Delete(request.Controls.Class, request.Controls.Class, getNoSqlKey(request), nil)

	if err != nil {
		errorMessage := "Elastic Search single delete error : " + err.Error()
		request.Log(errorMessage)
		response.GetErrorResponse(errorMessage)
	} else {
		response.IsSuccess = true
		response.Message = "Successfully deleted one in elastic search"
		request.Log(response.Message)
	}

	return response
}

func (repository ElasticRepository) Special(request *messaging.ObjectRequest) RepositoryResponse {
	request.Log("Special not implemented in Elastic Search repository")
	return getDefaultNotImplemented()
}

func (repository ElasticRepository) Test(request *messaging.ObjectRequest) {

}

func getConnection() func(request *messaging.ObjectRequest) *elastigo.Conn {

	var connection *elastigo.Conn

	return func(request *messaging.ObjectRequest) *elastigo.Conn {
		if connection == nil {
			host := request.Configuration.ServerConfiguration["ELASTIC"]["Host"]
			port := request.Configuration.ServerConfiguration["ELASTIC"]["Port"]
			request.Log("Establishing new connection for Elastic Search " + host + ":" + port)

			conn := elastigo.NewConn()
			conn.SetHosts([]string{host})
			conn.Port = port
			connection = conn
		}

		request.Log("Reusing existing Elastic Search connection ")
		return connection
	}
}
