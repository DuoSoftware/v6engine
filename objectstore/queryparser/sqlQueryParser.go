package queryparser

import (
	"fmt"
	"strings"
)

func GetFormattedQuery(oldQuery string) (newQuery string) {
	fmt.Println("Starting formatting query process...")

	if !checkIfException(oldQuery) {

		newQuery = oldQuery
		tables := getTables(oldQuery)

		fmt.Print("Tables in Query Before Converting to Hive : ")
		fmt.Println(tables)

		newTables := convertToHiveTableFormat(tables)

		fmt.Print("Hive Converted Table Names : ")
		fmt.Println(newTables)

		for key, _ := range tables {
			newQuery = strings.Replace(newQuery, tables[key], newTables[key], -1)
		}
	} else {
		newQuery = oldQuery
	}

	return
}

func getTables(query string) map[int]string {
	var tableNames map[int]string
	tableNames = make(map[int]string)

	//get lowercase query
	loweredQuery := RebuildQuery(query)

	//split at "from"
	fromSplit := strings.Split(loweredQuery, "from")

	//Get 2nd index and proceed
	//split string at "whitespaces"

	tokens := strings.Split(fromSplit[1], " ")

	//create a string to save next break point
	var breakpoint string
	breakpoint = ""

	for _, value := range tokens {

		if value == "where" {
			breakpoint = "where"
			break
		} else if value == "group" {
			breakpoint = "group"
			break
		} else if value == "having" {
			breakpoint = "having"
			break
		} else {
			breakpoint = query[(len(query) - 1):(len(query))]
			break
		}
	}

	//fmt.Println("Break Point : " + breakpoint)

	//get values between "from" and <breakpoint>

	tempQueryString := ""

	for _, value := range tokens {

		if value != breakpoint {
			tempQueryString += value + " "
		} else {
			break
		}
	}

	//	fmt.Println("Needed value set (BEFORE) : |" + tempQueryString + "|")

	//Identify how many tables used...

	tables := strings.Split(tempQueryString, ",")

	index := 0
	for _, value := range tables {
		//fmt.Println("Original Table : |" + value + "|")
		value = strings.TrimLeft(value, " ")
		value = strings.TrimRight(value, " ")
		tempTableSubValues := strings.Split(value, " ")
		tableNames[index] = strings.TrimSpace(tempTableSubValues[0])
		index++
	}

	//	fmt.Print("Finalized Tabled : ")
	//	fmt.Println(tableNames)

	return tableNames
}

func RebuildQuery(oldQuery string) (newQuery string) {

	//split the query in spaces

	spaceSplit := strings.Split(oldQuery, " ")

	//Lowercase the needed keywords
	for key, value := range spaceSplit {
		if value == "select" || value == "SELECT" {
			spaceSplit[key] = "select"
		}
		if value == "from" || value == "FROM" {
			spaceSplit[key] = "from"
		}
		if value == "where" || value == "WHERE" {
			spaceSplit[key] = "where"
		}
		if value == "group" || value == "GROUP" {
			spaceSplit[key] = "group by"
		}
		if value == "having" || value == "HAVING" {
			spaceSplit[key] = "having"
		}
	}

	//Rebuild the query

	newQuery = ""

	for key, value := range spaceSplit {
		if key != len(spaceSplit)-1 {
			newQuery += value + " "
		} else {
			newQuery += value
		}
	}

	return
}

func convertToHiveTableFormat(input map[int]string) map[int]string {

	var outMap map[int]string
	outMap = make(map[int]string)

	for key, value := range input {
		outMap[key] = value
	}

	for key, _ := range outMap {
		//replace first two dots to 8s
		outMap[key] = strings.Replace(outMap[key], ".", "", 2)
		//replace last dot to 0s
		//outMap[key] = strings.Replace(outMap[key], ".", "0", 1)
	}

	return outMap
}

func checkIfException(query string) (isException bool) {

	isException = false

	tempTokens := strings.Split(query, " ")

	var myMap map[int]string
	myMap = make(map[int]string)

	myMap[0] = "show"
	myMap[1] = "describe"

	for _, value := range myMap {

		if strings.TrimSpace(tempTokens[0]) == value {
			isException = true
			break
		}
	}

	return

}

func GetTablesInQuery(query string) (tempMap map[int]string, isException bool) {

	//check if no need to check tables

	isException = checkIfException(query)

	if !isException {
		tempMap = make(map[int]string)
		tempMap = getTables(query)
	} else {
		tempMap = nil
	}

	return tempMap, isException
}
