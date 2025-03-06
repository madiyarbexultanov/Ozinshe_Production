package models

type Movie struct {
	Title		string
	Category	[]Category
	GenreID		[]Genre
	Ages		[]Ages
	ReleaseYear	int
	Runtime		int
	KeyWords	[]string
	Description	string
	Director	string
	Producer	string
	Seasons     []Season    
	Cover       string
	Screenshots []string
}