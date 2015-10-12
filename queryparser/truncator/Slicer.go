package truncator

import (
	"duov6.com/queryparser/messaging"
	//"fmt"
	"strconv"
)

type Slicer struct {
}

func (s *Slicer) Begin(request *messaging.ParserRequest) (response messaging.ParserResponse) {

	response.IsSuccess = true

	//Identify Operations
	operationList := GetOperationKeywords(request.Body)

	if operationList == nil {
		response.IsSuccess = false
	}

	//get attributes and assign it to Attributes struct
	var attributeList messaging.Attributes
	attributeList = AttributeExtractor(operationList, request.Body)

	if attributeList.Select == nil && attributeList.From == "" {
		response.IsSuccess = false
	} else {
		response.QueryItems = attributeList
	}

	return response
}

func GetOperationKeywords(inMap map[string]string) (outMap map[string]string) {

	outMap = make(map[string]string)

	allowedOperations := []string{"select", "from", "where", "order by"}

	index := 0
	for _, value := range inMap {
		for _, allowedValue := range allowedOperations {
			if value == allowedValue {
				outMap[strconv.Itoa(index)] = value
				index++
				continue
			}
		}
	}
	return
}

func AttributeExtractor(operationList map[string]string, queryItems map[string]string) (outAttributes messaging.Attributes) {
	outAttributes = messaging.Attributes{}

	for _, value := range operationList {
		if value == "select" {
			outAttributes.Select = ExtractSelectAttributes(queryItems)
		}
		if value == "from" {
			outAttributes.From = ExtractFromAttributes(queryItems)
		}
		if value == "where" {
			outAttributes.Where = ExtractWhereAttributes(queryItems)
		}
		if value == "order by" {
			outAttributes.Orderby = ExtractOrderByAttributes(queryItems)
		}
	}
	return
}

func ExtractSelectAttributes(queryItems map[string]string) (outAttributes map[string]string) {

	outAttributes = make(map[string]string)

	outAttributes = GetWordSlice("select", "from", queryItems)

	//Remove Commas, whitespaces and nullspaces

	var tempMap map[string]string
	tempMap = make(map[string]string)

	for key, value := range outAttributes {
		if value != "," && value != "" && value != " " {
			tempMap[key] = value
		}
	}

	outAttributes = tempMap

	return
}

func ExtractFromAttributes(queryItems map[string]string) (outAttribute string) {

	arr := MapToArrayConverter(queryItems)

	for key, value := range arr {
		if value == "from" {
			outAttribute = arr[key+1]
		}
	}

	return
}

func ExtractWhereAttributes(queryItems map[string]string) (outAttributes map[string]string) {

	outAttributes = make(map[string]string)

	arr := MapToArrayConverter(queryItems)

	isOrderbyExist := false

	for _, value := range arr {
		if value == "order by" {
			isOrderbyExist = true
		}
	}

	if isOrderbyExist {
		//if ther is an order by clause
		outAttributes = GetWordSlice("where", "order by", queryItems)
	} else {
		//if there is NO order by clause
		outAttributes = GetWordSlice("where", "", queryItems)
	}

	return
}

func ExtractOrderByAttributes(queryItems map[string]string) (outAttributes map[string]string) {

	outAttributes = make(map[string]string)

	outAttributes = GetWordSlice("order by", "", queryItems)

	return
}

func GetWordSlice(start string, end string, input map[string]string) (output map[string]string) {

	output = make(map[string]string)

	//sort map to an array
	sortedArray := MapToArrayConverter(input)

	if start != "" && end != "" {

		//get starting and ending indexes
		startIndex := GetKeywordPosition(sortedArray, start)
		endIndex := GetKeywordPosition(sortedArray, end)

		//get slice to temp array

		tempSlice := sortedArray[(startIndex + 1):endIndex]
		//fmt.Println(tempSlice)
		//copy slice to map and return

		for key, value := range tempSlice {
			output[strconv.Itoa(key)] = value
		}
	} else if start != "" && end == "" {
		//get starting and ending indexes
		startIndex := GetKeywordPosition(sortedArray, start)
		//get slice to temp array

		tempSlice := sortedArray[(startIndex + 1):]

		//copy slice to map and return

		for key, value := range tempSlice {
			output[strconv.Itoa(key)] = value
		}
	} else if start == "" && end != "" {
		//get starting and ending indexes
		endIndex := GetKeywordPosition(sortedArray, end)
		//get slice to temp array

		tempSlice := sortedArray[:endIndex]

		//copy slice to map and return

		for key, value := range tempSlice {
			output[strconv.Itoa(key)] = value
		}
	}

	return
}

func GetKeywordPosition(array []string, keyword string) (index int) {

	index = -1
	for key, value := range array {
		if value == keyword {
			index = key
			break
		}
	}
	return
}

func MapToArrayConverter(inputMap map[string]string) (outArray []string) {

	noOfItems := len(inputMap)

	outArray = make([]string, noOfItems)

	for key, value := range inputMap {
		index, _ := strconv.Atoi(key)
		outArray[index] = value
		index++
	}
	return
}
