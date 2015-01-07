package common

type ScheduleObject struct {
	Timestamp     string
	OperationData map[string]interface{}
	ControlData   map[string]interface{}
}
