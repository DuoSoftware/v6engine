package common

import (
	"strings"
)

func GetSQLTableName(repo string, namespace string, class string) (table string) {
	if repo == "PSQL" {
		table = class
	} else if repo == "CDS" {
		table = class
	} else {
		table = "_" + strings.ToLower(strings.Replace(namespace, ".", "", -1)) + "." + class
	}
	return
}
