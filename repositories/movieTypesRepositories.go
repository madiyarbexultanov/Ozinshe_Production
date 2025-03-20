package repositories

import (
	"context"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MovieTypesRepository struct {
	db *pgxpool.Pool
}

func NewMovieTypesRepository(conn *pgxpool.Pool) *MovieTypesRepository {
	return &MovieTypesRepository{db: conn}
}

func (r *MovieTypesRepository) FindAll(c context.Context) ([]models.MovieType, error) {
	rows, err := r.db.Query(c, "SELECT id, title FROM movie_types")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movieTypes []models.MovieType
	for rows.Next() {
		var movieType models.MovieType
		err := rows.Scan(&movieType.Id, &movieType.Title)
		if err != nil {
			return nil, err
		}
		movieTypes = append(movieTypes, movieType)
	}

	return movieTypes, nil
}

func (r *MovieTypesRepository) FindById(c context.Context, id int) (models.MovieType, error) {
	var movieType models.MovieType
	row := r.db.QueryRow(c, "SELECT id, title FROM movie_types WHERE id = $1", id)
	err := row.Scan(&movieType.Id, &movieType.Title)
	if err != nil {
		return models.MovieType{}, err
	}
	return movieType, nil
}

func (r *MovieTypesRepository) Create(c context.Context, movieType models.MovieType) (int, error) {
	var id int
	row := r.db.QueryRow(c, "INSERT INTO movie_types (title) VALUES ($1) RETURNING id", movieType.Title)
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *MovieTypesRepository) Update(c context.Context, id int, movieType models.MovieType) error {
	_, err := r.db.Exec(c, "UPDATE movie_types SET title=$1 WHERE id=$2", movieType.Title, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *MovieTypesRepository) Delete(c context.Context, id int) error {
	_, err := r.db.Exec(c, "DELETE FROM movie_types WHERE id=$1", id)
	if err != nil {
		return err
	}
	return nil
}