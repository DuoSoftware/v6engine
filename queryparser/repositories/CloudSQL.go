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

	if request.Parameters["skip"].(string) != "" {
		skip = request.Parameters["skip"].(string)
	}

	if request.Parameters["take"].(string) != "" {
		take = request.Parameters["take"].(string)
	}

	queryPart := " WHERE "
	queryPart += " limit " + take
	queryPart += " offset " + skip + " "

	query := request.Query
	query = strings.Replace(query, ";", "", -1)
	query += queryPart + ";"
	response.Query = query

	return response
}
