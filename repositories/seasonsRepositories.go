package repositories

import (
	"context"
	"fmt"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SeasonsRepository struct {
	db *pgxpool.Pool
}

func NewSeasonsRepository(conn *pgxpool.Pool) *SeasonsRepository {
	return &SeasonsRepository{db: conn}
}

// Проверка, существует ли сезон у фильма
func (r *SeasonsRepository) Exists(c context.Context, movieID, seasonNumber int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM seasons WHERE movie_id=$1 AND number=$2)`
	err := r.db.QueryRow(c, query, movieID, seasonNumber).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if season exists: %w", err)
	}
	return exists, nil
}

// Создание сезона (с проверкой наличия)
func (r *SeasonsRepository) Create(c context.Context, season models.Season) (int, error) {
	exists, err := r.Exists(c, season.MovieID, season.Number)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, fmt.Errorf("season already exists for movie %d", season.MovieID)
	}

	var seasonID int
	query := `INSERT INTO seasons (movie_id, number) VALUES ($1, $2) RETURNING id`
	err = r.db.QueryRow(c, query, season.MovieID, season.Number).Scan(&seasonID)
	if err != nil {
		return 0, fmt.Errorf("failed to create season: %w", err)
	}
	return seasonID, nil
}

// Получение всех сезонов фильма
func (r *SeasonsRepository) FindAllByMovieID(c context.Context, movieID int) ([]models.Season, error) {
	rows, err := r.db.Query(c, `SELECT id, movie_id, number FROM seasons WHERE movie_id = $1`, movieID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch seasons: %w", err)
	}
	defer rows.Close()

	var seasons []models.Season
	for rows.Next() {
		var season models.Season
		if err := rows.Scan(&season.Id, &season.MovieID, &season.Number); err != nil {
			return nil, fmt.Errorf("failed to scan season: %w", err)
		}
		seasons = append(seasons, season)
	}
	return seasons, rows.Err()
}

// Обновление сезона
func (r *SeasonsRepository) Update(c context.Context, season models.Season) error {
	_, err := r.db.Exec(c, `UPDATE seasons SET number = $1 WHERE id = $2`, season.Number, season.Id)
	return err
}

// Удаление сезона
func (r *SeasonsRepository) Delete(c context.Context, seasonID int) error {
	_, err := r.db.Exec(c, `DELETE FROM seasons WHERE id = $1`, seasonID)
	return err
}