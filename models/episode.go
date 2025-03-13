package models

type Episode struct {
	Id        int
	Number    int    // Номер серии
	SeasonID  int    // ID сезона, к которому относится эпизод
	VideoURL  string // Ссылка на видео
}