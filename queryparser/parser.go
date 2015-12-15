package queryparser

import (
	"duov6.com/queryparser/analyzer"
	//	"duov6.com/queryparser/repositories"
	"errors"
	"fmt"
	"google.golang.org/cloud/datastore"
)

//This is the main entry point to the query parser

func GetElasticQuery(queryString string, namespace string, class string) (query string, err error) {
	query, _, err = getQuery(queryString, "ES", namespace, class)
	return
}

func GetDataStoreQuery(queryString string, namespace string, class string) (query *datastore.Query, err error) {
	_, query, err = getQuery(queryString, "CDS", namespace, class)
	return
}

func GetMsSQLQuery(queryString string, namespace string, class string) (query string, err error) {
	query, _, err = getQuery(queryString, "MSSQL", namespace, class)
	return
}

func GetCloudSQLQuery(queryString string, namespace string, class string) (query string, err error) {
	query, _, err = getQuery(queryString, "CSQL", namespace, class)
	return
}

func GetPostgresQuery(queryString string, namespace string, class string) (query string, err error) {
	query, _, err = getQuery(queryString, "PSQL", namespace, class)
	return
}

func GetMySQLQuery(queryString string, namespace string, class string) (query string, err error) {
	query, _, err = getQuery(queryString, "MYSQL", namespace, class)
	return
}

func GetHiveQuery(queryString string, namespace string, class string) (query string, err error) {
	query, _, err = getQuery(queryString, "HSQL", namespace, class)
	return
}

func getQuery(queryString string, repository string, namespace string, class string) (query string, queryItem *datastore.Query, err error) {
	//get type of query
	if queryType := analyzer.GetQueryType(queryString); queryType == "SQL" {

		//Split query to maps
		preparedQuery, basicValidation := analyzer.PrepareSQLStatement(queryString, repository, namespace, class)
		if !basicValidation {
			err = errors.New("GROUP BY queries are not allowed!")
			return
		}
		fmt.Println(preparedQuery)

		/*//Get Query struct
		queryStruct := analyzer.GetQueryMaps(preparedQuery)
		fmt.Println(queryStruct)

		//Do secondary validation.. for sql keywords
		err = analyzer.ValidateQuery(queryStruct)

		if err != nil {
			fmt.Println(err.Error())
			return
		}*/
	} else {
		//reply other query
		query, queryItem = analyzer.GetOtherQuery(queryString, repository)
	}

	return
}
