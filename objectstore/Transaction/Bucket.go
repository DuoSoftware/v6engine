package Transaction

import (
	"duov6.com/common"
	"duov6.com/objectstore/messaging"
	"encoding/json"
	"errors"
	"github.com/xuyu/goredis"
	"strings"
	"time"
)

func GetConnection(request *messaging.ObjectRequest) (client *goredis.Redis, err error) {
	client, err = goredis.DialURL("tcp://@" + request.Configuration.ServerConfiguration["REDIS"]["Host"] + ":" + request.Configuration.ServerConfiguration["REDIS"]["Port"] + "/0?timeout=60s&maxidle=60")
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New("Connection to REDIS Failed!")
	}
	return
}

func Create(request *messaging.ObjectRequest) (err error) {
	transactionID := request.Body.Parameters.TransactionID
	transactionStruct := request.Body.Transaction

	if transactionID == "" && transactionStruct.Type != "" {
		client, err := GetConnection(request)
		transactionID = GetTransactionID()
		bucketName := "Transactions." + request.Controls.Namespace + "." + request.Contorls.Class + "." + transactionID

		metadata := make(map[string]interface{})
		metadata["TimeStamp"] = time.Now().Format("2006-01-02 15:04:05")
		metadata["Namespace"] = request.Controls.Namespace
		metadata["Class"] = request.Controls.Class
		metadata["TransactionID"] = transactionID

		bucketValue, _ := json.Marshal(metadata)
		_, err = client.RPush(bucketName, string(bucketValue))
	} else {
		err = errors.New("No Transaction Command Found!")
	}
	return
}

func Add(request *messaging.ObjectRequest) (err error) {
	transactionID := request.Body.Parameters.TransactionID
	transactionStruct := request.Body.Transaction

	if transactionID != "" && transactionStruct.Type == "" {
		bucketName := "Transactions." + request.Controls.Namespace + "." + request.Contorls.Class + "." + transactionID
		bucketValue, _ := json.Marshal(request)
		_, err = client.RPush(bucketName, string(bucketValue))
	} else {
		err = errors.New("No TransactionID Found!")
	}
	return
}

func VerifyBlockList(request *messaging.ObjectRequest) (err error) {
	dbOperation := strings.ToLower(request.Controls.Operation)

	switch dbOperation {
	case "insert":
		break
	case "update":
		break
	case "delete":
		break
	default:
		break
	}

	return err
}

func AddToBlockList(request *messaging.ObjectRequest) (err error) {
	return
}

func DeleteFromBlockList(request *messaging.ObjectRequest) (err error) {
	return
}

func GetTransactionID() string {
	return common.GetGUID()
}
