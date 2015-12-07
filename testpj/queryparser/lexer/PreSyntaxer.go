package lexer

import (
	"fmt"
	"strconv"
)

type PreSyntaxer struct {
}

func (p PreSyntaxer) CheckPrimarySyntax(input map[string]string) (isCorrect bool) {

	//This method will check words against keywords
	isCorrect = true
	//Check for multipleKeywords
	if CheckMultipleKeywords(input) {
		isCorrect = false
	}

	if isCorrect {
		//Check for legal word usage
		if CheckIllegalKeywordUse(input) {
			isCorrect = false
		}
	}
	return

}

func CheckIllegalKeywordUse(input map[string]string) (isIllegal bool) {

	var checkMaps map[string]string
	checkMaps = make(map[string]string)

	isIllegal = false

	excludeList := []string{"select", "from", "where", "order by", "asc", "desc"}

	isThere := false
	index := 0
	for _, value := range input {
		for _, arrVal := range excludeList {
			if value != arrVal {
				isThere = false
			} else {
				isThere = true
				break
			}
		}
		if !isThere {
			checkMaps[strconv.Itoa(index)] = value
			index++
		}
	}
	isIllegal = CheckIfIllegal(checkMaps)

	return

}

func CheckIfIllegal(input map[string]string) (isIllegal bool) {
	//QUERY MAP = The whole SQL query
	//Input MAP = Keywords for GETWORDSLICE method

	isIllegal = false

	//Get restricted Keywords
	keywordDirectory := KeywordDirectory{}
	restrictedKeywords := keywordDirectory.GetKeywords()

	//Convert to lower case so can crossed checked with keywords
	restrictedKeywords = ConvertToLowerCase(restrictedKeywords)

	//check maps agaist restricted keywords

	for _, value := range input {

		for _, restricted := range restrictedKeywords {
			if value == restricted {
				fmt.Println("Failed at : " + value + " : " + restricted)
				isIllegal = true
				break
			}
		}
	}
	return

}

func CheckMultipleKeywords(input map[string]string) (isMultiEntry bool) {
	checkingWords := []string{"select", "from", "where", "group by", "order by"}
	isMultiEntry = false

	for _, value := range checkingWords {
		if CheckMultiEntry(input, value) {
			isMultiEntry = true
			break
		}
	}
	return
}

func GetWordCount(inputMap map[string]string, matchString string) (count int) {

	count = 0
	for _, value := range inputMap {
		if value == matchString {
			count++
		}
	}
	return
}

func CheckMultiEntry(inputMap map[string]string, matchString string) (isMultiEntry bool) {
	isMultiEntry = false

	if GetWordCount(inputMap, matchString) > 1 {
		isMultiEntry = true
	}
	return
}

func GetWordPosition(array []string, keyword string) (index int) {

	index = -1
	for key, value := range array {
		if value == keyword {
			index = key
			break
		}
	}
	return
}
