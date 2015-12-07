package lexer

import (
	"bytes"
	"duov6.com/queryparser/messaging"
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

type Tokenizer struct {
}

func (t *Tokenizer) GetTokens(request *messaging.ParserRequest) (response messaging.ParserResponse) {

	//Get Basic Tokens
	fmt.Println("Getting Basic Tokens")
	response = t.CreateTokens(request)
	fmt.Println("******")
	fmt.Println(response.Body)
	fmt.Println("******")

	//Convert all to Lower Case
	response.Body = ConvertToLowerCase(response.Body)

	//Normalize Tokens
	fmt.Println("Normalizing Basic Tokens")
	response.Body = NormalizeTokens(response.Body)

	//Check for grammar
	fmt.Println("Checking Tokens for Grammar")
	grammar := PreSyntaxer{}
	isCorrect := grammar.CheckPrimarySyntax(response.Body)

	if !isCorrect {
		fmt.Println("Incorrect or Unsupported SQL Syntax! Check query again!")
		response.Message = "Incorrect or Unsupported SQL Syntax! Check query again!"
		response.IsSuccess = false
		return
	} else {
		fmt.Println("Correct SQL Syntax!")
		response.Message = "Correct SQL Syntax!"
		response.IsSuccess = true
	}

	return
}

func (t Tokenizer) CreateTokens(request *messaging.ParserRequest) (response messaging.ParserResponse) {

	response.Body = make(map[string]string)
	b := bytes.NewBufferString(request.Query)
	var s scanner.Scanner
	s.Init(b)

	index := 0
	for {
		tok := s.Scan()
		if tok != scanner.EOF {
			response.Body[strconv.Itoa(index)] = s.TokenText()
			index++
		} else {
			break
		}

	}

	if index == 0 {
		response.IsSuccess = false
		response.Message = "Error! Nil Query Allocated!"
	} else {
		response.IsSuccess = false
		response.Message = "Success! Query Tokenized Successfully!"
	}

	return
}

func NormalizeTokens(inputMap map[string]string) (outMap map[string]string) {

	outMap = make(map[string]string)

	tempArray := MapToArrayConverter(inputMap)

	index := 0
	for x := 0; x < len(tempArray); x++ {

		if (tempArray[x] == "group") && (tempArray[x+1] == "by") {
			outMap[strconv.Itoa(index)] = "group by"
			index++
		} else if (tempArray[x] == "order") && (tempArray[x+1] == "by") {
			outMap[strconv.Itoa(index)] = "order by"
			index++
		} else {
			if tempArray[x] != "by" {
				outMap[strconv.Itoa(index)] = tempArray[x]
				index++
			}
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

func ConvertToLowerCase(inputMap map[string]string) (outputMap map[string]string) {

	outputMap = make(map[string]string)

	for key, value := range inputMap {
		outputMap[key] = strings.ToLower(value)
	}
	return
}

func ConvertToUpperCase(inputMap map[string]string) (outputMap map[string]string) {

	outputMap = make(map[string]string)

	for key, value := range inputMap {
		outputMap[key] = strings.ToUpper(value)
	}
	return
}

func GetWordSlice(start string, end string, input map[string]string) (output map[string]string) {

	output = make(map[string]string)

	//sort map to an array
	sortedArray := MapToArrayConverter(input)

	if start != "" && end != "" {

		//get starting and ending indexes
		startIndex := GetWordPosition(sortedArray, start)
		endIndex := GetWordPosition(sortedArray, end)

		//get slice to temp array

		tempSlice := sortedArray[(startIndex + 1):endIndex]
		fmt.Println(tempSlice)
		//copy slice to map and return

		for key, value := range tempSlice {
			output[strconv.Itoa(key)] = value
		}
	} else if start != "" && end == "" {
		//get starting and ending indexes
		startIndex := GetWordPosition(sortedArray, start)
		//get slice to temp array

		tempSlice := sortedArray[(startIndex + 1):]

		//copy slice to map and return

		for key, value := range tempSlice {
			output[strconv.Itoa(key)] = value
		}
	} else if start == "" && end != "" {
		//get starting and ending indexes
		endIndex := GetWordPosition(sortedArray, end)
		//get slice to temp array

		tempSlice := sortedArray[:endIndex]

		//copy slice to map and return

		for key, value := range tempSlice {
			output[strconv.Itoa(key)] = value
		}
	}

	return
}
