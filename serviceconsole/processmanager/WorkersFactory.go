package processmanager

import (
	"fmt"
)

func Create(code string) AbstractWorkers {

	fmt.Println("Excuting AbstractWorker : " + code)

	var worker AbstractWorkers
	switch code {
	case "Excel":
		worker = ExcelWorker{}
	case "Image":
		worker = ImageWorker{}
	case "WorkFlow.Disconnect":
		worker = WorkFlowWorker{}
	case "WorkFlow.Reconnect":
		worker = WorkFlowWorker2{}
	case "Queued":
		worker = QueuedObjectStoreWorker{}
	}
	return worker
}
