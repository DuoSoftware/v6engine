package drivers

import (
	"duov6.com/common"
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"fmt"
	"strconv"
)

type RepoDefault struct {
}

func (driver RepoDefault) VerifyMaxValueDB(request *messaging.ObjectRequest, amount int) (maxValue string) {
	class := request.Controls.Class

	headers := make(map[string]string)
	headers["securityToken"] = "ignore"

	var myMap map[string]interface{}
	err, byteArray := common.HTTP_GET("http://localhost:3000/"+request.Controls.Namespace+"/domainClassAttributes/"+class, headers, true)
	if err != nil {
		myMap = make(map[string]interface{})
	} else {
		json.Unmarshal(byteArray, &myMap)
	}

	if len(myMap) == 0 {
		maxValue = strconv.Itoa(amount)

		object := make(map[string]interface{})
		object["class"] = class
		object["maxCount"] = maxValue
		object["version"] = common.GetGUID()

		var parameterObj InsParameters
		parameterObj.KeyProperty = "class"
		var insertObj = new(InsertTemplate)
		insertObj.Object = object
		insertObj.Parameters = parameterObj

		convertedObj, _ := json.Marshal(insertObj)
		err, _ := common.HTTP_POST("http://localhost:3000/"+request.Controls.Namespace+"/domainClassAttributes/", headers, convertedObj, true)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		//count to return
		maxValue = strconv.Itoa(amount + 1)
	} else {
		maxCount, err := strconv.Atoi(myMap["maxCount"].(string))

		if maxCount < amount {
			maxCount = amount
		}

		maxValue = strconv.Itoa(maxCount)

		object := make(map[string]interface{})
		object["class"] = class
		object["maxCount"] = maxValue
		object["version"] = common.GetGUID()

		var parameterObj InsParameters
		parameterObj.KeyProperty = "class"
		var insertObj = new(InsertTemplate)
		insertObj.Object = object
		insertObj.Parameters = parameterObj

		convertedObj, _ := json.Marshal(insertObj)

		err, _ = common.HTTP_POST("http://localhost:3000/"+request.Controls.Namespace+"/domainClassAttributes/", headers, convertedObj, true)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		maxValue = strconv.Itoa(maxCount + 1)
	}

	return
}

type InsParameters struct {
	KeyProperty string `json:"KeyProperty"`
}
type InsertTemplate struct {
	Object     map[string]interface{}   `json:"Object"`
	Objects    []map[string]interface{} `json:"Objects"`
	Special    Special
	Query      Query
	Parameters InsParameters
}

type Special struct {
	Type       string //SPECIAL
	Extras     map[string]interface{}
	Parameters string
}

type Query struct {
	Type       string //QUERYING, SEARCHING, KEY, ALL
	Parameters string
}
