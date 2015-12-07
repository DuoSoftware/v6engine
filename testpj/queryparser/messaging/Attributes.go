package messaging

type Attributes struct {
	Select  map[string]string
	From    string
	Where   map[string]string
	Orderby map[string]string
}
