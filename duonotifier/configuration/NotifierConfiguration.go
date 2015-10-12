package configuration

type NotifierConfiguration struct {
	NotifyId            string
	NotifyMethodsConfig map[string]map[string]string
}
