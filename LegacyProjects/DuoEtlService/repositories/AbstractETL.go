package repositories

import (
	"duov6.com/DuoEtlService/messaging"
)

type AbstractETL interface {
	GetETLName() string
	ExecuteETLService(request *messaging.ETLRequest) messaging.ETLResponse
}
