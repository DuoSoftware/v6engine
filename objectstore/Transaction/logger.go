package Transaction

import (
	"bytes"
	"duov6.com/objectstore/cache"
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func TLog(request *messaging.ObjectRequest, TransactionID string) {
	//Get All Records
	TID := GetBucketName(TransactionID)
	results, err := cache.LRange(request, TID, cache.Transaction, 1, 999)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		requestArray := make([]messaging.ObjectRequest, len(results))
		objectArray := make([]map[string]interface{}, len(results))
		for x := 0; x < len(results); x++ {
			reqObject := messaging.ObjectRequest{}
			reqPointer := &reqObject
			err := json.Unmarshal([]byte(results[0]), &reqPointer)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				requestArray[x] = *reqPointer
			}
		}

		for x := 0; x < len(requestArray); x++ {
			singleMap := make(map[string]interface{})
			singleMap["ID"] = "-888"
			singleMap["StepNo"] = strconv.Itoa(x)
			singleMap["Request"] = requestArray[x]
			singleMap["Status"] = "pending"
			singleMap["TimeStamp"] = GetTimeStamp()
			singleMap["TransactionID"] = TransactionID
			objectArray[x] = singleMap
		}

		var parameterObj InsParameters
		parameterObj.KeyProperty = "ID"
		var insertObj = new(InsertTemplate)
		insertObj.Objects = objectArray
		insertObj.Parameters = parameterObj

		convertedObj, err := json.Marshal(insertObj)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			Post("localhost", "logs", "TransactionLogs", convertedObj)
		}
	}
}

func GetTimeStamp() (timeString string) {
	timeString = time.Now().Format("2006-01-02 15:04:05")
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

func Post(domain, namespace, class string, JSON_DATA []byte) (err error, body []byte) {

	securityToken := "ignore"
	url := "http://" + domain + ":3000/" + namespace + "/" + class + "?securityToken=" + securityToken

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(JSON_DATA))
	// req.Header.Set("securityToken", securityToken)
	// req.Header.Set("log", log)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err = errors.New("Connection Failed!")
	} else {
		body, _ = ioutil.ReadAll(resp.Body)
		if resp.StatusCode != 200 {
			err = errors.New(string(body))
		}
	}
	defer resp.Body.Close()
	return
}

func UpdateLogStatus(StepNo int, TransactionID string, value string) {
	//get matching records
	QueryObject := make(map[string]interface{})
	QueryStruct := Query{}
	QueryStruct.Type = "Query"
	QueryStruct.Parameters = "SELECT * from TransactionLogs WHERE StepNo='" + strconv.Itoa(StepNo) + "' AND TransactionID='" + TransactionID + "';"
	QueryObject["Query"] = QueryStruct
	queryBody, _ := json.Marshal(QueryObject)
	err, byteBody := Post("localhost", "logs", "TransactionLogs", queryBody)
	if err != nil {
		fmt.Println("Error : " + err.Error())
	} else {
		objectArray := make([]map[string]interface{}, 0)
		_ = json.Unmarshal(byteBody, &objectArray)

		for x := 0; x < len(objectArray); x++ {
			objectArray[x]["Status"] = value
		}

		var parameterObj InsParameters
		parameterObj.KeyProperty = "ID"
		var insertObj = new(InsertTemplate)
		insertObj.Objects = objectArray
		insertObj.Parameters = parameterObj

		convertedObj, err := json.Marshal(insertObj)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			Post("localhost", "logs", "TransactionLogs", convertedObj)
		}

	}
}
