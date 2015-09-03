package dcommands

import (
	"bytes"
	"duov6.com/ceb"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func GetAllMaps() (maps []ceb.CommandMap, err error) {

	files, err := ioutil.ReadDir("./config")

	if err == nil {
		maps = make([]ceb.CommandMap, len(files))

		for index, f := range files {
			tmp := ceb.CommandMap{}

			var buffer bytes.Buffer
			buffer.WriteString("./config/")
			buffer.WriteString(f.Name())
			buffer.WriteString("/map.json")
			data, err := ioutil.ReadFile(buffer.String())

			if err == nil {
				json.Unmarshal(data, &tmp)
				maps[index] = tmp
			}

		}
	} else {
		return
	}

	return
}

func CreateSampleMaps() {
	newComm := ceb.CommandMap{}
	newComm.Name = "Sample"
	newComm.Code = "samplecommand"

	newComm.Parameters = make([]ceb.CommandParameter, 2)
	newComm.Parameters[0] = ceb.CommandParameter{}
	newComm.Parameters[0].Key = "SampleKey1"
	newComm.Parameters[0].Caption = "Sample Key"
	newComm.Parameters[0].Description = "Description of Sample Key"

	newComm.Parameters[1] = ceb.CommandParameter{}
	newComm.Parameters[1].Key = "SampleKey2"
	newComm.Parameters[1].Caption = "Sample Key 2"
	newComm.Parameters[1].Description = "Description of Sample Key 2"

	strB, _ := json.Marshal(newComm)
	fmt.Println(string(strB))

	// write whole the body
	//ioutil.WriteFile(filename, data, perm)
	err := ioutil.WriteFile("map.txt", strB, 0644)
	if err != nil {
		panic(err)
	}
}
