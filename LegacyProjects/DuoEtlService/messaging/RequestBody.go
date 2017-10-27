package messaging

type RequestBody struct {
	Parameters ObjectParameters
	Query      Query
	Special    Special
	Object     map[string]interface{}
	Objects    []map[string]interface{}
}
