package Transaction

import (
	"duov6.com/objectstore/cache"
	"duov6.com/objectstore/messaging"
	//"duov6.com/objectstore/repositories"
	//"duov6.com/objectstore/storageengines"
	"errors"
)

func Execute(request *messaging.ObjectRequest) (err error) {
	//Get Length of List
	TransactionID := request.Body.Transaction.Parameters["TransactionID"].(string)
	if cache.ExistsKeyValue(request, GetBucketName(TransactionID)) && cache.GetListLength(request, GetBucketName(TransactionID)) > 1 {
		err = StartProcess(request)
	} else {
		err = errors.New("Transaction either already Rolledback or no transaction items found!")
	}
	return
}

func StartProcess(request *messaging.ObjectRequest) (err error) {
	//GetTask
	//ExecuteTask
	return
}

func GetTask(request *messaging.ObjectRequest) (retRequest *messaging.ObjectRequest) {

	return
}

func PushToSuccessList(request *messaging.ObjectRequest) {

}

func PushToInvertList(request *messaging.ObjectRequest) {

}
