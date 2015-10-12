package configuration

type ETLConfiguration struct {
	DataPath  string
	EtlConfig map[string]map[string]string
}
