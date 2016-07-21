package endpoints

import (
	"strings"
)

func AAA(input []byte) (output []byte) {
	correctedSingleQuotes := singleQuotes(string(input))
	output = []byte(correctedSingleQuotes)
	return
}

func singleQuotes(inputQuery string) (correctedString string) {
	input := inputQuery
	inputCopy := inputQuery
	lastPoint := 0
	splitArray := make([]string, 0)
	for x := 0; x < len(input); x++ {
		if (string(input[x]) == "'") && (string(input[x+1]) != "'") && (string(input[x-1]) != "'") {
			stringPart := input[lastPoint : x+1]
			stringPart = strings.Replace(stringPart, "'", "''", -1)
			splitArray = append(splitArray, stringPart)
			inputCopy = input[x+1:]
			lastPoint = x + 1
		}
	}

	if len(inputCopy) > 0 {
		splitArray = append(splitArray, inputCopy)
		inputCopy = ""
	}
	correctedString = strings.Join(splitArray, "")
	return
}
