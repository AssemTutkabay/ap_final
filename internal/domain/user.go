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
	ID        string    `json:"id"`
	Role      UserRole  `json:"role"`
	Contact   string    `json:"contact"` // phone/email одним полем
	CreatedAt time.Time `json:"createdAt"`
}
