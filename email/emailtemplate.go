package email

type Emailtemplate struct {
	Id         string
	Subject    string
	Body       string
	Signature  string
	Parameters map[int]string
}
