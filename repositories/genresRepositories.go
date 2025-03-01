package repositories

import (
	"context"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GenresRepository struct {
	db *pgxpool.Pool
}

func NewGenresRepository(conn *pgxpool.Pool) *GenresRepository {
	return &GenresRepository{db: conn}
}

func (r *GenresRepository) FindAll(c context.Context) ([]models.Genre, error) {
	rows, err := r.db.Query(c, "select id, title, poster_url from genres")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []models.Genre
	for rows.Next() {
		var genre models.Genre
		err := rows.Scan(&genre.Id, &genre.Title, &genre.PosterUrl)
		if err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}

	return genres, nil
}

func (r *GenresRepository) FindById(c context.Context, id int) (models.Genre, error) {
	var genres models.Genre
	row := r.db.QueryRow(c, "select id, title, poster_url from genres where id = $1", id)
	err := row.Scan(&genres.Id, &genres.Title, &genres.PosterUrl)
	if err != nil {
		return models.Genre{}, err
	}
	return genres, nil
}

func (r *GenresRepository) Create(c context.Context, genres models.Genre) (int, error) {
	var id int
	row := r.db.QueryRow(c, "insert into genres (title, poster_url) values ($1, $2) returning id", genres.Title, genres.PosterUrl)
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *GenresRepository) Update(c context.Context, id int, genre models.Genre) error {
	_, err := r.db.Exec(c, "update genres set title=$1 where id=$2", genre.Title, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *GenresRepository) Delete(c context.Context, id int) error {
	_, err := r.db.Exec(c, "delete from genres where id=$1", id)
	if err != nil {
		return err
	}
	return nil
}
