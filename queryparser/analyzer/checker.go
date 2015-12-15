package analyzer

import (
	"duov6.com/queryparser/structs"
	"errors"
	"strings"
)

func GetQueryType(query string) (queryType string) {
	inputQuery := strings.TrimSpace(query)
	tokenArray := strings.Split(inputQuery, " ")

	if strings.EqualFold(strings.ToLower(tokenArray[0]), "select") {
		queryType = "SQL"
	} else {
		queryType = "OTHER"
	}

	return
}

func CaseInsensitiveContains(s, substr string) bool {
	s, substr = strings.ToUpper(s), strings.ToUpper(substr)
	return strings.Contains(s, substr)
}

func ValidateQuery(queryObject structs.QueryObject) (err error) {
	//check for selected fields
	for _, fieldName := range queryObject.SelectedFields {
		if ValidateSqlToken(fieldName) {
			err = errors.New("SQL keyword used as Fieldname : " + fieldName)
			return
		}
	}
	//check for where clauses if available
	for _, clause := range queryObject.Where {
		if len(clause[0]) > 1 {
			for _, value := range clause[0] {
				if ValidateSqlToken(value) {
					err = errors.New("SQL keyword used inside WHERE condition : " + value)
					return
				}
			}
		}
	}
	//check for order by clauses
	for key, _ := range queryObject.Orderby {
		if ValidateSqlToken(key) {
			err = errors.New("SQL keyword used inside ORDER BY condition : " + key)
			return
		}
	}
	return
}
