package messaging

type BulkHeader struct {
	Source string `json:"src"`
	Dest string `json:"dest"`
	Details []BulkDetails `json:"details"`
}

type BulkDetails struct {
	Class string `json:"class"`
	Type string `json:"type"`
	Params map[string]interface{} `json:"params"`
}
