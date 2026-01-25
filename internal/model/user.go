package model

type User struct {
	ID       int
	FullName string
	Phone    string
	Role     string // client/admin
}
