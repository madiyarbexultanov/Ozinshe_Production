package models

import "time"

type User struct {
	Id           int
	Email        string
	PasswordHash string
	Name         string
	Phone        string
	Birthday     time.Time
	RoleID       int
}