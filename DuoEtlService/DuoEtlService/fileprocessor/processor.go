package fileprocessor

import (
	"duov6.com/DuoEtlService/logger"
	"duov6.com/DuoEtlService/messaging"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Process(rootPath string, operation string) []messaging.RequestBody {
	var allObjects []messaging.RequestBody

	logger.Log("Starting File Processor.....")

	//read the file list from the NEW folder
	readFilePath := ""
	if operation == "POST" {
		readFilePath = rootPath + "/add"
	} else if operation == "PUT" {
		readFilePath = rootPath + "/edit"
	} else if operation == "DELETE" {
		readFilePath = rootPath + "/delete"
	}

	fmt.Println("File Path : ")
	fmt.Println(readFilePath)
	fileNames := getFileList(readFilePath + "/new/")
	fmt.Println("File Names : ")
	fmt.Println(fileNames)
	//Copy those files to PROCESSING folderand
	for _, fileName := range fileNames {
		moveFileToProcessingFolder(readFilePath, operation, fileName)
	}
	//Read one by one and return as Objects
	newFileNames := getFileList(readFilePath + "/processing/")
	fmt.Println("Updated file Names : ")
	fmt.Println(newFileNames)

	for _, newFileName := range newFileNames {
		allObjects = append(allObjects, getObjectsFromFiles(newFileName))
	}
	//fmt.Println("All Objects : ")
	//fmt.Println(allObjects)
	return allObjects
}

func getFileList(rootPath string) []string {
	allFiles, _ := filepath.Glob(rootPath + "*.txt")
	return allFiles
}

func moveFileToProcessingFolder(rootPath string, operation string, filename string) {
	tokens := strings.Split(filename, "/")

	content, err := ioutil.ReadFile(filename)
	err = ioutil.WriteFile((rootPath + "/processing/" + tokens[(len(tokens)-1)]), content, 0666)
	err = os.Remove(filename)

	if err != nil {
		logger.Log(err.Error())
	}
}

func getObjectsFromFiles(filename string) messaging.RequestBody {
	var object messaging.RequestBody
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Log(err.Error())
	} else {
		err = json.Unmarshal(content, &object)
		if err != nil {
			logger.Log(err.Error())
		}
	}
	return object
}

func ClearCompletedFiles(rootPath string) {
	//move to COMPLETED folder and delete all from processing folder

	//Move completed POST files

	for _, fileName := range getFileList(rootPath + "/add/processing/") {
		tokens := strings.Split(fileName, "/")
		content, err := ioutil.ReadFile(fileName)
		err = ioutil.WriteFile((rootPath + "/add/completed/" + tokens[(len(tokens)-1)]), content, 0666)
		err = os.Remove(fileName)
		if err != nil {
			logger.Log(err.Error())
		}
	}

	//Move completed PUT files

	for _, fileName := range getFileList(rootPath + "/edit/processing/") {
		tokens := strings.Split(fileName, "/")
		content, err := ioutil.ReadFile(fileName)
		err = ioutil.WriteFile((rootPath + "/edit/completed/" + tokens[(len(tokens)-1)]), content, 0666)
		err = os.Remove(fileName)
		if err != nil {
			logger.Log(err.Error())
		}
	}

	//Move completed DELETE files

	for _, fileName := range getFileList(rootPath + "/delete/processing/") {
		tokens := strings.Split(fileName, "/")
		content, err := ioutil.ReadFile(fileName)
		err = ioutil.WriteFile((rootPath + "/delete/completed/" + tokens[(len(tokens)-1)]), content, 0666)
		err = os.Remove(fileName)
		if err != nil {
			logger.Log(err.Error())
		}
	}
}
