package models

type Role struct {
	ID          int    			`json:"id"`
	Name        string 			`json:"name"`
	Permissions []Permission 	`json:"permissions"`
}

type Permission struct {
	ID     int    			`json:"id"`
	Name   string 			`json:"name"`
	Action string 			`json:"action"`
}