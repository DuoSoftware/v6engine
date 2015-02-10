package processmanager

import (
	"duov6.com/serviceconsole/messaging"
	//"fmt"
	"log"
)

type ImageWorker struct {
}

func (worker ImageWorker) GetWorkerName() string {
	return "ImageWorker"
}

func (worker ImageWorker) ExecuteWorker(request *messaging.ServiceRequest) messaging.ServiceResponse {
	//fmt.Println("Not Implemented in ImageWorker")

	var temp = messaging.ServiceResponse{}
	if request.Body != nil {
		log.Printf("Received a message: %s", request.Body)
		temp.IsSuccess = true
	} else {
		temp.IsSuccess = false
	}

	return temp
}
