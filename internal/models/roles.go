// internal/models/role.go

package models

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type UserRole struct {
	UserID int `json:"user_id"`
	RoleID int `json:"role_id"`
}
