package JSON_Purifier

import (
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"strings"
)

func Purify(input []byte) (output []byte) {

	var requestBody messaging.RequestBody
	err := json.Unmarshal(input, &requestBody)

	if err != nil {
		return
	}

	if requestBody.Object != nil || requestBody.Objects != nil {
		output = []byte(CorrectSingleQuotes(string(input)))
	} else {
		output = input
	}
	return
}

func CorrectSingleQuotes(inputQuery string) (correctedString string) {
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
