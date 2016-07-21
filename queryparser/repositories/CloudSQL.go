package repositories

import (
	"duov6.com/queryparser/structs"
	"strings"
)

type CloudSQL struct {
}

func (repository CloudSQL) GetName(request structs.RepoRequest) string {
	return ("CLOUDSQL : " + request.Repository)
}

func (repository CloudSQL) GetQuery(request structs.RepoRequest) structs.RepoResponse {
	response := structs.RepoResponse{}
	skip := "0"
	take := "1000000"

	isOrderByAsc := false
	isOrderByDesc := false
	orderbyfield := ""
	isExistingOrderBys := false

	if request.Parameters["skip"] != nil {
		if request.Parameters["skip"].(string) != "" {
			skip = request.Parameters["skip"].(string)
		}
	}

	if request.Parameters["take"] != nil {
		if request.Parameters["take"].(string) != "" {
			take = request.Parameters["take"].(string)
		}
	}

	if request.Parameters["orderby"] != nil {
		if request.Parameters["orderby"].(string) != "" {
			orderbyfield = request.Parameters["orderby"].(string)
			isOrderByAsc = true
		}
	} else if request.Parameters["orderbydsc"] != nil {
		if request.Parameters["orderbydsc"].(string) != "" {
			orderbyfield = request.Parameters["orderbydsc"].(string)
			isOrderByDesc = true
		}
	}

	orderByQueryPart := ""

	if len(request.Queryobject.Orderby) > 0 {
		isExistingOrderBys = true
	}

	if isOrderByAsc {
		orderByQueryPart += " ORDER BY " + orderbyfield + " asc "
	} else if isOrderByDesc {
		orderByQueryPart += " ORDER BY " + orderbyfield + " desc "
	}

	queryPart := " limit " + take
	queryPart += " offset " + skip + " "

	query := request.Query

	query = strings.Replace(query, ";", "", -1)

	if isExistingOrderBys {
		if isOrderByAsc || isOrderByDesc {
			orderByQueryPart += " , "
			query = strings.Replace(query, "ORDER BY", orderByQueryPart, 1)
		}
	} else {
		query += orderByQueryPart
	}

	query += queryPart + ";"
	response.Query = query
	return response
}
