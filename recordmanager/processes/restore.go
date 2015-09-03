package processes

import (
	"encoding/json"
	//"fmt"
	//	"github.com/mattbaird/elastigo/lib"
	"io/ioutil"
	"path/filepath"
	//	"strings"
	//"time"
)

func RestoreServer(ipAddress string) (status bool) {
	status = true
	for _, value := range GetBackupFileList() {
		content, _ := ioutil.ReadFile(value)
		var array []map[string]interface{}
		array = make([]map[string]interface{}, 9999999)
		_ = json.Unmarshal(content, &array)
		status = InsertElastic(ipAddress, array)
	}

	return

}

func GetBackupFileList() []string {
	files1, _ := filepath.Glob("*.txt")
	return files1
}

func InsertElastic(ipAddress string, array []map[string]interface{}) (status bool) {
	conn := getConnection(ipAddress)
	status = true
	for _, obj := range array {
		nosqlid, class := getRecordID(obj)

		var allMaps map[string]interface{}
		allMaps = make(map[string]interface{})

		for key, value := range obj {

			if key != "OriginalIndex" {
				allMaps[key] = value
			} else {
				//do nothing
			}
		}

		_, err := conn.Index(class, class, nosqlid, nil, allMaps)

		if err != nil {
			status = false
		} else {
			status = true
		}
	}
	return
}

func getRecordID(inputMap map[string]interface{}) (key string, class string) {
	var keyproperty string
	var Class string

	for key, value := range inputMap {
		if key == "OriginalIndex" {
			keyproperty = value.(string)
		}

		if key == "__osHeaders" {
			for k1, v1 := range value.(map[string]interface{}) {
				if k1 == "Class" {
					Class = v1.(string)
				}
			}
		}
	}
	key = keyproperty
	class = Class
	return
}
