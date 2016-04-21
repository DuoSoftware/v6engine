package drivers

import (
	"database/sql"
	"duov6.com/common"
	"duov6.com/objectstore/messaging"
	"encoding/base64"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"strings"
	"time"
)

type CloudSql struct {
}

var CloudSqlConnection map[string]*sql.DB

func (driver CloudSql) getConnection(request *messaging.ObjectRequest) (conn *sql.DB, err error) {

	if CloudSqlConnection == nil {
		CloudSqlConnection = make(map[string]*sql.DB)
	}

	if CloudSqlConnection[request.Controls.Namespace] == nil {
		var c *sql.DB
		mysqlConf := request.Configuration.ServerConfiguration["MYSQL"]
		c, err = sql.Open("mysql", mysqlConf["Username"]+":"+mysqlConf["Password"]+"@tcp("+mysqlConf["Url"]+":"+mysqlConf["Port"]+")/")
		c.SetMaxIdleConns(10)
		c.SetMaxOpenConns(0)
		c.SetConnMaxLifetime(time.Duration(120) * time.Second)
		conn = c
		CloudSqlConnection[request.Controls.Namespace] = c
	} else {
		if CloudSqlConnection[request.Controls.Namespace].Ping(); err != nil {
			_ = CloudSqlConnection[request.Controls.Namespace].Close()
			CloudSqlConnection[request.Controls.Namespace] = nil
			var c *sql.DB
			mysqlConf := request.Configuration.ServerConfiguration["MYSQL"]
			c, err = sql.Open("mysql", mysqlConf["Username"]+":"+mysqlConf["Password"]+"@tcp("+mysqlConf["Url"]+":"+mysqlConf["Port"]+")/")
			c.SetMaxIdleConns(10)
			c.SetMaxOpenConns(0)
			c.SetConnMaxLifetime(time.Duration(120) * time.Second)
			conn = c
			CloudSqlConnection[request.Controls.Namespace] = c
		} else {
			conn = CloudSqlConnection[request.Controls.Namespace]
		}
	}
	return conn, err
}
func (driver CloudSql) UpdateCloudSqlRecordID(request *messaging.ObjectRequest, amount int) {
	fmt.Println("huehuehue")
}

func (driver CloudSql) UpdateCloudSqlRecordID1(request *messaging.ObjectRequest, amount int) {

	session, isError := driver.getConnection(request)
	if isError != nil {
		return
	} else {
		driver.verifyDBTableAvailability(session, request)
		db := driver.getDatabaseName(request.Controls.Namespace)
		class := strings.ToLower(request.Controls.Class)

		readQuery := "SELECT maxCount FROM " + db + ".domainClassAttributes where class = '" + class + "';"
		myMap, _ := driver.executeQueryOne(session, readQuery, (db + ".domainClassAttributes"))

		if len(myMap) == 0 {
			insertNewClassQuery := "INSERT INTO " + db + ".domainClassAttributes (class,maxCount,version) values ('" + class + "', '2000', '" + common.GetGUID() + "');"
			err := driver.executeNonQuery(session, insertNewClassQuery)
			if err != nil {
				return
			}
		} else {
			//Inrement one and UPDATE
			maxCount := 0
			maxCount, err := strconv.Atoi(myMap["maxCount"].(string))
			maxCount += amount
			returnID := strconv.Itoa(maxCount)
			updateQuery := "UPDATE " + db + ".domainClassAttributes SET maxCount='" + returnID + "' WHERE class = '" + class + "' ;"
			err = driver.executeNonQuery(session, updateQuery)
			if err != nil {
				return
			}
		}
	}

	return
}

func (driver CloudSql) verifyDBTableAvailability(conn *sql.DB, request *messaging.ObjectRequest) {
	db := driver.getDatabaseName(request.Controls.Namespace)
	class := strings.ToLower(request.Controls.Class)
	driver.VerifyDBAvailability(conn, db)
	driver.VerifyTableAvailability(conn, db, class)
}

func (driver CloudSql) VerifyDBAvailability(conn *sql.DB, database string) {
	query := "CREATE DATABASE IF NOT EXISTS " + database + ";"
	err := driver.executeNonQuery(conn, query)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (driver CloudSql) VerifyTableAvailability(conn *sql.DB, database string, class string) {
	createDomainAttrQuery := "create table " + database + ".domainClassAttributes ( class VARCHAR(255) primary key, maxCount text, version text);"
	err := driver.executeNonQuery(conn, createDomainAttrQuery)
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}

func (driver CloudSql) getDatabaseName(namespace string) string {
	return "_" + strings.ToLower(strings.Replace(namespace, ".", "", -1))
}

func (driver CloudSql) executeNonQuery(conn *sql.DB, query string) (err error) {

	var stmt *sql.Stmt
	stmt, err = conn.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec()

	if err != nil {
		return err
	}
	_ = stmt.Close()
	return
}

func (driver CloudSql) executeQueryOne(conn *sql.DB, query string, tableName interface{}) (result map[string]interface{}, err error) {
	rows, err := conn.Query(query)

	if err == nil {
		var resultSet []map[string]interface{}
		resultSet, err = driver.rowsToMap(rows, tableName)
		if len(resultSet) > 0 {
			result = resultSet[0]
		}

	} else {
		if strings.HasPrefix(err.Error(), "Error 1146") {
			err = nil
			result = make(map[string]interface{})
		}
	}

	return
}

func (driver CloudSql) rowsToMap(rows *sql.Rows, tableName interface{}) (tableMap []map[string]interface{}, err error) {

	columns, _ := rows.Columns()
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

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
				v = driver.sqlToGolang(b, "TEXT")
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

func (driver CloudSql) sqlToGolang(b []byte, t string) interface{} {

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
			outData = tmp
		} else {
			outData = fData
		}
		break
	case "BIT":
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
	case "DOUBLE":
		fData, err := strconv.ParseFloat(tmp, 64)
		if err != nil {
			outData = tmp
		} else {
			outData = fData
		}
		break

	default:
		if len(tmp) == 4 {
			if strings.ToLower(tmp) == "null" {
				outData = nil
				break
			}
		}

		if string(tmp[0]) == "^" {
			byteData := []byte(tmp)
			bdata := string(byteData[1:])
			decData, _ := base64.StdEncoding.DecodeString(bdata)
			outData = driver.getInterfaceValue(string(decData))

		} else {
			outData = driver.getInterfaceValue(tmp)
		}

		break
	}

	return outData
}

func (driver CloudSql) getInterfaceValue(tmp string) (outData interface{}) {
	var m interface{}
	if string(tmp[0]) == "{" || string(tmp[0]) == "[" {
		err := json.Unmarshal([]byte(tmp), &m)
		if err == nil {
			outData = m
		} else {
			outData = tmp
		}
	} else {
		outData = tmp
	}
	return
}

// else {
// 			insertQuery := "INSERT INTO " + database + ".domainClassAttributes (class, maxCount,version) VALUES ('" + class + "','2000','" + common.GetGUID() + "')"
// 			err = driver.executeNonQuery(conn, insertQuery)
// 			if err != nil {
// 				fmt.Println(err.Error())
// 			}
// 		}
