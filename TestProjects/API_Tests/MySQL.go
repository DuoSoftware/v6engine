package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

func main() {
	fmt.Println("1")
	conn, err := CreateConnection("root", "DuoS123", "localhost", "3306", -1, 100, 10)

	query := `call _domainname_report_db3.ReportMaker1();`

	// 	query = `CREATE TABLE IF NOT EXISTS _domainname_report_db3.ProfileMaster
	// (
	// profileID int,
	// billingAddress blob,
	// firstName varchar(255),
	// lastName varchar(255),
	// phone varchar(255),
	// mobile varchar(255),
	// PRIMARY KEY (profileID)
	// );`

	result, err := conn.Exec(query)
	if err == nil {
		fmt.Println("YAY")
		fmt.Println(result)
	} else {
		fmt.Println("NAY")
		fmt.Println(err.Error())
	}

}

func CreateConnection(username, password, url, port string, IdleLimit, OpenLimit, TTL int) (conn *sql.DB, err error) {
	conn, err = sql.Open("mysql", username+":"+password+"@tcp("+url+":"+port+")/")
	conn.SetMaxIdleConns(IdleLimit)
	conn.SetMaxOpenConns(OpenLimit)
	conn.SetConnMaxLifetime(time.Duration(TTL) * time.Minute)
	return
}
