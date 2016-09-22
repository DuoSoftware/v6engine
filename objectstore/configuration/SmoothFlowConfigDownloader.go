package configuration

import (
	"strings"
)

type SmoothFlowConfigDownloader struct {
}

func (c SmoothFlowConfigDownloader) DownloadConfiguration(securityToken string, namespace string, class string, parameters map[string]interface{}) StoreConfiguration {
	config := StoreConfiguration{}
	config.StoreId = "Default"
	config.StorageEngine = "REPLICATED"
	config.ServerConfiguration = make(map[string]map[string]string)
	config.StoreConfiguration = make(map[string]map[string]string)
	config.StoreConfiguration = getStoreConfigs(parameters["DB_Type"].(string))
	config.ServerConfiguration = getServerConfigs(parameters)
	return config
}

func getStoreConfigs(db string) map[string]map[string]string {
	config := make(map[string]map[string]string)

	if strings.EqualFold(db, "mysql") || strings.EqualFold(db, "cloudsql") {
		db = "CLOUDSQL"
	}

	setting := make(map[string]string)
	setting["priority1"] = db

	config["GET-ALL"] = setting
	config["GET-SEARCH"] = setting
	config["GET-KEY"] = setting
	config["GET-QUERY"] = setting
	config["INSERT-MULTIPLE"] = setting
	config["INSERT-SINGLE"] = setting
	config["UPDATE-MULTIPLE"] = setting
	config["UPDATE-SINGLE"] = setting
	config["DELETE-SINGLE"] = setting
	config["DELETE-MULTIPLE"] = setting
	config["SPECIAL"] = setting

	return config
}

func getServerConfigs(params map[string]interface{}) map[string]map[string]string {
	config := make(map[string]map[string]string)

	db_type := params["DB_Type"].(string)

	if strings.EqualFold(db_type, "mysql") || strings.EqualFold(db_type, "cloudsql") {
		settings := make(map[string]string)
		settings["Username"] = params["Username"].(string)
		settings["Password"] = params["Password"].(string)
		settings["Url"] = params["Url"].(string)
		settings["Port"] = params["Port"].(string)
		config["MYSQL"] = settings
	} else if strings.EqualFold(db_type, "CASSANDRA") {
		settings := make(map[string]string)
		settings["Url"] = params["Url"].(string)
		config["CASSANDRA"] = settings
	} else if strings.EqualFold(db_type, "ELASTIC") {
		settings := make(map[string]string)
		settings["Host"] = params["Host"].(string)
		settings["Port"] = params["Port"].(string)
		config["ELASTIC"] = settings
	} else if strings.EqualFold(db_type, "COUCH") {
		settings := make(map[string]string)
		settings["Url"] = params["Url"].(string)
		settings["Bucket"] = params["Bucket"].(string)
		config["COUCH"] = settings
	} else if strings.EqualFold(db_type, "GoogleBigTable") {
		settings := make(map[string]string)
		settings["type"] = params["type"].(string)
		settings["private_key_id"] = params["private_key_id"].(string)
		settings["private_key"] = params["private_key"].(string)
		settings["client_email"] = params["client_email"].(string)
		settings["client_id"] = params["client_id"].(string)
		settings["auth_uri"] = params["auth_uri"].(string)
		settings["token_uri"] = params["token_uri"].(string)
		settings["auth_provider_x509_cert_url"] = params["auth_provider_x509_cert_url"].(string)
		settings["client_x509_cert_url"] = params["client_x509_cert_url"].(string)
		config["GoogleBigTable"] = settings
	} else if strings.EqualFold(db_type, "GoogleDataStore") {
		settings := make(map[string]string)
		settings["type"] = params["type"].(string)
		settings["private_key_id"] = params["private_key_id"].(string)
		settings["private_key"] = params["private_key"].(string)
		settings["client_email"] = params["client_email"].(string)
		settings["client_id"] = params["client_id"].(string)
		settings["auth_uri"] = params["auth_uri"].(string)
		settings["token_uri"] = params["token_uri"].(string)
		settings["auth_provider_x509_cert_url"] = params["auth_provider_x509_cert_url"].(string)
		settings["client_x509_cert_url"] = params["client_x509_cert_url"].(string)
		config["GoogleDataStore"] = settings
	} else if strings.EqualFold(db_type, "HIVE") {
		settings := make(map[string]string)
		settings["Host"] = params["Host"].(string)
		settings["Port"] = params["Port"].(string)
		config["HIVE"] = settings
	} else if strings.EqualFold(db_type, "MSSQL") {
		settings := make(map[string]string)
		settings["Username"] = params["Username"].(string)
		settings["Password"] = params["Password"].(string)
		settings["Server"] = params["Server"].(string)
		settings["Port"] = params["Port"].(string)
		config["MSSQL"] = settings
	} else if strings.EqualFold(db_type, "MONGO") {
		settings := make(map[string]string)
		settings["Url"] = params["Url"].(string)
		config["MONGO"] = settings
	} else if strings.EqualFold(db_type, "POSTGRES") {
		settings := make(map[string]string)
		settings["Username"] = params["Username"].(string)
		settings["Password"] = params["Password"].(string)
		settings["Url"] = params["Url"].(string)
		settings["Port"] = params["Port"].(string)
		config["POSTGRES"] = settings
	} else if strings.EqualFold(db_type, "REDIS") {
		settings := make(map[string]string)
		settings["Host"] = params["Host"].(string)
		settings["Port"] = params["Port"].(string)
		config["REDIS"] = settings
	}

	return config
}
