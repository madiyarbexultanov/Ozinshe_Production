package models

type Season struct {
	ID      	int
	Number  	int			// Номер сезона
	MovieID 	int			// ID фильма, к которому относится сезон
	Episodes 	[]Episode	// Связь один ко многим
}
