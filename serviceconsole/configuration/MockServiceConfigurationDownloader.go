package configuration

type MockServiceConfigurationDownloader struct {
}

func (c MockServiceConfigurationDownloader) DownloadConfiguration() StoreServiceConfiguration {
	config := StoreServiceConfiguration{}

	config.ServerConfiguration = make(map[string]map[string]string)
	config.PublisherConfiguration = make(map[string]map[string]RoutingKeys)

	//SERVER CONFIGURATIONS

	var DuoV6ServiceServer map[string]string
	DuoV6ServiceServer = make(map[string]string)
	DuoV6ServiceServer["Host"] = "192.168.1.194"
	DuoV6ServiceServer["Port"] = "5672"
	DuoV6ServiceServer["UserName"] = "admin"
	DuoV6ServiceServer["Password"] = "admin"
	config.ServerConfiguration["DuoV6ServiceServer"] = DuoV6ServiceServer

	var LocalTestServer map[string]string
	LocalTestServer = make(map[string]string)
	LocalTestServer["Host"] = "192.168.1.194"
	LocalTestServer["Port"] = "5672"
	LocalTestServer["UserName"] = "admin"
	LocalTestServer["Password"] = "admin"
	config.ServerConfiguration["LocalTestServer"] = LocalTestServer

	//PUBLISHER CONFIGURATIONS

	var publisher_01 map[string]RoutingKeys
	publisher_01 = make(map[string]RoutingKeys)
	publisher_01["Excel"] = RoutingKeys{}
	publisher_01["Image"] = RoutingKeys{}
	publisher_01["Queued"] = RoutingKeys{}
	//Specify branch based routing keys here
	var routeKeyMap map[string]string
	routeKeyMap = make(map[string]string)
	routeKeyMap["1"] = "Disconnect"
	routeKeyMap["2"] = "Reconnect"
	publisher_01["WorkFlow"] = RoutingKeys{routeKeyMap}
	config.PublisherConfiguration["publisher_01"] = publisher_01

	return config
}
