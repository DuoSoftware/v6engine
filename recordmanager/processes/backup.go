package processes

import (
	"encoding/json"
	"fmt"
	"github.com/mattbaird/elastigo/lib"
	"io/ioutil"
	"net/http"
	"strings"
)

func BackupServer(ipAddress string) (status bool) {
	status = true
	for _, namespace := range getDownloadAllowList(getNamespaces(ipAddress)) {
		for _, class := range getClasses(namespace, ipAddress) {
			fmt.Println("Backing up Namespace : " + namespace + " Class : " + class)
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
			ioutil.WriteFile((namespace + "-" + class + ".objectfile"), recordByte, 0666)
		}

	}

	return
}

func getDownloadAllowList(allList []string) []string {
	var returnList []string
	content, _ := ioutil.ReadFile("downloadAllowList.config")
	var allowList []string
	err := json.Unmarshal(content, &allowList)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(allowList)
	if allowList[0] == "*" {
		fmt.Println("Get All Namespaces!")
		returnList = allList
	} else {
		fmt.Println("Get Selected Namespaces!")
		for _, allValue := range allList {
			for _, allowValue := range allowList {
				if allValue == allowValue {
					returnList = append(returnList, allValue)
					break
				}
			}
		}
	}

	fmt.Print("Allowed List : ")
	fmt.Println(returnList)
	return returnList
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

func getElasticByCURL(url string, path string) (returnByte []byte) {
	url = "http://" + url + "/" + path
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("CURL Request Failed")
	} else {
		fmt.Println("CURL Request Success!")
		body, _ := ioutil.ReadAll(resp.Body)
		returnByte = body
	}
	defer resp.Body.Close()

	return
}

func getNamespaces(ipAddress string) []string {
	returnByte := getElasticByCURL(ipAddress, ("_mapping"))
	var mainMap map[string]interface{}
	mainMap = make(map[string]interface{})
	_ = json.Unmarshal(returnByte, &mainMap)
	var retArray []string
	//range through namespaces
	for index, _ := range mainMap {
		retArray = append(retArray, index)
	}
	return retArray
}

func getClasses(namespace string, ipAddress string) []string {
	returnByte := getElasticByCURL(ipAddress, (namespace + "/_mapping"))
	var mainMap map[string]interface{}
	mainMap = make(map[string]interface{})
	_ = json.Unmarshal(returnByte, &mainMap)
	var retArray []string
	//range through namespaces
	for _, index := range mainMap {
		for feature, typeDef := range index.(map[string]interface{}) {
			//if feature is MAPPING
			if feature == "mappings" {
				for typeName, _ := range typeDef.(map[string]interface{}) {
					retArray = append(retArray, typeName)
				}
			}
		}
	}

	return retArray
}

func getInstanceData(namespace string, class string, ipAddress string) []map[string]interface{} {
	var returnMap []map[string]interface{}
	conn := getConnection(ipAddress)
	skip := "0"
	take := "100000000"

	query := "{\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + "*" + "\"}}}"
	data, err := conn.Search(namespace, class, nil, query)

	if err != nil {
		returnMap = nil
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

		returnMap = allMaps
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

//Anomaly Remover

func RemoveAnomaly(ipAddress string) (status bool) {
	status = true
	for _, namespace := range getDownloadAllowList(getNamespaces(ipAddress)) {
		for _, class := range getClasses(namespace, ipAddress) {
			fmt.Println("Resolving Namespace : " + namespace + " Class : " + class)
			getInstanceData(namespace, class, ipAddress)

		}

	}
	return status
}

func resolveInstance(namespace string, class string, ipAddress string) {
	conn := getConnection(ipAddress)
	skip := "0"
	take := "100000000"

	query := "{\"from\": " + skip + ", \"size\": " + take + ", \"query\":{\"query_string\" : {\"query\" : \"" + "*" + "\"}}}"
	data, err := conn.Search(namespace, class, nil, query)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		for _, hit := range data.Hits.Hits {
			if strings.Contains(hit.Id, (namespace + "." + class + "." + namespace + "." + class)) {
				fmt.Println(hit.Id)
			}
		}

	}
}
