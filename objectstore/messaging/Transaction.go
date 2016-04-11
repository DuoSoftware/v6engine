package messaging

type Transaction struct {
	Type       string
	Parameters map[string]interface{}
	Extras     map[string]interface{}
}

type TransactionResponse struct {
	IsSuccess     bool
	Message       string
	TransactionID string
	Extras        map[string]interface{}
}
