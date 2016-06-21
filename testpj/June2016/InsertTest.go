package main

import (
	"database/sql"
	"duov6.com/common"
	"encoding/json"
	"fmt"
	"github.com/go-martini/martini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/martini-contrib/cors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {

	StartMartini()

}

func StartMartini() {

	fmt.Println("Insert Tester Listening on Port : 5775")
	m := martini.Classic()
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"securityToken", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	m.Post("/:namespace/:class", handleRequest)

	m.RunOnAddr(":5775")
}

func handleRequest(params martini.Params, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request Recieved")
	responseMessage, isSuccess := InsertData()

	if isSuccess {
		w.WriteHeader(200)

	} else {
		w.WriteHeader(500)
	}

	fmt.Fprintf(w, "%s", responseMessage)
}

func InsertData() (responseMessage string, isSuccess bool) {

	if !isSchemaValidated {
		validateSchema()
	}
	id := common.GetGUID()
	storeQuery := "INSERT INTO inserttestdb.testplan (id,one,two,three,four,five,six,seven,eight,nine) VALUES ('" + id + "','1','1','1','1','1','1','1','1','1');"
	conn, _ := getConnection()
	err := executeNonQuery(conn, storeQuery)
	if err != nil {
		isSuccess = false
		responseMessage = err.Error()
	} else {
		isSuccess = true
		responseMessage = "Successfully Inserted!"
	}

	return
}

var isSchemaValidated bool
var connection *sql.DB

func validateSchema() {
	fmt.Println("Validating Schema")
	dbQuery := "CREATE DATABASE IF NOT EXISTS InsertTestDB;"
	tableQuery := "create table IF NOT EXISTS InsertTestDB.testPlan (id VARCHAR(255) primary key, one text, two text, three text, four text, five text, six text, seven text, eight text, nine text);"

	conn, _ := getConnection()

	err := executeNonQuery(conn, dbQuery)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	} else {
		err = executeNonQuery(conn, tableQuery)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(0)
		} else {
			isSchemaValidated = true
		}
	}
}

func executeNonQuery(conn *sql.DB, query string) (err error) {
	_, err = conn.Exec(query)
	if err == nil {
		//do nothing mate
	} else {
		fmt.Println(err.Error())
	}

	return
}

func getConnection() (conn *sql.DB, err error) {

	mysqlConf := GetSettings()

	username := mysqlConf["Username"]
	password := mysqlConf["Password"]
	url := mysqlConf["Url"]
	port := mysqlConf["Port"]
	IdleLimit := -1
	OpenLimit := 100
	TTL := 5

	if mysqlConf["IdleLimit"] != "" {
		IdleLimit, err = strconv.Atoi(mysqlConf["IdleLimit"])
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	if mysqlConf["OpenLimit"] != "" {
		OpenLimit, err = strconv.Atoi(mysqlConf["OpenLimit"])
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	if mysqlConf["TTL"] != "" {
		TTL, err = strconv.Atoi(mysqlConf["TTL"])
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	if connection == nil {
		conn, err = CreateConnection(username, password, url, port, IdleLimit, OpenLimit, TTL)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		connection = conn
	} else {
		if connection.Ping(); err != nil {
			_ = connection.Close()
			connection = nil
			conn, err = CreateConnection(username, password, url, port, IdleLimit, OpenLimit, TTL)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			connection = conn
		} else {
			conn = connection
		}
	}
	return conn, err
}

func CreateConnection(username, password, url, port string, IdleLimit, OpenLimit, TTL int) (conn *sql.DB, err error) {
	conn, err = sql.Open("mysql", username+":"+password+"@tcp("+url+":"+port+")/")
	conn.SetMaxIdleConns(IdleLimit)
	conn.SetMaxOpenConns(OpenLimit)
	conn.SetConnMaxLifetime(time.Duration(TTL) * time.Minute)
	return
}

func GetSettings() (object map[string]string) {
	content, _ := ioutil.ReadFile("settings.config")
	object = make(map[string]string)
	_ = json.Unmarshal(content, &object)
	return
}
