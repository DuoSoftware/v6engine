package processes

import (
	"encoding/json"
	"fmt"
	"github.com/couchbaselabs/go-couchbase"
	"io/ioutil"
	//	"path/filepath"
)

func ExportToCouchServer(ipAddress string, bucket string) (status bool) {
	status = true
	for _, value := range GetBackupFileList() {
		content, _ := ioutil.ReadFile(value)
		var array []map[string]interface{}
		_ = json.Unmarshal(content, &array)
		namespace, class := getNamespaceAndClass(value)
		status = ExportCouch(ipAddress, namespace, class, array, bucket)
	}
	return
}

/*func GetBackupFileList() []string {
	files1, _ := filepath.Glob("*.txt")
	return files1
}*/

func ExportCouch(ipAddress string, namespace string, class string, array []map[string]interface{}, bucket string) (status bool) {
	conn, _, _ := getCouchBucket(ipAddress, bucket)
	status = true
	for _, obj := range array {
		nosqlid := getRecordID(obj)
		fmt.Println("Restoring Namespace : " + namespace + " Class : " + class)
		var allMaps map[string]interface{}
		allMaps = make(map[string]interface{})

		for key, value := range obj {

			if key != "OriginalIndex" {
				allMaps[key] = value
			} else {
				//do nothing
			}
		}

		err := conn.Set(nosqlid, 0, allMaps)

		if err != nil {
			status = false
		} else {
			status = true
		}
	}
	return
}

/*func getRecordID(inputMap map[string]interface{}) (key string, class string, namespace string) {
	var keyproperty string
	var Class string
	var Namespace string

	for key, value := range inputMap {
		if key == "OriginalIndex" {
			keyproperty = value.(string)
		}

		if key == "__osHeaders" {
			for k1, v1 := range value.(map[string]interface{}) {
				if k1 == "Class" {
					Class = v1.(string)
				}
				if k1 == "Namespace" {
					Namespace = v1.(string)
				}
			}
		}
	}
	key = keyproperty
	class = Class
	namespace = Namespace
	return
}
*/
func getCouchBucket(url string, bucketName string) (bucket *couchbase.Bucket, errorMessage string, isError bool) {

	isError = false
	fmt.Println("Getting store configuration settings for Couchbase")

	setting_host := url
	setting_bucket := bucketName
	fmt.Println("Store configuration settings recieved for Couchbase Host : " + setting_host + " , Bucket : " + setting_bucket)

	c, err := couchbase.Connect(setting_host)
	if err != nil {
		isError = true
		errorMessage = "Error connecting Couchbase to :  " + setting_host
		fmt.Println(errorMessage)
	}

	pool, err := c.GetPool("default")
	if err != nil {
		isError = true
		errorMessage = "Error getting pool: "
		fmt.Println(errorMessage)
	}

	returnBucket, err := pool.GetBucket(setting_bucket)

	if err != nil {
		isError = true
		errorMessage = "Error getting Couchbase bucket: " + setting_bucket
		fmt.Println(errorMessage)
	} else {
		fmt.Println("Successfully recieved Couchbase bucket")
		bucket = returnBucket
	}

	return
}
