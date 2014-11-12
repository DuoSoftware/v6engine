package messaging

type RequestBody struct {
	Parameters ObjectParameters
	Query      Query
	Object     map[string]interface{}
	Objects    []map[string]interface{}
}
