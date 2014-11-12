package configuration

type StoreConfiguration struct {
	StoreId             string
	StorageEngine       string
	ServerConfiguration map[string]map[string]string
	StoreConfiguration  map[string]map[string]string
}
