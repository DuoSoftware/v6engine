package main

import (
	"database/sql"
	"duov6.com/common"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

func getConnection() (conn *sql.DB, err error) {

	var c *sql.DB
	c, err = sql.Open("mysql", "root"+":"+"DuoS123"+"@tcp("+"192.168.1.194"+":"+"3306"+")/")
	c.SetMaxIdleConns(100)
	c.SetMaxOpenConns(0)
	c.SetConnMaxLifetime(time.Duration(2) * time.Minute)
	conn = c

	return conn, err
}

func main() {

	conn, err := getConnection()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for x := 0; x < 10000; x++ {
		GUID := common.GetGUID()
		query := "insert into huehuehue.one values ('" + GUID + "', 'name', 'age', 'address', 'school', 'pet', 'mom', 'dad', 'major', 'book', 'hobby');"
		executeNonQuery(conn, query)
		fmt.Print("Inserting : ")
		fmt.Println(x)
	}

	conn.Close()
}

func executeNonQuery(conn *sql.DB, query string) {

	var stmt *sql.Stmt
	stmt, err := conn.Prepare(query)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	_, err = stmt.Exec()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	_ = stmt.Close()
}
