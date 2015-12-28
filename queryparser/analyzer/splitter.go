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
		qObject = getBasicMapping(query)
		qObject = appendWhere(query, qObject)
	} else if !strings.Contains(query, " WHERE ") && strings.Contains(query, " ORDER BY ") {
		queryType = "GetOrdered"
		qObject = getBasicMapping(query)
		qObject = appendOrderBy(query, qObject)
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
	whereSet = prepareWhereClause(whereSet)
	stringSet := getWhereSets(whereSet)
	fmt.Println("--------")
	for x := 0; x < len(stringSet); x++ {
		fmt.Println(stringSet[x])
	}
	fmt.Println("--------")
	// fmt.Println("--------")
	// for x := 0; x < len(stringSet); x++ {
	// 	fmt.Println(createArrayFromWhereString(stringSet[x]))
	// }
	// fmt.Println("--------")
	return
}

func getTableNameFromQuery(query string, queryType string) (tableName string) {
	switch queryType {
	case "BasicGet":
		fromIndex := (strings.Index(query, "FROM") + 4)
		tableName = extractTableNameFromString(query[fromIndex:])
		break
	case "GetOrdered":
		fromIndex := strings.Index(query, "FROM") + 4
		orderByIndex := strings.Index(query, "ORDER BY")
		tableName = extractTableNameFromString(query[fromIndex:orderByIndex])
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
	output = strings.Replace(output, "  ", " ", -1)
	output = strings.Replace(output, " = ", "=", -1)
	output = strings.Replace(output, " =", "=", -1)
	output = strings.Replace(output, "= ", "=", -1)
	output = strings.Replace(output, " <> ", "!=", -1)
	output = strings.Replace(output, "<> ", "!=", -1)
	output = strings.Replace(output, " <>", "!=", -1)
	output = strings.Replace(output, " != ", "!=", -1)
	output = strings.Replace(output, "!= ", "!=", -1)
	output = strings.Replace(output, " !=", "!=", -1)
	output = strings.Replace(output, " > ", ">", -1)
	output = strings.Replace(output, " >", ">", -1)
	output = strings.Replace(output, "> ", ">", -1)
	output = strings.Replace(output, " < ", "<", -1)
	output = strings.Replace(output, "< ", "<", -1)
	output = strings.Replace(output, " <", "<", -1)
	output = strings.Replace(output, " >= ", ">=", -1)
	output = strings.Replace(output, " >=", ">=", -1)
	output = strings.Replace(output, ">= ", ">=", -1)
	output = strings.Replace(output, " <= ", "<=", -1)
	output = strings.Replace(output, " <=", "<=", -1)
	output = strings.Replace(output, "<= ", "<=", -1)
	output = strings.Replace(output, "NOT BETWEEN", "NOTBETWEEN", -1)
	output = strings.Replace(output, "IN (", "IN(", -1)
	output = strings.Replace(output, "NOT IN (", "NOTIN(", -1)
	output = strings.Replace(output, "NOT IN(", "NOTIN(", -1)
	return
}

func whereSeparator(whereClause string) (retArr []map[int][]string) {
	whereClause = prepareWhereClause(whereClause)
	whereTokens := strings.Split(whereClause, " ")

	preFormattedQuery := ""
	for x := 0; x < len(whereTokens); x++ {
		token := whereTokens[x]
		if token == "NOTBETWEEN" {
			preFormattedQuery += "NOT BETWEEN" + " " + whereTokens[x+1] + " " + whereTokens[x+2] + " " + whereTokens[x+3]
			x += 3
		} else {
			preFormattedQuery += (token + " ")
		}
	}

	return
}

func getWhereSets(whereClause string) (set map[int]string) {
	index := 0
	set = make(map[int]string)
	tokens := strings.Split(whereClause, " ")

	for x := 0; x < len(tokens); x++ {
		if strings.EqualFold(tokens[x], "between") && strings.EqualFold(tokens[x+2], "and") {
			set[index-1] = (tokens[x-1] + " " + tokens[x] + " " + tokens[x+1] + " " + tokens[x+2] + " " + tokens[x+3])
			x += 4
		} else if strings.EqualFold(tokens[x], "notbetween") && strings.EqualFold(tokens[x+2], "and") {
			set[index-1] = (tokens[x-1] + " " + tokens[x] + " " + tokens[x+1] + " " + tokens[x+2] + " " + tokens[x+3])
			x += 4
		} else if strings.Contains(tokens[x], "IN(") || strings.Contains(tokens[x], "NOTIN(") {
			endIndex := -1
			for y := x; y < len(tokens); y++ {
				if strings.Contains(tokens[y], ")") {
					endIndex = y + 1
					break
				}
			}
			indexValue := tokens[x-1] + " "

			for z := x; z < endIndex; z++ {
				indexValue += tokens[z] + " "
				fmt.Println(tokens[z])
			}

			indexValue = strings.TrimSpace(indexValue)

			set[index-1] = indexValue
			x += len(tokens[x:endIndex]) + 1
		} else {
			set[index] = tokens[x]
			index += 1
		}
	}
	return
}

func createArrayFromWhereString(input string) (output []string) {
	tokens := strings.Split(input, " ")

	if len(tokens) > 3 {
		if tokens[1] == "BETWEEN" && tokens[3] == "AND" {
			output = tokens
		} else if tokens[1] == "NOTBETWEEN" && tokens[3] == "AND" {
			output = tokens
		}
	} else {
		if strings.Contains(input, "!=") {
			words := strings.Split(input, "!=")
			output = makeFinalwhereArray("!=", words)
		} else if strings.Contains(input, ">=") {
			words := strings.Split(input, ">=")
			output = makeFinalwhereArray(">=", words)
		} else if strings.Contains(input, ">") {
			words := strings.Split(input, ">")
			output = makeFinalwhereArray(">", words)
		} else if strings.Contains(input, "<=") {
			words := strings.Split(input, "<=")
			output = makeFinalwhereArray("<=", words)
		} else if strings.Contains(input, "<") {
			words := strings.Split(input, "<")
			output = makeFinalwhereArray("<", words)
		} else if strings.Contains(input, "=") {
			words := strings.Split(input, "=")
			output = makeFinalwhereArray("=", words)
		}
	}
	return
}

func makeFinalwhereArray(sign string, words []string) []string {
	index := 0
	newArr := make([]string, (len(words) + len(words) + 1))
	for x := 0; x < len(newArr); x++ {
		if x%2 == 0 {
			newArr[x] = words[index]
			index += 1
		} else {
			newArr[x] = sign
		}
	}
	return newArr
}
