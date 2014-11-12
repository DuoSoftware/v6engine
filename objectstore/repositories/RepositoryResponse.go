package repositories

type RepositoryResponse struct {
	ResponseJson  string
	IsSuccess     bool
	IsImplemented bool
	Message       string
	Body          []byte
}

func (r *RepositoryResponse) GetErrorResponse(errorMessage string) {
	r.IsSuccess = false
	r.IsImplemented = true
	r.Message = errorMessage
}

func (r *RepositoryResponse) GetResponseWithBody(body []byte) {
	r.IsSuccess = true
	r.IsImplemented = true
	r.Message = "Operation Success!!!"
	r.Body = body
}
