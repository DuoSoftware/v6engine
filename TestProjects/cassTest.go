package main

import (
	//"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
	"reflect"
)

func main() {
	keyspace := "system"
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = keyspace

	conn, _ := cluster.CreateSession()
	query := "select column_name,validator from system.schema_columns WHERE keyspace_name='ddf' AND columnfamily_name='hue';"
	iter := conn.Query(query).Iter()
	resultSet, err := iter.SliceMap()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(resultSet)
		fmt.Println(reflect.TypeOf(resultSet[0]["validator"]))
	}
}
