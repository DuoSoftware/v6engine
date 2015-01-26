package api

type Doc struct {
	Name        string
	Description string
	Methods     []Method
}

type Method struct {
	Name         string
	Description  string
	Method       string
	URI          string
	OutPutBody   string
	OutPutType   string
	InParameters []Parameters
}

type Parameters struct {
	Name        string
	Description string
	Type        string
	ParamBody   string
}

type ApiHandler struct {
	Document Doc
}

func (api *ApiHandler) NewDoc(Name, Description string) {
	api.Document = Doc{}
	api.Document.Name = Name
	api.Document.Description = Description
	var tmp []Method
	tmp = make([]Method, 0)
	api.Document.Methods = tmp
}

func (api *ApiHandler) AddMethod(m Method) {
	//appendMethod(api.Document.Methods, m)
	api.Document.Methods = []Method{m}
}

func appendMethod(slice []Method, data ...Method) []Method {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) {
		newSlice := make([]Method, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

func (api *ApiHandler) Save() {

}
