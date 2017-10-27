package authlib

import (
//"duov6.com/objectstore/client"
//"duov6.com/term"
//"encoding/json"
)

type UserHandler struct {
}

type UserRole struct {
	RoleID        string
	Name          string
	Description   string
	NumberofUsers int
}

type RoleMinimum struct {
	RoleID string
	Name   string
}

type UsersRoles struct {
	UserID string
	Roles  []RoleMinimum
}

type RolesForUsers struct {
	RolePageID string
	Users      []string
}

type UserPage struct {
	USerID      string
	RolepPageID []string
}
