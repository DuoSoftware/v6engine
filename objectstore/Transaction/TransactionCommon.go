package Transaction

import (
	"duov6.com/common"
	"duov6.com/objectstore/messaging"
)

func GetTransactionID() string {
	return common.GetGUID()
}

func GetBucketName(id string) (name string) {
	name = "Transactions." + id
	return
}

func GetBlockEntryName(request *messaging.ObjectRequest, TransactionID string) (name string) {
	name = "TransactionBlockEntry." + request.Controls.Namespace + "." + request.Controls.Class + "." + ".{" + TransactionID + "}"
	return
}

func GetSuccessBucketName(id string) (name string) {
	name = "Transactions.SuccessList." + id
	return
}

func GetInvertBucketName(id string) (name string) {
	name = "Transactions.InvertList." + id
	return
}
