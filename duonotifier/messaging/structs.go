package messaging

type TemplateRequest struct {
	TemplateID    string
	DefaultParams map[string]string
	CustomParams  map[string]string
	Namespace     string
}
