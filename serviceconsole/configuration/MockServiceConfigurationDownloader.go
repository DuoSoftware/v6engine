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
	DuoV6ServiceServer["Host"] = "localhost"
	DuoV6ServiceServer["Port"] = "5672"
	DuoV6ServiceServer["UserName"] = "guest"
	DuoV6ServiceServer["Password"] = "guest"
	config.ServerConfiguration["DuoV6ServiceServer"] = DuoV6ServiceServer

	var WindowsFileServer map[string]string
	WindowsFileServer = make(map[string]string)
	WindowsFileServer["SavePath"] = "D:/FileServer/"
	WindowsFileServer["GetPath"] = "ftp://127.0.0.1/"
	config.ServerConfiguration["WindowsFileServer"] = WindowsFileServer

	var LinuxFileServer map[string]string
	LinuxFileServer = make(map[string]string)
	LinuxFileServer["SavePath"] = "/FileServer/"
	LinuxFileServer["GetPath"] = "ftp://127.0.0.1/"
	config.ServerConfiguration["LinuxFileServer"] = LinuxFileServer

	//PUBLISHER CONFIGURATIONS

	var publisher_01 map[string]RoutingKeys
	publisher_01 = make(map[string]RoutingKeys)
	publisher_01["Excel"] = RoutingKeys{}
	publisher_01["Image"] = RoutingKeys{}
	publisher_01["Queued"] = RoutingKeys{}
	//Specify branch based routing keys here
	/*var routeKeyMap map[string]string
	routeKeyMap = make(map[string]string)
	routeKeyMap["1"] = "Disconnect"
	routeKeyMap["2"] = "Reconnect"
	publisher_01["WorkFlow"] = RoutingKeys{routeKeyMap}*/
	config.PublisherConfiguration["publisher_01"] = publisher_01

	return config
}
