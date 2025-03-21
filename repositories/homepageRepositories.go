package repositories

import (
	"context"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HomepageRepository struct {
	db *pgxpool.Pool
}

func NewHomepageRepository(conn *pgxpool.Pool) *HomepageRepository {
	return &HomepageRepository{db: conn}
}

func (r *HomepageRepository) GetRecommendedMovies(c context.Context) ([]models.Movie, error) {
	rows, err := r.db.Query(c, `
        SELECT 
            m.id, m.title, m.release_year, m.runtime, 
            m.keywords, m.description, m.director, 
            m.producer, m.cover, m.screenshots, m.movie_type_id
        FROM recommended_movies rm
        JOIN movies m ON rm.movie_id = m.id
        ORDER BY rm.position
    `)
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