package messaging

type ParserRequest struct {
	Body       map[string]string
	Query      string
	QueryItems Attributes
}
