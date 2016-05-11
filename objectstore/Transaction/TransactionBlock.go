package Transaction

import (
	"duov6.com/objectstore/cache"
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"errors"
)

func AppendBlockEntry(request *messaging.ObjectRequest, TransactionID string) (err error) {
	entry := GetBlockEntryName(request, TransactionID)
	if cache.ExistsKeyValue(request, entry) {
		var data []interface{}
		byteValue := cache.GetKeyValue(request, entry)
		err = json.Unmarshal(byteValue, &data)
		data = append(data, request)
		byteValue = nil
		byteValue, err = json.Marshal(data)
		err = cache.StoreKeyValue(request, entry, string(byteValue))
	} else {
		var data []interface{}
		data = append(data, request)
		byteValue, _ := json.Marshal(data)
		err = cache.StoreKeyValue(request, entry, string(byteValue))
	}
	return
}

func DeleteBlockEntry(request *messaging.ObjectRequest, TransactionID string) (err error) {
	entry := GetBlockEntryName(request, TransactionID)
	status := cache.DeleteKey(request, entry)
	if !status {
		err = errors.New("Delete Failed!")
	}
	return
}

func VerifyBlockSafe(request *messaging.ObjectRequest) (status bool) {
	// 	TransactionID := request.Body.Parameters.TransactionID
	// 	entry := GetBlockEntryName(request, TransactionID)
	// 	if cache.ExistsKeyValue(request, entry) {
	// 		var data []interface{}
	// 		byteValue := cache.GetValue(request, entry)
	// 		err := json.Unmarshal(byteValue, &data)

	// 		var objects []map[string]interface{}
	// 		if request.Body.Object != nil {
	// 			objects = make([]map[string]interface{}, 1)
	// 			objects[0] = request.Body.Object
	// 		} else {
	// 			objects = make([]map[string]interface{}, len(request.Body.Objects))
	// 			objects = request.Body.Objects
	// 		}

	// 		for _, singularObject := range objects {
	// 			for _, singleData := range data {
	// 				if singleData.Body.Object != nil {
	// 					if (singularObject[request.Body.Parameters.KeyProperty].(string) == singleData[request.Body.Parameters.KeyProperty].(string)) && (TransactionID != singleData.Body.Parameters.TransactionID) {
	// 						status = false
	// 						return
	// 					}
	// 				} else {
	// 					for _, subObjects := range singleData.Body.Objects {
	// 						if (singularObject[request.Body.Parameters.KeyProperty].(string) == subObjects[request.Body.Parameters.KeyProperty].(string)) && (TransactionID != singleData.Body.Parameters.TransactionID) {
	// 							status = false
	// 							return
	// 						}
	// 					}
	// 				}
	// 			}
	// 		}

	// 	} else {
	// 		status = false
	// 	}
	status = true
	return
}
