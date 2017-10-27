package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"duov6.com/gorest"
	"duov6.com/objectstore/messaging"
	"duov6.com/objectstore/client"
)


type BulkTransferService struct {
	gorest.RestService
	transfer  gorest.EndPoint `method:"POST" path:"/transfer" postdata:"BulkHeader"`
}


func (h *BulkTransferService) Start() {
	gorest.RegisterService(h)

	err := http.ListenAndServe(":3001", gorest.Handle())
	if err != nil {
		fmt.Println(err.Error())
		return
	}else{
		fmt.Println ("Bulk service started in port 3001")
	}
}

func (A BulkTransferService) Transfer(c messaging.BulkHeader) {
	for _, detail := range (c.Details){
		switch (detail.Type){
			case "id":
				bytes, err := client.Go("", c.Source, detail.Class).GetOne().ByUniqueKey(detail.Params["key"].(string)).Ok();
				isErr, m := A.getMap(bytes, err)
				if (!isErr){
					keyField := detail.Params["keyField"].(string)
					client.Go("", c.Dest, detail.Class).StoreObject().WithKeyField(keyField).AndStoreOne(m).Ok();
				}else{
					fmt.Println("K OBJECT NOT FOUND!!!!")
				}
				break
			case "filter":
				bytes, err := client.Go("", c.Source, detail.Class).GetMany().ByQuerying(detail.Params["filter"].(string)).Ok();
				isErr, m := A.getMapArray(bytes, err)
				if (!isErr){
					keyField := detail.Params["keyField"].(string)
					client.Go("", c.Dest, detail.Class).StoreObject().WithKeyField(keyField).AndStoreMany(m).Ok();
				}else{
					fmt.Println("F OBJECT NOT FOUND!!!!")
				}
				break
			case "search":
				bytes, err := client.Go("", c.Source, detail.Class).GetMany().BySearching(detail.Params["filter"].(string)).Ok();
				isErr, m := A.getMapArray(bytes, err)
				if (!isErr){
					keyField := detail.Params["keyField"].(string)
					client.Go("", c.Dest, detail.Class).StoreObject().WithKeyField(keyField).AndStoreMany(m).Ok();
				}else{
					fmt.Println("S OBJECT NOT FOUND!!!!")
				}
				break
			case "all":
				break
			default:
				fmt.Println ("UNKNOWN OPERATION!!!!")
			 	break
		}
	}

	A.ResponseBuilder().SetResponseCode(200).WriteAndOveride([]byte("{}"))
	return
}

func (A *BulkTransferService) getMap(bytes []byte, err string) (isError bool, outData map[string]interface{}) {
	outData = make(map[string]interface{})
	isError = true			

	if err == "" {
		if bytes != nil {
			err := json.Unmarshal(bytes, &outData)
			if err == nil {
				isError = false
			}
		}
	}

	return
}


func (A *BulkTransferService) getMapArray(bytes []byte, err string) (isError bool, outData []interface{}) {
	outData = make([]interface{},0)
	isError = true			

	if err == "" {
		if bytes != nil {
			err := json.Unmarshal(bytes, &outData)
			if err == nil {
				isError = false
			}
		}
	}

	return
}

func (h *BulkTransferService) Stop() {
}


