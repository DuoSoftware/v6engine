package processes

import (
	"encoding/json"
	"github.com/mattbaird/elastigo/lib"
	"io/ioutil"
	"strings"
)

func BackupServer(ipAddress string) (status bool) {
	status = true
	for _, namespace := range getNamespaces(ipAddress) {
		for _, class := range getClasses(namespace, ipAddress) {
			records := getInstanceData(namespace, class, ipAddress)
			var array []map[string]interface{}
			array = make([]map[string]interface{}, len(records))
			for key, value := range records {
				array[key] = value
			}
			recordByte, err := json.Marshal(array)
			if err != nil {
				status = false
			}
			ioutil.WriteFile((namespace + "-" + class + ".txt"), recordByte, 0666)
		}

	}

	return
}

func getConnection(ipAddr string) *elastigo.Conn {

	var connection *elastigo.Conn
	if connection == nil {
		ipParts := strings.Split(ipAddr, ":")

		conn := elastigo.NewConn()
		conn.SetHosts([]string{ipParts[0]})
		conn.Port = ipParts[1]
		connection = conn
	}
	return connection
}

func getNamespaces(ipAddress string) map[int]string {
	var returnMap map[int]string
	returnMap = make(map[int]string)
	conn := getConnection(ipAddress)
	skip := "0"
	take := "1000000"

	query := "{\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + "*" + "\"}}}"

	data, err := conn.Search("", "", nil, query)

	if err != nil {
	} else {
		var allMaps []map[string]interface{}
		allMaps = make([]map[string]interface{}, data.Hits.Len())

		for index, hit := range data.Hits.Hits {
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})

			byteData, _ := hit.Source.MarshalJSON()

			json.Unmarshal(byteData, &currentMap)

			allMaps[index] = currentMap
		}

		var m map[int]string
		m = make(map[int]string)

		count := 0
		for _, value := range allMaps {
			for key, oo := range value {
				if key == "__osHeaders" {
					for key2, mapContent := range oo.(map[string]interface{}) {
						if key2 == "Namespace" {
							m[count] = mapContent.(string)
							count++
						}
					}
				}
			}
		}

		//Get Unique Class Names
		mm := getUniqueRecordMap(m)
		returnMap = mm

	}

	return returnMap
}

func getClasses(namespace string, ipAddress string) map[int]string {
	var returnMap map[int]string
	returnMap = make(map[int]string)
	conn := getConnection(ipAddress)
	skip := "0"
	take := "100000000"

	query := "{\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + "*" + "\"}}}"
	data, err := conn.Search("", "", nil, query)

	if err != nil {
	} else {

		var allMaps []map[string]interface{}
		allMaps = make([]map[string]interface{}, data.Hits.Len())

		for index, hit := range data.Hits.Hits {
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})

			byteData, _ := hit.Source.MarshalJSON()

			json.Unmarshal(byteData, &currentMap)

			allMaps[index] = currentMap
		}
		var m map[int]string
		m = make(map[int]string)

		count := 0
		for _, value := range allMaps {
			for key, oo := range value {
				if key == "__osHeaders" {

					var tempmap map[string]string
					tempmap = make(map[string]string)

					for key2, mapContent := range oo.(map[string]interface{}) {

						if key2 == "Namespace" {
							tempmap[key2] = mapContent.(string)
						}

						if key2 == "Class" {
							tempmap[key2] = mapContent.(string)
						}

					}

					if tempmap["Namespace"] == namespace {
						m[count] = tempmap["Class"]
						count++
					}

				}
			}
		}

		//Get Unique Class Names
		mm := getUniqueRecordMap(m)
		returnMap = mm
	}
	return returnMap
}

func getInstanceData(namespace string, class string, ipAddress string) map[int]map[string]interface{} {
	var returnMap map[int]map[string]interface{}
	returnMap = make(map[int]map[string]interface{})
	conn := getConnection(ipAddress)
	skip := "0"
	take := "100000000"

	query := "{\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + "*" + "\"}}}"
	data, err := conn.Search("", "", nil, query)

	if err != nil {
	} else {

		var allMaps []map[string]interface{}
		allMaps = make([]map[string]interface{}, data.Hits.Len())

		for index, hit := range data.Hits.Hits {
			var currentMap map[string]interface{}
			currentMap = make(map[string]interface{})
			currentMap["OriginalIndex"] = hit.Id
			byteData, _ := hit.Source.MarshalJSON()

			json.Unmarshal(byteData, &currentMap)

			allMaps[index] = currentMap
		}

		//get count of data
		count := 0
		for _, value := range allMaps {
			for key, oo := range value {
				if key == "__osHeaders" {

					var tempmap map[string]string
					tempmap = make(map[string]string)

					for key2, mapContent := range oo.(map[string]interface{}) {

						if key2 == "Namespace" {
							tempmap[key2] = mapContent.(string)
						}

						if key2 == "Class" {
							tempmap[key2] = mapContent.(string)
						}

					}

					if tempmap["Namespace"] == namespace && tempmap["Class"] == class {
						count++
					}

				}
			}
		}

		var m map[int]map[string]interface{}
		m = make(map[int]map[string]interface{})

		count = 0
		for _, value := range allMaps {
			for key, oo := range value {
				if key == "__osHeaders" {

					var tempmap map[string]string
					tempmap = make(map[string]string)

					for key2, mapContent := range oo.(map[string]interface{}) {

						if key2 == "Namespace" {
							tempmap[key2] = mapContent.(string)
						}

						if key2 == "Class" {
							tempmap[key2] = mapContent.(string)
						}

					}

					if tempmap["Namespace"] == namespace && tempmap["Class"] == class {
						m[count] = value
						count++
					}

				}
			}
		}

		returnMap = m
	}

	return returnMap
}

func getUniqueRecordMap(inputMap map[int]string) map[int]string {

	var outputMap map[int]string
	outputMap = make(map[int]string)

	count := 0
	for _, value := range inputMap {
		if len(outputMap) == 0 {
			outputMap[count] = value
			count++
		} else {
			isAvailable := false
			for _, value2 := range outputMap {
				if value == value2 {
					isAvailable = true
				}
			}
			if isAvailable {
				//Do Nothing
			} else {
				outputMap[count] = value
				count++
			}
		}
	}
	return outputMap
}
