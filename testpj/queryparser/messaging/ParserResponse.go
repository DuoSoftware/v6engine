package messaging

type ParserResponse struct {
	IsSuccess  bool
	Body       map[string]string
	Message    string
	QueryItems Attributes
}
