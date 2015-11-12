package main

import (
	"fmt"
	"github.com/derekgr/hivething"
)

func main() {
	db, err := hivething.Connect("159.203.73.174:10000", hivething.DefaultOptions)
	if err != nil {
		// handle
	}
	defer db.Close()

	results, err := db.Query("SHOW DATABASES")
	if err != nil {
		// handle
	}

	status, err := results.Wait()
	if err != nil {
		// handle
	}

	if status.IsSuccess() {
		var tableName string
		for results.Next() {
			results.Scan(&tableName)
			fmt.Println(tableName)
		}
	} else {
		// handle status.Error
	}
}
