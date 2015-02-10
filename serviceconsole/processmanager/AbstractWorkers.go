package processmanager

import (
	"duov6.com/serviceconsole/messaging"
)

type AbstractWorkers interface {
	GetWorkerName() string
	ExecuteWorker(request *messaging.ServiceRequest) messaging.ServiceResponse
}
