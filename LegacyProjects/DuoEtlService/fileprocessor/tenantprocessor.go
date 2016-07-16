package fileprocessor

import (
	"io/ioutil"
)

func GetClassPaths(rootPath string) []string {
	var classpaths []string
	//get all namespace paths
	namespaces := getFolderNames(rootPath)
	//iterate and get all classes and append to classpaths
	for _, namespace := range namespaces {
		classes := getFolderNames((rootPath + "/" + namespace))
		for _, class := range classes {
			classpaths = append(classpaths, (rootPath + "/" + namespace + "/" + class))
		}
	}
	return classpaths
}

func getFolderNames(rootPath string) []string {
	var folders map[int]string
	folders = make(map[int]string)
	index := 0
	files, _ := ioutil.ReadDir(rootPath)
	for _, f := range files {
		if f.IsDir() {
			folders[index] = f.Name()
			index++
		}
	}

	var folderArray []string
	folderArray = make([]string, len(folders))
	index = 0
	for _, value := range folders {
		folderArray[index] = value
		index++
	}
	return folderArray
}
