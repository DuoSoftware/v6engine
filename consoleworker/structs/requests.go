package structs

type ServiceRequest struct {
	RefId              string
	RefType            string
	OperationCode      string
	TimeStamp          string
	TimeStampReadable  string
	ControlParameters  map[string]interface{}
	Parameters         map[string]interface{}
	ScheduleParameters map[string]interface{}
	Body               interface{} //previously []byte
}

type ServiceResponse struct {
	Err error
}

type QueueDispatchRequest struct {
	Objects []ServiceRequest
}
