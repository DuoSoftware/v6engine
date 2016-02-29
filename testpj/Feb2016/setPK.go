package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
)

// This program will change the tables primary key TEXT type to varchar and set __os_id as the primary key.

func main() {
	fmt.Println("Staring Getting Database List!")
	conn, err := getConnection()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	dbs := getDBs(conn)
	fmt.Println(dbs)

	for _, db := range dbs {
		tables := getTables(conn, db)
		for _, table := range tables {
			err = changePKType(conn, table, db)
			if err == nil {
				err = addPKConstraint(conn, table, db)
				if err == nil {
					fmt.Println("Alter Success : " + db + "." + table)
				} else {
					fmt.Println("Alter Failed " + db + "." + table + " : " + err.Error())
				}
			}
		}
	}

	conn.Close()
}

func changePKType(conn *sql.DB, table string, db string) (err error) {
	query := "ALTER TABLE " + db + "." + table + " MODIFY COLUMN  __os_id varchar(255);"
	err = executeNonQuery(conn, query)
	return
}

func addPKConstraint(conn *sql.DB, table string, db string) (err error) {
	query := "ALTER TABLE " + db + "." + table + " ADD CONSTRAINT pk_" + table + " PRIMARY KEY (__os_id);"
	err = executeNonQuery(conn, query)
	return
}

func getDBs(conn *sql.DB) (dbs []string) {
	query := "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME != 'information_schema' AND SCHEMA_NAME !='mysql' AND SCHEMA_NAME !='performance_schema';"
	databases, err := executeQueryMany(conn, query, nil)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for _, value := range databases {
			dbs = append(dbs, value["SCHEMA_NAME"].(string))
		}
	}
	return
}

func getTables(conn *sql.DB, db string) (tables []string) {
	query := "SELECT DISTINCT TABLE_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='" + db + "';"
	tablenames, err := executeQueryMany(conn, query, nil)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for _, value := range tablenames {
			tables = append(tables, value["TABLE_NAME"].(string))
		}
	}
	return
}

func getConnection() (conn *sql.DB, err error) {
	fmt.Println("Creating Connection to MySQL!")
	var c *sql.DB
	//c, err = sql.Open("mysql", "duoFWuser"+":"+"duoFWpass"+"@tcp("+"173.194.237.13"+":"+"3306"+")/")
	c, err = sql.Open("mysql", "duoDevUser"+":"+"duoDevPass"+"@tcp("+"173.194.238.163"+":"+"3306"+")/")
	conn = c
	if err == nil {
		fmt.Println("Connection Created!")
	}
	return conn, err
}

func executeQueryMany(conn *sql.DB, query string, tableName interface{}) (result []map[string]interface{}, err error) {
	rows, err := conn.Query(query)

	if err == nil {
		result, err = rowsToMap(rows, tableName)
	} else {
		if strings.HasPrefix(err.Error(), "Error 1146") {
			err = nil
			result = make([]map[string]interface{}, 0)
		}
	}

	return
}

func rowsToMap(rows *sql.Rows, tableName interface{}) (tableMap []map[string]interface{}, err error) {

	columns, _ := rows.Columns()
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

	var cacheItem map[string]string

	for rows.Next() {

		for i, _ := range columns {
			valuePtrs[i] = &values[i]
		}

		rows.Scan(valuePtrs...)

		rowMap := make(map[string]interface{})

		for i, col := range columns {
			if col == "__os_id" || col == "__osHeaders" {
				continue
			}
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				if cacheItem != nil {
					t, ok := cacheItem[col]
					if ok {
						v = sqlToGolang(b, t)
					}
				}

				if v == nil {
					if b == nil {
						v = nil
					} else if strings.ToLower(string(b)) == "null" {
						v = nil
					} else {
						v = string(b)
					}

				}
			} else {
				v = val
			}
			rowMap[col] = v
		}
		tableMap = append(tableMap, rowMap)
	}

	return
}

func sqlToGolang(b []byte, t string) interface{} {

	if b == nil {
		return nil
	}

	if len(b) == 0 {
		return b
	}

	var outData interface{}
	tmp := string(b)
	switch t {
	case "bit(1)":
		if len(b) == 0 {
			outData = false
		} else {
			if b[0] == 1 {
				outData = true
			} else {
				outData = false
			}
		}

		break
	case "double":
		fData, err := strconv.ParseFloat(tmp, 64)
		if err != nil {
			fmt.Println(err.Error())
			outData = tmp
		} else {
			outData = fData
		}
		break
	//case "text":
	//case "blob":
	default:
		if len(tmp) == 4 {
			if strings.ToLower(tmp) == "null" {
				outData = nil
				break
			}
		}

		/*
			var m map[string]interface{}
			var ml []map[string]interface{}


			if (string(tmp[0]) == "{"){
				err := json.Unmarshal([]byte(tmp), &m)
				if err == nil {
					outData = m
				}else{
					fmt.Println(err.Error())
					outData = tmp
				}
			}else if (string(tmp[0]) == "["){
				err := json.Unmarshal([]byte(tmp), &ml)
				if err == nil {
					outData = ml
				}else{
					fmt.Println(err.Error())
					outData = tmp
				}
			}else{
				outData = tmp
			}
		*/
		if string(tmp[0]) == "^" {
			byteData := []byte(tmp)
			bdata := string(byteData[1:])
			decData, _ := base64.StdEncoding.DecodeString(bdata)
			outData = getInterfaceValue(string(decData))

		} else {
			outData = getInterfaceValue(tmp)
		}

		break
	}

	return outData
}

func getInterfaceValue(tmp string) (outData interface{}) {
	var m interface{}
	if string(tmp[0]) == "{" || string(tmp[0]) == "[" {
		err := json.Unmarshal([]byte(tmp), &m)
		if err == nil {
			outData = m
		} else {
			fmt.Println(err.Error())
			outData = tmp
		}
	} else {
		outData = tmp
	}
	return
}

func executeNonQuery(conn *sql.DB, query string) (err error) {
	var stmt *sql.Stmt
	stmt, err = conn.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec()

	if err != nil {
		return err
	}
	return err
}
