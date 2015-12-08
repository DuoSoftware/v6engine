package queryparser

type queryObject struct {
	operation      string             //select, show, describe, explain
	selectedFields []string           //["Name", "Age"]
	table          string             //student
	where          []map[int][]string // [0]map[0]["age", ">=", "24"]
	orderby        map[string]string  //map["Name"] = "ASC"
}
