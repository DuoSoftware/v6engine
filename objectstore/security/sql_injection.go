package security

import (
	"strings"
)

func CheckForSQLInjection(value string) (status bool) {
	if strings.Contains(value, "'") && strings.Contains(value, "=") {
		status = true
	}
	return
}
