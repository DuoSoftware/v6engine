package security

func ValidateSecurity(value string) (status bool) {
	status = CheckForSQLInjection(value)
	return
}
