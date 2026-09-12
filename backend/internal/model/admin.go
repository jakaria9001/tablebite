package model

type Admin struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         string
	IsActive     bool
}
