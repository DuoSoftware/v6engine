package queryparser

import (
	"fmt"
	"google.golang.org/cloud/datastore"
)

//This is the main entry point to the query parser

func GetElasticQuery(queryString string) (query string) {
	query, _ = getQuery(queryString, "ES")
	return
}

func GetDataStoreQuery(queryString string) (query *datastore.Query) {
	_, query = getQuery(queryString, "CDS")
	return
}

func GetMsSQLQuery(queryString string) (query string) {
	query, _ = getQuery(queryString, "MSSQL")
	return
}

func GetCloudSQLQuery(queryString string) (query string) {
	query, _ = getQuery(queryString, "CSQL")
	return
}

func GetPostgresQuery(queryString string) (query string) {
	query, _ = getQuery(queryString, "PSQL")
	return
}

func GetMySQLQuery(queryString string) (query string) {
	query, _ = getQuery(queryString, "MYSQL")
	return
}

func GetHiveQuery(queryString string) (query string) {
	query, _ = getQuery(queryString, "HSQL")
	return
}

func getQuery(queryString string, repository string) (query string, queryItem *datastore.Query) {
	//get type of query
}
