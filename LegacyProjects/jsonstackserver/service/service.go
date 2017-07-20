package service

import (
	"bytes"
	"duov6.com/common"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func Log(message string) {
	fmt.Println(message)
	common.PublishLog("JsonStackLog.log", message)
}

func Start(etlhost string) {
	//get new files for all insert, update and delete
	postFileNames := getPostFiles()
	putFileNames := getPutFiles()
	deleteFileNames := getDeleteFiles()

	//post them in loop
	//INSERTS
	if len(postFileNames) > 0 {
		for _, filename := range postFileNames {
			namespace, class := getDomainData(filename)
			insert(readFileContent(filename), etlhost, namespace, class)
			createNewFile("JsonStack/Old/POST/"+getAbstractFileName(filename), readFileContent(filename))
			deleteExistingFile(filename)
		}
	} else {
		Log("No new POST Objects Found!")
	}
	//UPDATES
	if len(putFileNames) > 0 {
		for _, filename := range putFileNames {
			namespace, class := getDomainData(filename)
			update(readFileContent(filename), etlhost, namespace, class)
			createNewFile("JsonStack/Old/PUT/"+getAbstractFileName(filename), readFileContent(filename))
			deleteExistingFile(filename)
		}
	} else {
		Log("No new PUT Objects Found!")
	}
	//DELETES
	if len(deleteFileNames) > 0 {
		for _, filename := range deleteFileNames {
			namespace, class := getDomainData(filename)
			delete(readFileContent(filename), etlhost, namespace, class)
			createNewFile("JsonStack/Old/DELETE/"+getAbstractFileName(filename), readFileContent(filename))
			deleteExistingFile(filename)
		}
	} else {
		Log("No new DELETE Objects Found!")
	}

	Log("Transfer Cycle Completed!")
}

func getAbstractFileName(fileName string) (abstractName string) {
	tokens := strings.Split(fileName, "/")
	Log("Abstract File Name : " + tokens[(len(tokens)-1)])
	return tokens[(len(tokens) - 1)]
}

func getPostFiles() []string {
	return getFileList("JsonStack/New/POST/*.txt")
}

func getPutFiles() []string {
	return getFileList("JsonStack/New/PUT/*.txt")
}

func getDeleteFiles() []string {
	return getFileList("JsonStack/New/DELETE/*.txt")
}

func readFileContent(filename string) (content []byte) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		Log(err.Error())
		content = nil
	}
	return content
}

func createNewFile(filename string, content []byte) {
	Log("Creating : " + getAbstractFileName(filename))
	err := ioutil.WriteFile(filename, content, 0666)
	if err != nil {
		Log(err.Error())
	}
}

func deleteExistingFile(filename string) {
	Log("Deleting : " + getAbstractFileName(filename))
	err := os.Remove(filename)
	if err != nil {
		Log(err.Error())
	}
}

func getFileList(path string) []string {
	files, _ := filepath.Glob(path)
	return files
}

func getDomainData(fileName string) (namespace string, class string) {
	abstractFileName := getAbstractFileName(fileName)
	tokens := strings.Split(abstractFileName, "-")
	namespace = tokens[0]
	class = tokens[1]
	return
}

func insert(JSON_Document []byte, host string, namespace string, class string) {

	url := "http://" + host + "/" + namespace + "/" + class
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(JSON_Document))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		Log(err.Error())
	} else {
		Log("Successfully transferred one INSERT object to ETL interface!")
	}
	defer resp.Body.Close()

}

func delete(JSON_Document []byte, host string, namespace string, class string) {
	url := "http://" + host + "/" + namespace + "/" + class
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(JSON_Document))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		Log(err.Error())
	} else {
		Log("Successfully transferred one DELETE object to ETL interface!")
	}
	defer resp.Body.Close()
}

func update(JSON_Document []byte, host string, namespace string, class string) {
	url := "http://" + host + "/" + namespace + "/" + class
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(JSON_Document))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		Log(err.Error())
	} else {
		Log("Successfully transferred one UPDATE object to ETL interface!")
	}
	defer resp.Body.Close()
}
