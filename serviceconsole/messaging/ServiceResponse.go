package messaging

type ServiceResponse struct {
	IsSuccess bool
	Message   string
	Stack     []string
}
