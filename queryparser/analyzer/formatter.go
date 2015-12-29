package analyzer

import (
	"duov6.com/queryparser/common"
	"errors"
	"google.golang.org/cloud/datastore"
	"strings"
)

func GetOtherQuery(query string, repository string) (retQuery interface{}) {

	switch repository {
	case "CDS":
		retQuery = datastore.NewQuery(query)
		break
	default:
		retQuery = query
	}

	return
}

func PrepareSQLStatement(input string, repo string, namespace string, class string) (query string, isValid error) {
	query = ""
	isValid = nil

	//check for complex queries...
	fromIndex := strings.Index(input, " FROM ") + 5
	fromSlice := input[fromIndex:]
	if strings.Contains(fromSlice, "(") && strings.Contains(fromSlice, ")") {
		switch repo {
		case "CDS":
			query = ""
			isValid = errors.New("Complex queries are not allowed in Cloud SQL Repository!")
			return
			break
		case "ES":
			query = ""
			isValid = errors.New("Complex queries are not allowed in Elastic Repository!")
			return
			break
		default:
			query = ""
			isValid = nil
		}
	}

	trailerRemovedInput := strings.Replace(input, ";", "", -1)
	queryTokens := strings.Split(trailerRemovedInput, " ")

	for index := 0; index < len(queryTokens); index++ {
		if strings.EqualFold(queryTokens[index], "select") {
			queryTokens[index] = "SELECT"
		} else if strings.EqualFold(queryTokens[index], "between") {
			queryTokens[index] = "BETWEEN"
		} else if strings.EqualFold(queryTokens[index], "in") && strings.Contains(queryTokens[index+1], "(") {
			queryTokens[index] = "IN"
		} else if strings.EqualFold(queryTokens[index], "not") && strings.EqualFold(queryTokens[index+1], "between") {
			queryTokens[index] = "NOT"
		} else if strings.EqualFold(queryTokens[index], "not") && strings.EqualFold(queryTokens[index+1], "in") {
			queryTokens[index] = "NOT"
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
			isValid = errors.New("GROUP BY queries are not allowed!")
		} else if strings.EqualFold(queryTokens[index], "asc") {
			queryTokens[index] = "ASC"
		} else if strings.EqualFold(queryTokens[index], "desc") {
			queryTokens[index] = "DESC"
		} else if strings.EqualFold(queryTokens[index], "asc,") {
			queryTokens[index] = "ASC,"
		} else if strings.EqualFold(queryTokens[index], "desc,") {
			queryTokens[index] = "DESC,"
		} else if strings.EqualFold(queryTokens[index], ",asc") {
			queryTokens[index] = ",ASC"
		} else if strings.EqualFold(queryTokens[index], ",desc") {
			queryTokens[index] = ",DESC"
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
		retQurey = one + " " + common.GetSQLTableName(repo, namespace, class) + two

	} else if strings.Contains(query, " WHERE ") && !strings.Contains(query, " ORDER BY ") {
		fromIndex := strings.Index(query, " FROM ") + 5
		whereIndex := strings.Index(query, " WHERE ")
		one := query[:fromIndex]
		two := query[whereIndex:]
		retQurey = one + " " + common.GetSQLTableName(repo, namespace, class) + two

	} else if !strings.Contains(query, " WHERE ") && strings.Contains(query, " ORDER BY ") {
		fromIndex := strings.Index(query, " FROM ") + 5
		orderByIndex := strings.Index(query, " ORDER BY ")
		one := query[:fromIndex]
		two := query[orderByIndex:]
		retQurey = one + " " + common.GetSQLTableName(repo, namespace, class) + two

	} else if !strings.Contains(query, " WHERE ") && !strings.Contains(query, " ORDER BY ") {
		fromIndex := strings.Index(query, " FROM ") + 5
		queryWithoutClass := query[:fromIndex]
		retQurey = queryWithoutClass + " " + common.GetSQLTableName(repo, namespace, class)
	}

	return
}
