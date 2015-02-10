package configuration

type StoreServiceConfiguration struct {
	ServerConfiguration    map[string]map[string]string
	PublisherConfiguration map[string]map[string]RoutingKeys
}
