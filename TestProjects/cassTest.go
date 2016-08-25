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
	query := "INSERT INTO gg.wp (firstname, lastname, age, email, city) VALUES ('John', 'Smith', 50, 'johnsmith@email.com', 'Sacramento');"
	//iter := conn.Query(query).Iter()
	//resultSet, err := iter.SliceMap()
	err := conn.Query(query).Exec()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("yay")
		// for _, kk := range resultSet {
		// 	fmt.Println(string(kk["osheaders"].([]byte)))
		// }
	}
}
