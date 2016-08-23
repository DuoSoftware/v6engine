package main

import (
	//"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
)

func main() {
	keyspace := "system"
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = keyspace

	conn, _ := cluster.CreateSession()
	query := "select osheaders from db_test.lod;"
	iter := conn.Query(query).Iter()
	resultSet, err := iter.SliceMap()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for _, kk := range resultSet {
			fmt.Println(string(kk["osheaders"].([]byte)))
		}
	}
}
