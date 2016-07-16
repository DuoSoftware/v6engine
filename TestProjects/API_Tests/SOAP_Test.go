package main

import (
	"bytes"
	"fmt"
	"github.com/clbanning/x2j"
	"io/ioutil"
	"net/http"
	"reflect"
)

func main() {
	query := "<s:Envelope xmlns:s=\"http://www.w3.org/2003/05/soap-envelope\" xmlns:a=\"http://www.w3.org/2005/08/addressing\"><s:Header><a:Action s:mustUnderstand=\"1\">http://tempuri.org/IAccount/GetAccountInfoByGuAccountID</a:Action><a:MessageID>urn:uuid:f09d4bdd-8ab5-4eb2-b62c-589b34381b11</a:MessageID><a:ReplyTo><a:Address>http://www.w3.org/2005/08/addressing/anonymous</a:Address></a:ReplyTo><a:To s:mustUnderstand=\"1\">http://192.168.1.194/DuoSubscribe5/V5Services/SMSAccountService/duosubscribermanagement/customer/Account.svc</a:To></s:Header><s:Body><GetAccountInfoByGuAccountID xmlns=\"http://tempuri.org/\"><GUAccountID>2012060605412127</GUAccountID><SecurityToken>d56d3e9cbf7ae07549ef7b1544b423a6</SecurityToken></GetAccountInfoByGuAccountID></s:Body></s:Envelope>"
	GetSoapEnvelope(query)
}

const url = "http://192.168.1.194/DuoSubscribe5/V5Services/SMSAccountService/DuoSubscriberManagement/Customer/Account.svc"

func GetSoapEnvelope(query string) {
	httpClient := new(http.Client)
	resp, err := httpClient.Post(url, "application/soap+xml", bytes.NewBufferString(query))
	if err != nil {
		fmt.Println(err.Error())
	}
	b, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		fmt.Println(e.Error())

	} else {

		data := make(map[string]interface{})

		doc, err4 := x2j.DocToMap(string(b))
		if err4 != nil {
			fmt.Println(err4.Error())
		} else {
			for key, value := range doc["Envelope"].(map[string]interface{}) {
				if key == "Body" {
					for k1, v1 := range value.(map[string]interface{}) {
						if k1 == "GetAccountInfoByGuAccountIDResponse" {
							for k2, v2 := range v1.(map[string]interface{}) {
								if k2 == "GetAccountInfoByGuAccountIDResult" {
									for k3, v3 := range v2.(map[string]interface{}) {
										if reflect.TypeOf(v3).String() == "map[string]interface {}" && v3.(map[string]interface{})["-nil"] != nil {
											data[k3] = ""
										} else {
											data[k3] = v3
										}

									}
								}
							}
						}
					}
				}
			}

		}

		fmt.Println(data)
	}

	resp.Body.Close()
}
