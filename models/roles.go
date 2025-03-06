package models

type Role struct {
	Id    				int
	Name  				string
	CanEditProjects   	bool
	CanEditCategories 	bool
	CanEditUsers      	bool
	CanEditRoles      	bool
	CanEditGenres     	bool
	CanEditAges      	bool
}