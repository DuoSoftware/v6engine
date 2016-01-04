package repositories

import (
	"duov6.com/queryparser/structs"
	"fmt"
	"strings"
)

type ElasticSearch struct {
}

func (repository ElasticSearch) GetName(request structs.RepoRequest) string {
	return ("Elastic Search Query")
}

func (repository ElasticSearch) GetQuery(request structs.RepoRequest) structs.RepoResponse {
	response := structs.RepoResponse{}
	response.Query = request.Query

	queryString := ""

	fmt.Println("-----------------------------------")
	queryString += "{" + repository.GetFieldsJson(request) + ", "

	// fmt.Println(repository.GetFieldsJson(request))

	if len(request.Queryobject.Orderby) != 0 {
		// 	fmt.Println(repository.GetOrderByJson(request))
		queryString += repository.GetOrderByJson(request) + ", "
	}

	if len(request.Queryobject.Where) != 0 {
		// 	fmt.Println(repository.GetWhereJson(request))
		queryString += "\"query\":{\"query_string\" : {\"query\" : \"" + repository.GetWhereJson(request) + "\"}}"
	} else {
		queryString += "\"query\":{\"query_string\" : {\"query\" : \"" + "*" + "\"}}"
	}

	queryString += "}"

	response.Query = queryString

	fmt.Println("-----------------------------------")
	return response
}

func (repository ElasticSearch) GetFieldsJson(request structs.RepoRequest) (json string) {
	fields := request.Queryobject.SelectedFields
	json = "\"_source\": [\""

	for x := 0; x < (len(fields)); x++ {
		if x != len(fields)-1 {
			json += fields[x] + "\",\""
		} else {
			json += fields[x] + "\""
		}
	}

	json += "]"
	return
}

func (repository ElasticSearch) GetOrderByJson(request structs.RepoRequest) (json string) {
	orderBy := request.Queryobject.Orderby
	json = "\"sort\" : ["

	index := 0
	for order, field := range orderBy {
		if index != len(orderBy)-1 {
			json += "{\"" + field + "\" : { \"order\" : \"" + order + "\"}},"
		} else {
			json += "{\"" + field + "\" : { \"order\" : \"" + order + "\"}}"
		}
		index += 1
	}

	json += "]"
	return
}

func (repository ElasticSearch) GetWhereJson(request structs.RepoRequest) (json string) {

	json = ""

	for x := 0; x < len(request.Queryobject.Where); x++ {
		if status := repository.checkIfCombiner(request.Queryobject.Where[x]); status {
			json += " " + request.Queryobject.Where[x][0] + " "
		} else {
			//normalize operators
			operatorPattern := "NOTBETWEEN=LIKE!NOTIN<BETWEEN>IN=!="
			if strings.Contains(operatorPattern, request.Queryobject.Where[x][1]) {
				operator := repository.getElasticOperator(request.Queryobject.Where[x][1])
				request.Queryobject.Where[x][1] = operator
				//fmt.Println(operator)
			}
			//check if complex operator... if not process as simple if condition
			complexPattern := "BETWEEN-NOTBETWEEN-IN-NOTIN-LIKE"
			if strings.Contains(complexPattern, request.Queryobject.Where[x][1]) {
				//fmt.Println("cc")
				json += repository.processComplexOperatorString(request.Queryobject.Where[x])
			} else {
				//fmt.Println("ss")
				json += repository.processNonComplexOperatorString(request.Queryobject.Where[x])
			}
		}
	}

	return
}

func (repository ElasticSearch) processComplexOperatorString(arr []string) (output string) {
	operator := arr[1]

	switch operator {
	case "LIKE":
		keyWord := strings.Replace(arr[2], "'", "", -1)
		output = "(" + arr[0] + ":" + strings.Replace(keyWord, "%", "*", -1) + ")"
		break
	case "BETWEEN":
		output += "("
		output += arr[0] + ":>" + arr[2] + " AND " + arr[0] + ":<" + arr[4] + ")"
		break
	case "NOTBETWEEN":
		output += "("
		output += arr[0] + ":<" + arr[2] + " AND " + arr[0] + ":>" + arr[4] + ")"
		break
	case "IN":
		output += "("
		for x := 2; x < len(arr); x++ {
			if x != len(arr)-1 {
				output += arr[0] + ":" + strings.Replace(arr[x], "'", "", -1) + " OR "
			} else {
				output += arr[0] + ":" + strings.Replace(arr[x], "'", "", -1)
			}
		}
		output += ")"
		break
	case "NOTIN":
		output += "("
		for x := 2; x < len(arr); x++ {
			if x != len(arr)-1 {
				output += "NOT " + arr[0] + ":" + strings.Replace(arr[x], "'", "", -1) + " OR "
			} else {
				output += "NOT " + arr[0] + ":" + strings.Replace(arr[x], "'", "", -1)
			}
		}
		output += ")"
		break
	default:
		output = ""
		break
	}
	return output
}

func (repository ElasticSearch) processNonComplexOperatorString(arr []string) (output string) {
	fmt.Println("*****")
	for x := 0; x < len(arr); x++ {
		fmt.Println(arr[x])
	}
	fmt.Println("*****")
	output += "("
	if arr[1] == "NOT:" {
		output += "NOT "
		arr[1] = ":"
		for x := 0; x < len(arr); x++ {
			val := ""
			val = strings.Replace(arr[x], "(", "", -1)
			// val = strings.Replace(val, ")", "", -1)
			val = strings.Replace(arr[x], "'", "", -1)
			output = val
			//output += arr[x]
		}
	} else {
		for x := 0; x < len(arr); x++ {
			val := ""
			//val = strings.Replace(arr[x], "(", "", -1)
			// val = strings.Replace(val, ")", "", -1)
			val = strings.Replace(arr[x], "'", "", -1)
			output = val
			//output += arr[x]
		}
	}
	output += ")"
	return output
}

func (repository ElasticSearch) getProcessedArray(arr []string) (output string) {
	for x := 0; x < len(arr); x++ {
		arr[x] = strings.Replace(arr[x], "'", "", -1)
	}
	return output
}

func (repository ElasticSearch) checkIfCombiner(arr []string) (status bool) {
	//checks if this is logic combine..  AND / OR
	status = false
	if len(arr) == 1 {
		if strings.EqualFold(arr[0], "AND") || strings.EqualFold(arr[0], "OR") {
			status = true
		}
	}
	return status
}

func (repository ElasticSearch) getElasticOperator(input string) (operator string) {

	switch input {
	case "=":
		operator = ":"
		break
	case "!=":
		operator = "NOT:"
		break
	case ">":
		operator = ":>"
		break
	case ">=":
		operator = ":>="
		break
	case "<":
		operator = ":<"
		break
	case "<=":
		operator = ":<="
		break
	case "LIKE":
		operator = "LIKE"
		break
	case "BETWEEN":
		operator = "BETWEEN"
		break
	case "NOTBETWEEN":
		operator = "NOTBETWEEN"
		break
	case "IN":
		operator = "IN"
		break
	case "NOTIN":
		operator = "NOTIN"
		break
	default:
		operator = ":"
		break
	}
	return operator
}
