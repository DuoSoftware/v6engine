package statgen

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
)

func getAllFiles() (fileNames map[int]string) {
	fileNames = make(map[int]string)

	sucFiles := getSuccessFiles()
	erFiles := getErrorFiles()

	index := 0

	for _, value := range sucFiles {
		fileNames[index] = value
		index++
	}

	for _, value := range erFiles {
		fileNames[index] = value
		index++
	}

	return fileNames
}

func getErrorFiles() (fileNames map[int]string) {
	fileNames = make(map[int]string)

	tempFiles := readDisk("*.err")

	for index, value := range tempFiles {
		fileNames[index] = value
	}

	return fileNames
}

func getSuccessFiles() (fileNames map[int]string) {
	fileNames = make(map[int]string)

	tempFiles := readDisk("*.suc")

	for index, value := range tempFiles {
		fileNames[index] = value
	}

	return fileNames
}

func getFilesByPattern(pattern string) (fileNames map[int]string) {
	fileNames = make(map[int]string)

	tempFiles := readDisk("duov6.com/duoauth/" + pattern)

	for index, value := range tempFiles {
		fileNames[index] = value
	}
	return fileNames
}

func readDisk(pattern string) []string {
	files1, _ := filepath.Glob(pattern)
	return files1
}

func readFileContent(fileName string) (array []record) {

	content, _ := ioutil.ReadFile(fileName)
	records := strings.Split(string(content), "\r")

	actualRecordCount := 0

	for _, tempVal := range records {
		if tempVal == "\r" || tempVal == "\n" {
			//Do Nothing..  Just eliminating EOF
		} else {
			actualRecordCount++
		}
	}

	array = make([]record, actualRecordCount)

	for index, singleRecord := range records {
		if singleRecord == "\r" || singleRecord == "\n" {
			//Do Nothing.. Just eliminating EOF
		} else {
			temp := []byte(singleRecord)
			tempRecord := record{}
			_ = json.Unmarshal(temp, &tempRecord)
			array[index] = tempRecord
		}
	}
	return
}
