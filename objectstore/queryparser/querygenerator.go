package queryparser

import (
	"fmt"
	"strconv"
	"strings"
)

func GetQuery(queryString string) (returnQuery string, isSelectedFields bool, selectedFields []string, fromClass string) {
	isSelectedFields = false

	//Identify Query Pattern
	tempArr := strings.Split(queryString, " ")
	isTSQL := false
	if strings.ToLower(tempArr[0]) == "select" {
		isTSQL = true
		isSelectedFields = true
	} else {
		isTSQL = false
	}

	if isTSQL {
		//IF DEFAULT T-SQL STYLE
		fmt.Println("T-SQL is Executing!")
		whereTags, selectTags, class := ConvertToTSQLTags(queryString)
		isSelectedFields = true
		selectedFields = MapToArrayConverter(selectTags)
		//returnQuery = GetElasticQuery(whereTags)
		if whereTags != nil {
			returnQuery = GetElasticQuery(whereTags)
		} else {
			returnQuery = "*"
		}
		fromClass = class

	} else {
		//IF SUPUN's SQL STYLE
		fmt.Println("Custom SQL is Executing!")

		//Get Query Tags
		whereTags, selectTags := ConvertToTags(queryString)

		if selectTags != nil {
			isSelectedFields = true
			selectedFields = MapToArrayConverter(selectTags)
		}
		fromClass = ""
		returnQuery = GetElasticQuery(whereTags)

	}

	return

}

func GetElasticQuery(tagMap map[string]string) (queryString string) {

	//convert map to array
	tagArray := MapToArrayConverter(tagMap)

	var tempMap map[string]string
	tempMap = make(map[string]string)

	index := 0
	fmt.Print("Tag Array : ")
	fmt.Println(tagArray)

	for key, tag := range tagArray {
		fmt.Print(tag + " : ")
		if tag == "=" && (tagArray[key-1] != "!" && tagArray[key-1] != ">" && tagArray[key-1] != "<") {
			fmt.Println("Equal Operator")
			tempMap[strconv.Itoa(index)] = ":"
			index++
			continue
		} else if tag == "!" && tagArray[key+1] == "=" {
			fmt.Println("NOT Equal Operator")
			tempMap[strconv.Itoa(index-1)] = "NOT " + tagArray[key-1]
			tempMap[strconv.Itoa(index)] = " : "
			index++
			continue
		} else if tag == ">" && tagArray[key+1] != "=" {
			fmt.Println("Greater Than Operator")
			tempMap[strconv.Itoa(index)] = ":>"
			index++
			continue
		} else if tag == "<" && tagArray[key+1] != "=" {
			fmt.Println("Lesser Than Operator")
			tempMap[strconv.Itoa(index)] = ":<"
			index++
			continue
		} else if tag == ">" && tagArray[key+1] == "=" {
			fmt.Println("Greater Than OR Equal Operator")
			tempMap[strconv.Itoa(index)] = ":>="
			index++
			continue
		} else if tag == "<" && tagArray[key+1] == "=" {
			fmt.Println("Lesser Than OR Equal Operator")
			tempMap[strconv.Itoa(index)] = ":<="
			index++
			continue
		} else {
			fmt.Println("SomeThingElse")
			if tag != "=" {
				var tempArray []string

				if strings.ContainsAny(tag, " ") {
					tempArray = strings.Split(tag, " ")
				} else if strings.ContainsAny(tag, "@") {
					tempArray = strings.Split(tag, "@")
				} else if strings.ContainsAny(tag, ".") {
					tempArray = strings.Split(tag, ".")
				} else if strings.ContainsAny(tag, "-") {
					tempArray = strings.Split(tag, "-")
				} else {
					tempArray = strings.Split(tag, " ")
				}

				//because elastic by design cant string match with 2 operators
				if len(tempArray) != 1 && tagArray[key-2] == "!" {

					tempString := ""
					tempString += tempArray[0]

					for key, _ := range tempArray {
						if key != 0 {
							tempString += " AND NOT " + tagArray[key-1] + " : " + tempArray[key]
						}
					}
					tempMap[strconv.Itoa(index)] = tempString
				} else if len(tempArray) != 1 && tagArray[key-2] != "!" {
					tempString := ""
					tempString += tempArray[0]

					for key, _ := range tempArray {
						if key != 0 {
							tempString += " AND " + tagArray[key-1] + " : " + tempArray[key]
							fmt.Println(tempString)
						}
					}
					tempMap[strconv.Itoa(index)] = tempString
				} else {
					tempMap[strconv.Itoa(index)] = tag
				}
				index++
			}
			continue
		}
	}
	//get array of ordered map

	orderdArray := MapToArrayConverter(tempMap)

	queryString = ""

	for _, value := range orderdArray {
		if value == "AND" || value == "OR" {
			queryString = queryString + " " + value + " "
		} else {
			queryString = queryString + value + ""
		}
	}

	return
}
