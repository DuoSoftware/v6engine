package authlib

type AppCertificate struct {
	AuthKey       string
	UserID        string
	ApplicationID string
	AppSecretKey  string
	Otherdata     map[string]interface{}
}
