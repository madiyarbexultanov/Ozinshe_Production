package models

type Movie struct {
	Id			int
	Title		string
	Categories	[]Category
	Genres		[]Genre
	Ages		[]Ages
	ReleaseYear	int
	Runtime		int
	KeyWords	[]string
	Description	string
	Director	string
	Producer	string
	Seasons     []Season    
	MovieTypeId int
	MovieType	string
	Media       MovieMedia
}

type Moviesfilters struct {
	GenreIds	string
	CategoryIds	string
	TypeIds		string
	AgeIds		string
}