package queryparser

import (
	"bytes"
	//"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

func ConvertToTags(queryString string) (whereTags map[string]string, selectTags map[string]string) {

	arr := strings.Split(queryString, "select")

	//var whereTags map[string]string
	whereTags = make(map[string]string)

	//var selectTags map[string]string
	selectTags = make(map[string]string)

	if len(arr) == 1 {
		//When no SELECT is present
		whereTags = GetTags(arr[0])
	} else {
		//When Where and Select are both present
		whereTags = GetTags(arr[0])
		selectTags = GetTags(arr[1])
	}

	return

}

func GetTags(query string) (outMap map[string]string) {
	outMap = make(map[string]string)
	b := bytes.NewBufferString(query)
	var s scanner.Scanner
	s.Init(b)

	index := 0
	for {
		tok := s.Scan()
		if tok != scanner.EOF {
			outMap[strconv.Itoa(index)] = s.TokenText()
			index++
		} else {
			break
		}

	}
	return
}

func MapToArrayConverter(inputMap map[string]string) (outArray []string) {

	noOfItems := len(inputMap)

	outArray = make([]string, noOfItems)

	isLastNegatable := false
	index := 0

	for key, value := range inputMap {
		if value != "," && value != "Select" && value != "select" && value != "SELECT" {
			if isLastNegatable {
				index, _ = strconv.Atoi(key)
				index--
				isLastNegatable = false
			} else {
				index, _ = strconv.Atoi(key)
				isLastNegatable = false
			}
			outArray[index] = value
		} else {
			isLastNegatable = true
		}
		index++
	}

	return
}
