package api

import ()

type AuthResponse struct {
	Status  bool
	Message string
	Data    interface{}
}

type User struct {
	ObjectID     string
	EmailAddress string
	Name         string
	Country      string
	Tenants      []UserTenant
}

type UserTenant struct {
	IsDefault string
	IsAdmin   string
	TenantID  string
}

type Tenant struct {
	TenantID   string
	TenantName string
	Owner      string
	Location   string
	Type       string
}

type UserCreateInfo struct {
	Name     string
	Email    string
	Country  string
	Password string
	TenantID string
}
