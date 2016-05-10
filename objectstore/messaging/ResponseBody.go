package messaging

type ResponseBody struct {
	IsSuccess   bool
	Message     string
	Stack       []string
	Data        []map[string]interface{}
	Transaction TransactionResponse
}
