package repositories

import (
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"strings"
)

func getNoSqlKeyById(request *messaging.ObjectRequest, obj map[string]interface{}) string {
	key := request.Controls.Namespace + "." + request.Controls.Class + "." + obj[request.Body.Parameters.KeyProperty].(string)
	return key
}

func getNoSqlKey(request *messaging.ObjectRequest) string {
	key := request.Controls.Namespace + "." + request.Controls.Class + "." + request.Controls.Id
	return key
}

func getStringByObject(obj interface{}) string {
	result, err := json.Marshal(obj)
	if err == nil {
		return string(result)
	} else {
		return "{}"
	}
}

func checkEmptyByteArray(input []byte) (status bool) {
	status = false
	if len(input) == 4 || len(input) == 2 || len(input) < 2 || input == nil {
		status = true
	}
	return
}

func getSearchResultKey(request *messaging.ObjectRequest) string {

	skip := "skip=0"
	take := "take=100"
	orderBy := "orderyBy="
	orderByDsc := "orderByDsc="
	keyword := "keyword=" + request.Body.Query.Parameters

	if request.Extras["skip"] != nil {
		skip = "skip=" + request.Extras["skip"].(string)
	}

	if request.Extras["take"] != nil {
		take = "take=" + request.Extras["take"].(string)
	}

	if request.Extras["orderby"] != nil {
		orderBy = "orderyBy=" + request.Extras["orderby"].(string)
	} else if request.Extras["orderbydsc"] != nil {
		orderByDsc = "orderByDsc=" + request.Extras["orderbydsc"].(string)
	}

	namespace := request.Controls.Namespace
	class := request.Controls.Class

	url := namespace + ":" + class + ":" + keyword + ":" + skip + ":" + take + ":" + orderBy + ":" + orderByDsc

	return url
}

func getQueryResultKey(request *messaging.ObjectRequest) string {
	query := request.Body.Query.Parameters
	namespace := request.Controls.Namespace
	class := request.Controls.Class

	skip := "0"
	take := "1000000"

	if request.Extras["skip"] != nil {
		if request.Extras["skip"].(string) != "" {
			skip = request.Extras["skip"].(string)
		}
	}

	if request.Extras["take"] != nil {
		if request.Extras["take"].(string) != "" {
			take = request.Extras["take"].(string)
		}
	}

	queryPart := " limit " + take
	queryPart += " offset " + skip + " "

	query = strings.Replace(query, ";", "", -1)
	query += queryPart + ";"

	query = strings.Replace(query, " ", "", -1)

	url := namespace + ":" + class + ":Query:" + query
	return url
}
