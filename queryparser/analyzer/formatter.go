package analyzer

import (
	"duov6.com/queryparser/common"
	"fmt"
	"google.golang.org/cloud/datastore"
	"strings"
)

func GetOtherQuery(query string, repository string) (retQuery string, retDquery *datastore.Query) {

	switch repository {
	case "CDS":
		retQuery = ""
		retDquery = nil
		break
	default:
		retQuery = query
	}

	return
}

func PrepareSQLStatement(input string, repo string, namespace string, class string) (query string, isValid bool) {
	query = ""
	isValid = true

	//check for complex queries...
	fromIndex := strings.Index(input, " FROM ") + 5
	fromSlice := input[fromIndex:]
	if strings.Contains(fromSlice, "(") && strings.Contains(fromSlice, ")") {

	}

	queryTokens := strings.Split(input, " ")

	for index := 0; index < len(queryTokens); index++ {
		if strings.EqualFold(queryTokens[index], "select") {
			queryTokens[index] = "SELECT"
		} else if strings.EqualFold(queryTokens[index], "from") {
			queryTokens[index] = "FROM"
		} else if strings.EqualFold(queryTokens[index], "where") {
			queryTokens[index] = "WHERE"
		} else if strings.EqualFold(queryTokens[index], "and") {
			queryTokens[index] = "AND"
		} else if strings.EqualFold(queryTokens[index], "or") {
			queryTokens[index] = "OR"
		} else if strings.EqualFold(queryTokens[index], "not") {
			queryTokens[index] = "NOT"
		} else if strings.EqualFold(queryTokens[index], "order") && strings.EqualFold(queryTokens[index+1], "by") {
			queryTokens[index] = "ORDER"
			queryTokens[index+1] = "BY"
		} else if strings.EqualFold(queryTokens[index], "group") && strings.EqualFold(queryTokens[index+1], "by") {
			queryTokens[index] = "GROUP"
			queryTokens[index+1] = "BY"
			isValid = false
		}
	}

	for index := 0; index < len(queryTokens); index++ {
		query += queryTokens[index] + " "
	}
	query = strings.TrimSpace(query)
	query = formatTableNames(repo, namespace, class, query)
	return
}

func formatTableNames(repo string, namespace string, class string, query string) (retQurey string) {

	if strings.Contains(query, " WHERE ") && strings.Contains(query, " ORDER BY ") {
		fromIndex := strings.Index(query, " FROM ") + 5
		whereIndex := strings.Index(query, " WHERE ")
		one := query[:fromIndex]
		two := query[whereIndex:]
		fmt.Println(1)
		fmt.Println(one)
		fmt.Println(two)
		retQurey = one + " " + common.GetSQLTableName(repo, namespace, class) + two

	} else if strings.Contains(query, " WHERE ") && !strings.Contains(query, " ORDER BY ") {
		fromIndex := strings.Index(query, " FROM ") + 5
		whereIndex := strings.Index(query, " WHERE ")
		fmt.Println(2)
		one := query[:fromIndex]
		two := query[whereIndex:]
		fmt.Println(one)
		fmt.Println(two)
		retQurey = one + " " + common.GetSQLTableName(repo, namespace, class) + two

	} else if !strings.Contains(query, " WHERE ") && strings.Contains(query, " ORDER BY ") {
		fromIndex := strings.Index(query, " FROM ") + 5
		orderByIndex := strings.Index(query, " ORDER BY ")
		one := query[:fromIndex]
		two := query[orderByIndex:]
		fmt.Println(3)
		fmt.Println(one)
		fmt.Println(two)
		retQurey = one + " " + common.GetSQLTableName(repo, namespace, class) + two

	} else if !strings.Contains(query, " WHERE ") && !strings.Contains(query, " ORDER BY ") {
		fromIndex := strings.Index(query, " FROM ") + 5
		queryWithoutClass := query[:fromIndex]
		fmt.Println(4)
		fmt.Println(queryWithoutClass)
		retQurey = queryWithoutClass + " " + common.GetSQLTableName(repo, namespace, class)
	}

	return
}
