package messaging

type FileResponse struct {
	IsSuccess bool
	FilePath  string //Absolute path
	Message   string
	body      []byte
}
