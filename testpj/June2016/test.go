package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

func Invoke(FlowData map[string]interface{}) {

	// getting the details from the input arguments
	inputData := make(map[string]interface{})
	inputData["customer"] = FlowData["customer"].(string)
	inputData["guCustomerID"] = FlowData["guCustomerID"].(string)
	inputData["guAccountID"] = FlowData["guAccountID"].(string)
	inputData["paymentDate"] = FlowData["paymentDate"].(string)
	inputData["paymentMethod"] = FlowData["paymentMethod"].(string)
	inputData["amount"] = FlowData["amount"].(string)
	inputData["bankCharges"] = FlowData["bankCharges"].(string)
	inputData["note"] = FlowData["note"].(string)
	inputData["guTranID"] = FlowData["guTranID"].(string)
	inputData["status"] = FlowData["status"].(string)
	inputData["createdUser"] = FlowData["createdUser"].(string)
	inputData["createdDate"] = FlowData["createdDate"].(string)

	JSON_DOC, _ := json.Marshal(inputData)

	url := FlowData["URL"].(string)
	securityToken := FlowData["securityToken"].(string)

	err, message := SendRequest(JSON_DOC, url, securityToken)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(message)
}

// global method used to create a JSON from the map provided.
func SendRequest(JSON []byte, url string, token string) (err error, response string) {
	fmt.Println(token)
	fmt.Println(string(JSON))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(JSON))
	req.Header.Set("securityToken", token)
	client := &http.Client{}
	resp, err := client.Do(req)

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		err = errors.New(string(body))
		return
	}
	fmt.Println(resp.Status)
	fmt.Println(body)
	response = string(body)
	defer resp.Body.Close()
	return
}
func main() {
	inputData := make(map[string]interface{})
	inputData["customer"] = "dd"
	inputData["guCustomerID"] = "1"
	inputData["guAccountID"] = "1"
	inputData["paymentDate"] = "1"
	inputData["paymentMethod"] = "1"
	inputData["amount"] = "1"
	inputData["bankCharges"] = "1"
	inputData["note"] = "1"
	inputData["guTranID"] = "1"
	inputData["status"] = "1"
	inputData["createdUser"] = "1"
	inputData["createdDate"] = "1"
	inputData["URL"] = "http://cloudcharge.com/services/duosoftware.payment.service/payment/makePayment"
	inputData["securityToken"] = "78d2a5c15ea3254f273e437f49f2f3c9"

	Invoke(inputData)
}
