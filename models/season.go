package models

type Season struct {
	Id      	int
	Number  	int			// Номер сезона
	MovieID 	int			// ID фильма, к которому относится сезон
	Episodes 	[]Episode	// Связь один ко многим
}
