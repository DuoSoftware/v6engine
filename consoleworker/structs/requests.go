package structs

type ServiceRequest struct {
	RefID              string
	RefType            string
	OperationCode      string
	TimeStamp          string
	ControlParameters  map[string]interface{}
	Parameters         map[string]interface{}
	ScheduleParameters map[string]interface{}
	Body               []byte
}

type ServiceResponse struct {
	Err error
}

type QueueDispatchRequest struct {
	Objects []ServiceRequest
}
