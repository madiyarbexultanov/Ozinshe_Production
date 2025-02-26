package models

type User struct {
	Id				int
	Email			string
	PasswordHash	string
	Name			string
	Role			string
}