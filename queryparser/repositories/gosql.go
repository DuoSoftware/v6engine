package repositories

import (
	"duov6.com/queryparser/common"
	"fmt"
	"strings"
)

func GetSQLQuery(repo string, input string, namespace string, class string) (query string) {
	if strings.Contains(query, " WHERE ") && strings.Contains(query, " ORDER BY ") {
		fromIndex := strings.Index(input, " FROM ") + 5
		whereIndex := strings.Index(input, " WHERE ")
		one := query[:fromIndex]
		two := query[whereIndex:]
		fmt.Println(1)
		fmt.Println(one)
		fmt.Println(two)
		query = one + " " + common.GetSQLTableName(repo, namespace, class) + two
	} else if strings.Contains(query, " WHERE ") && !strings.Contains(query, " ORDER BY ") {
		fromIndex := strings.Index(input, " FROM ") + 5
		whereIndex := strings.Index(input, " WHERE ")
		fmt.Println(2)
		one := query[:fromIndex]
		two := query[whereIndex:]
		fmt.Println(one)
		fmt.Println(two)
		query = one + " " + common.GetSQLTableName(repo, namespace, class) + two
	} else if !strings.Contains(query, " WHERE ") && strings.Contains(query, " ORDER BY ") {
		fromIndex := strings.Index(input, " FROM ") + 5
		orderByIndex := strings.Index(input, " ORDER BY ")
		one := query[:fromIndex]
		two := query[orderByIndex:]
		fmt.Println(3)
		fmt.Println(one)
		fmt.Println(two)
		query = one + " " + common.GetSQLTableName(repo, namespace, class) + two
	} else if !strings.Contains(query, " WHERE ") && !strings.Contains(query, " ORDER BY ") {
		fromIndex := strings.Index(input, " FROM ") + 5
		queryWithoutClass := input[:fromIndex]
		fmt.Println(4)
		fmt.Println(queryWithoutClass)
		query = queryWithoutClass + " " + common.GetSQLTableName(repo, namespace, class)
	}
	return
}
