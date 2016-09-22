package connmanager

var connections map[string]interface{}

func getConnection(db string, ns string)(connList map[string]interface{}){
	if connections == nil{
		connections = make(map[string]interface{})
	}

	if connections[db] == nil{
		connections[db] = make(map[string]interface{})
	}

	nsConnections := connections[db].(map[string]interface{})
/*
	if nsConnections[ns] == nil{
		nsConnections[ns] = make(map[string]interface{})
	}
*/
	connList = nsConnections//[ns].(map[string]interface{})
	return
}

func Get(db string, ns string) (conn interface{}){
	connList := getConnection(db,ns)
	conn = connList[ns]
	return 
}

func Set(db string,ns string,conn interface{}){
	connList := getConnection(db,ns)
	connList[ns] = conn
}
