package domain

import "time"

type UserRole string

const (
	RoleClient        UserRole = "client"
	RoleSalonAdmin    UserRole = "salon_admin"
	RoleMaster        UserRole = "master"
	RolePlatformAdmin UserRole = "platform_admin"
)

type User struct {
	ID           string    `json:"id"`
	Role         UserRole  `json:"role"`
	FullName     string    `json:"fullName,omitempty"`
	Email        string    `json:"email,omitempty"`
	Phone        string    `json:"phone,omitempty"`
	PasswordHash string    `json:"-"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
}
