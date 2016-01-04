package queryparser

import (
	"duov6.com/queryparser/analyzer"
	"duov6.com/queryparser/repositories"
	"duov6.com/queryparser/structs"
	"fmt"
	"google.golang.org/cloud/datastore"
)

//This is the main entry point to the query parser

func GetElasticQuery(queryString string, namespace string, class string) (query string, err error) {
	if queryResult, er := getQuery(queryString, "ES", namespace, class); er == nil {
		query = queryResult.(string)
	} else {
		err = er
	}
	return
}

func GetDataStoreQuery(queryString string, namespace string, class string) (query *datastore.Query, err error) {
	if queryResult, er := getQuery(queryString, "CDS", namespace, class); er == nil {
		query = queryResult.(*datastore.Query)
	} else {
		err = er
	}
	return
}

func GetMsSQLQuery(queryString string, namespace string, class string) (query string, err error) {
	if queryResult, er := getQuery(queryString, "MSSQL", namespace, class); er == nil {
		query = queryResult.(string)
	} else {
		err = er
	}
	return
}

func GetCloudSQLQuery(queryString string, namespace string, class string) (query string, err error) {
	if queryResult, er := getQuery(queryString, "CSQL", namespace, class); er == nil {
		query = queryResult.(string)
	} else {
		err = er
	}
	return
}

func GetPostgresQuery(queryString string, namespace string, class string) (query string, err error) {
	if queryResult, er := getQuery(queryString, "PSQL", namespace, class); er == nil {
		query = queryResult.(string)
	} else {
		err = er
	}
	return
}

func GetMySQLQuery(queryString string, namespace string, class string) (query string, err error) {
	if queryResult, er := getQuery(queryString, "MYSQL", namespace, class); er == nil {
		query = queryResult.(string)
	} else {
		err = er
	}
	return
}

func GetHiveQuery(queryString string, namespace string, class string) (query string, err error) {
	if queryResult, er := getQuery(queryString, "HSQL", namespace, class); er == nil {
		query = queryResult.(string)
	} else {
		err = er
	}
	return
}

func getQuery(queryString string, repository string, namespace string, class string) (queryResult interface{}, err error) {
	//get type of query
	if queryType := analyzer.GetQueryType(queryString); queryType == "SQL" {
		fmt.Println("SQL Query Identified!")
		//Check is valid for preprocessing. Create normalized query
		preparedQuery, err := analyzer.PrepareSQLStatement(queryString, repository, namespace, class)
		if err != nil {
			return queryResult, err
		}

		fmt.Println("Prepared Query : " + preparedQuery)
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

		response := repositories.Execute(queryRequest)
		if response.Err != nil {
			err = response.Err
			return response.Query, err
		}

		queryResult = response.Query

	} else {
		//reply other query
		fmt.Println("OTHER")
		queryResult = analyzer.GetOtherQuery(queryString, repository)
	}
	return
}
