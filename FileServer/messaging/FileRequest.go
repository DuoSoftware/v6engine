package messaging

import (
	"net/http"
)

type FileRequest struct {
	//Use when not using an REST interface
	FileName     string
	FilePath     string //Relative path
	Body         []byte
	RootSavePath string
	RootGetPath  string
	//use when using an interface
	WebResponse http.ResponseWriter
	WebRequest  *http.Request
	//common
	Parameters map[string]string //id, namespace, class = preferrebly form martini
}
