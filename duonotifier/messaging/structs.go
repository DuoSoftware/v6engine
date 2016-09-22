package messaging

type TemplateRequest struct {
	TemplateID    string
	DefaultParams map[string]string
	CustomParams  map[string]string
	Namespace     string
}

type NotifierResponse struct {
	IsSuccess bool
	Message   string
}
