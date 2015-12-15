package common

import (
	"strings"
)

func GetSQLTableName(repo string, namespace string, class string) (table string) {
	if repo == "PSQL" {
		table = class
	} else {
		table = "_" + strings.Replace(namespace, ".", "", -1) + "." + class
	}
	return
}
