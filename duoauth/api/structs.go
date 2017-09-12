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
	Scopes       []string
	Tenants      []UserTenant
	Avatar       string
}

type UserTenant struct {
	IsDefault bool
	IsAdmin   bool
	TenantID  string
}

type Tenant struct {
	ObjectID string
	TenantID string
	Admin    string
	Country  string
	Type     string
}

type UserCreateInfo struct {
	Name       string
	Email      string
	Country    string
	Password   string
	TenantID   string
	TenantType string
	AvatarUrl  string
}
