package queryparser

import (
	"fmt"
	"strings"
)

func ConvertToTSQLTags(queryString string) (whereTags map[string]string, selectTags map[string]string, class string) {

	arr := strings.Split(queryString, "from")

	//for select clause

	selectTags = make(map[string]string)
	selectTags = GetTags(arr[0])

	//From Class
	tempClass := strings.Split(arr[1], " ")
	class = tempClass[1]
	fmt.Println("Class : " + class)

	//where clauses

	whereArr := strings.Split(queryString, "where")

	//whereTags = make(map[string]string)
	//whereTags = GetTags(whereArr[1])
	
	if len(whereArr) > 1 {
		whereTags = make(map[string]string)
		whereTags = GetTags(whereArr[1])
	} else {
		//No Where Tags...
		fmt.Println("No Where Tags!")
		whereTags = nil
	}

	return
}
