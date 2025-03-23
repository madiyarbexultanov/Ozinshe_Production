package repositories

import (
	"context"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WatchlistRepository struct {
	db *pgxpool.Pool
}

func NewWatchlistRepository(conn *pgxpool.Pool) *WatchlistRepository {
	return &WatchlistRepository{db: conn}
}


func (r *WatchlistRepository) AddToWatchlist(c context.Context, userID, movieID int) error {
	_, err := r.db.Exec(c, `
		INSERT INTO watchlist (user_id, movie_id)
		VALUES ($1, $2) ON CONFLICT DO NOTHING;
	`, userID, movieID)

	return err
}

func (r *WatchlistRepository) GetWatchlist(c context.Context, userID int) ([]models.Movie, error) {
	rows, err := r.db.Query(c, `
		SELECT m.id, m.title, m.release_year, m.runtime, 
		       m.keywords, m.description, m.director, 
		       m.producer, m.cover, m.screenshots, m.movie_type_id
		FROM watchlist w
		JOIN movies m ON w.movie_id = m.id
		WHERE w.user_id = $1
		ORDER BY w.created_at DESC;
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var movie models.Movie
		if err := rows.Scan(
			&movie.Id, &movie.Title, &movie.ReleaseYear, &movie.Runtime,
			&movie.KeyWords, &movie.Description, &movie.Director,
			&movie.Producer, &movie.Media.Cover, &movie.Media.Screenshots, &movie.MovieTypeId,
		); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	return movies, nil
}

func (r *WatchlistRepository) RemoveFromWatchlist(c context.Context, userID, movieID int) error {
	_, err := r.db.Exec(c, `
		DELETE FROM watchlist WHERE user_id = $1 AND movie_id = $2;
	`, userID, movieID)
	return err
}

func (r *WatchlistRepository) IsInWatchlist(c context.Context, userID, movieID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(c, `
		SELECT EXISTS(SELECT 1 FROM watchlist WHERE user_id = $1 AND movie_id = $2);
	`, userID, movieID).Scan(&exists)
	return exists, err
}

