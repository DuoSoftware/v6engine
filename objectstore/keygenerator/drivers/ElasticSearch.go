package drivers

import (
	"duov6.com/common"
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"fmt"
	"github.com/mattbaird/elastigo/lib"
	"strconv"
)

type ElasticSearch struct {
}

func (driver ElasticSearch) getConnection(request *messaging.ObjectRequest) (connection *elastigo.Conn) {
	host := request.Configuration.ServerConfiguration["ELASTIC"]["Host"]
	port := request.Configuration.ServerConfiguration["ELASTIC"]["Port"]

	conn := elastigo.NewConn()
	conn.SetHosts([]string{host})
	conn.Port = port
	connection = conn
	return
}

func (driver ElasticSearch) VerifyMaxValueDB(request *messaging.ObjectRequest, amount int) (maxValue string) {
	conn := driver.getConnection(request)
	class := request.Controls.Class
	key := getDomainClassAttributesKey(request)

	var myMap map[string]interface{}
	data, err := conn.Get(request.Controls.Namespace, "domainClassAttributes", key, nil)
	if err != nil {
		myMap = make(map[string]interface{})
	} else {
		bytes, _ := data.Source.MarshalJSON()
		json.Unmarshal(bytes, &myMap)
	}

	if len(myMap) == 0 {
		maxValue = strconv.Itoa(amount + 1)

		object := make(map[string]interface{})
		object["__os_id"] = key
		object["class"] = class
		object["maxCount"] = maxValue
		object["version"] = common.GetGUID()

		_, err = conn.Index(request.Controls.Namespace, "domainClassAttributes", key, nil, object)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	} else {
		maxCount, err := strconv.Atoi(myMap["maxCount"].(string))
		if maxCount <= amount {
			maxCount = amount + 1
		} else {
			maxCount += 1
		}
		maxValue = strconv.Itoa(maxCount)

		object := make(map[string]interface{})
		object["__os_id"] = key
		object["class"] = class
		object["maxCount"] = maxValue
		object["version"] = common.GetGUID()

		_, err = conn.Index(request.Controls.Namespace, "domainClassAttributes", key, nil, object)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	return
}
