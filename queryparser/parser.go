package queryparser

import (
	"duov6.com/queryparser/analyzer"
	"duov6.com/queryparser/repositories"
	"duov6.com/queryparser/structs"
	//	"duov6.com/queryparser/repositories"
	"fmt"
	"google.golang.org/cloud/datastore"
)

//This is the main entry point to the query parser

func GetElasticQuery(queryString string, namespace string, class string) (query string, err error) {
	if queryResult, err := getQuery(queryString, "ES", namespace, class); err == nil {
		query = queryResult.(string)
	}
	return
}

func GetDataStoreQuery(queryString string, namespace string, class string) (query *datastore.Query, err error) {
	if queryResult, err := getQuery(queryString, "CDS", namespace, class); err == nil {
		query = queryResult.(*datastore.Query)
	}
	return
}

func GetMsSQLQuery(queryString string, namespace string, class string) (query string, err error) {
	if queryResult, err := getQuery(queryString, "MSSQL", namespace, class); err == nil {
		query = queryResult.(string)
	}
	return
}

func GetCloudSQLQuery(queryString string, namespace string, class string) (query string, err error) {
	if queryResult, err := getQuery(queryString, "CSQL", namespace, class); err == nil {
		query = queryResult.(string)
	}
	return
}

func GetPostgresQuery(queryString string, namespace string, class string) (query string, err error) {
	if queryResult, err := getQuery(queryString, "PSQL", namespace, class); err == nil {
		query = queryResult.(string)
	}
	return
}

func GetMySQLQuery(queryString string, namespace string, class string) (query string, err error) {
	if queryResult, err := getQuery(queryString, "MYSQL", namespace, class); err == nil {
		query = queryResult.(string)
	}
	return
}

func GetHiveQuery(queryString string, namespace string, class string) (query string, err error) {
	if queryResult, err := getQuery(queryString, "HSQL", namespace, class); err == nil {
		query = queryResult.(string)
	}
	return
}

func getQuery(queryString string, repository string, namespace string, class string) (queryResult interface{}, err error) {
	//get type of query
	if queryType := analyzer.GetQueryType(queryString); queryType == "SQL" {
		//Check is valid for preprocessing. Create normalized query
		preparedQuery, err := analyzer.PrepareSQLStatement(queryString, repository, namespace, class)
		if err != nil {
			return queryResult, err
		}

		fmt.Println(preparedQuery)
		//Create Query map from the normalized query
		queryStruct := analyzer.GetQueryMaps(preparedQuery)
		fmt.Println(queryStruct.Operation)
		fmt.Println(queryStruct.SelectedFields)
		fmt.Println(queryStruct.Table)
		fmt.Println(queryStruct.Where)
		fmt.Println(queryStruct.Orderby)

		//Do secondary validation.. for sql keywords
		err = analyzer.ValidateQuery(queryStruct)
		if err != nil {
			fmt.Println(err.Error())
			return "error", err
		}

		queryRequest := structs.RepoRequest{}
		queryRequest.Repository = repository
		queryRequest.Query = preparedQuery
		queryRequest.Queryobject = queryStruct
		queryResult = repositories.Execute(queryRequest)

	} else {
		//reply other query
		queryResult = analyzer.GetOtherQuery(queryString, repository)
	}
	fmt.Println("huehuehue")
	return
}
