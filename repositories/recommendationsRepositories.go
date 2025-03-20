package repositories

import (
	"context"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RecommendationsRepository struct {
	db *pgxpool.Pool
}

func NewRecommendationsRepository(conn *pgxpool.Pool) *RecommendationsRepository {
	return &RecommendationsRepository{db: conn}
}

func (r *RecommendationsRepository) FindAll(c context.Context) ([]models.RecommendedMovie, error) {
	rows, err := r.db.Query(c, "SELECT id, movie_id, position FROM recommended_movies ORDER BY position ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recommendations []models.RecommendedMovie
	for rows.Next() {
		var rec models.RecommendedMovie
		if err := rows.Scan(&rec.Id, &rec.MovieID, &rec.Position); err != nil {
			return nil, err
		}
		recommendations = append(recommendations, rec)
	}
	return recommendations, nil
}

func (r *RecommendationsRepository) FindById(c context.Context, id int) (models.RecommendedMovie, error) {
	var rec models.RecommendedMovie
	row := r.db.QueryRow(c, "SELECT id, movie_id, position FROM recommended_movies WHERE id = $1", id)
	err := row.Scan(&rec.Id, &rec.MovieID, &rec.Position)
	if err != nil {
		return models.RecommendedMovie{}, err
	}
	return rec, nil
}

func (r *RecommendationsRepository) Create(c context.Context, movieID, position int) (int, error) {
	var id int
	row := r.db.QueryRow(c, "INSERT INTO recommended_movies (movie_id, position) VALUES ($1, $2) RETURNING id", movieID, position)
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *RecommendationsRepository) Update(c context.Context, id, newPosition int) error {
	_, err := r.db.Exec(c, "UPDATE recommended_movies SET position = $1 WHERE id = $2", newPosition, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *RecommendationsRepository) Delete(c context.Context, id int) error {
	_, err := r.db.Exec(c, "DELETE FROM recommended_movies WHERE id = $1", id)
	if err != nil {
		return err
	}

	return nil
}
