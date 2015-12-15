package analyzer

import (
	//"duov6.com/queryparser/analyzer"
	"duov6.com/queryparser/structs"
	"fmt"
	"strings"
)

func GetQueryMaps(query string) (qObject structs.QueryObject) {
	//Types:
	//BasicGet - (select, from)
	//GetOrdered - (select, from, orderby)
	//GetFiltered - (select, from, where)
	//GetFilteredOrdered - (select, from, where, orderby)

	queryType := ""
	if strings.Contains(query, " WHERE ") && strings.Contains(query, " ORDER BY ") {
		queryType = "GetFilteredOrdered"
	} else if strings.Contains(query, " WHERE ") && !strings.Contains(query, " ORDER BY ") {
		queryType = "GetFiltered"
	} else if !strings.Contains(query, " WHERE ") && strings.Contains(query, " ORDER BY ") {
		queryType = "GetOrdered"
	} else if !strings.Contains(query, " WHERE ") && !strings.Contains(query, " ORDER BY ") {
		queryType = "BasicGet"
		qObject = getBasicMapping(query)
	}
	fmt.Println("Type of Query : " + queryType)
	return
}

func getBasicMapping(query string) (qObject structs.QueryObject) {
	selectIndex := (strings.Index(query, "SELECT")) + 6
	fromIndex := strings.Index(query, "FROM")
	unformattedFields := query[selectIndex:fromIndex]
	unformattedFields = strings.TrimSpace(unformattedFields)
	individualFields := strings.Split(unformattedFields, ",")
	qObject.Operation = "SELECT"
	qObject.SelectedFields = individualFields
	qObject.Table = getTableNameFromQuery(query, "BasicGet")
	return
}

func appendOrderBy(query string, object structs.QueryObject) (qObject structs.QueryObject) {
	qObject = object

	OrderBySet := query[((strings.Index(query, "ORDER BY")) + 8):]
	OrderBySet = strings.Replace(OrderBySet, ";", "", -1)

	orderbys := strings.Split(OrderBySet, ",")

	var orderbymap map[string]string
	orderbymap = make(map[string]string)

	for index, _ := range orderbys {
		orderbys[index] = strings.TrimSpace(orderbys[index])
		if elements := strings.Split(orderbys[index], " "); len(elements) == 1 {
			orderbymap["ASC"] = elements[0]
		} else if len(elements) == 2 {
			orderbymap[elements[1]] = elements[0]
		}
	}

	qObject.Orderby = orderbymap
	return
}

func appendWhere(query string, object structs.QueryObject) (qObject structs.QueryObject) {
	qObject = object

	whereIndex := strings.Index(query, "WHERE") + 5
	whereSet := ""
	if strings.Contains(query[whereIndex:], "ORDER BY") {
		orderbyIndex := strings.Index(query, "ORDER BY")
		whereSet = query[whereIndex:orderbyIndex]
	} else {
		whereSet = query[whereIndex:]
		whereSet = strings.Replace(whereSet, ";", "", -1)
	}

	whereSet = strings.TrimSpace(whereSet)

	return
}

func getTableNameFromQuery(query string, queryType string) (tableName string) {
	switch queryType {
	case "BasicGet":
		fromIndex := (strings.Index(query, "FROM") + 4)
		tableName = extractTableNameFromString(query[fromIndex:])
		break
	default:
		tableName = "undefined"
		break
	}
	return
}

func extractTableNameFromString(input string) (tablename string) {
	input = strings.TrimSpace(input)
	tokens := strings.Split(input, ",")
	//only take the first element. Because already validated in earlier steps
	items := strings.Split(tokens[0], " ")
	tablename = items[0]
	return
}

func prepareWhereClause(input string) (output string) {
	output = input
	strings.Replace(input, " = ", "=", -1)
	strings.Replace(input, " <> ", "!=", -1)
	strings.Replace(input, " != ", "!=", -1)
	strings.Replace(input, " > ", ">", -1)
	strings.Replace(input, " < ", "<", -1)
	strings.Replace(input, " >= ", ">=", -1)
	strings.Replace(input, " <= ", "<=", -1)
	strings.Replace(input, "NOT BETWEEN", "NOTBETWEEN", -1)
	strings.Replace(input, "IN (", "IN(", -1)

	return
}
