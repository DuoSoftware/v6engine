package fws

func GlobalConfigRecieved(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
	SetGlobalConfig(data["class"].(string), data["config"].([]interface{}))
}
