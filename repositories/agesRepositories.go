package repositories

import (
	"context"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AgesRepository struct {
	db *pgxpool.Pool
}

func NewAgesRepository(conn *pgxpool.Pool) *AgesRepository {
	return &AgesRepository{db: conn}
}

func (r *AgesRepository) FindAll(c context.Context) ([]models.Ages, error)  {
	rows, err := r.db.Query(c, "select id, title, poster_url from ages")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ages []models.Ages
	for rows.Next() {
		var age models.Ages
		err := rows.Scan(&age.Id, &age.Title, &age.PosterUrl)
		if err != nil {
			return nil, err
		}
		ages = append(ages, age)
	}

	return ages, nil
}

func (r *AgesRepository) FindAllByIds(c context.Context, ids []int) ([]models.Ages, error) {
	rows, err := r.db.Query(c, "select id, title from ages where id = any($1)", ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ages := make([]models.Ages, 0)
	for rows.Next() {
		var age models.Ages
		err := rows.Scan(&age.Id, &age.Title)
		if err != nil {
			return nil, err
		}
		ages = append(ages, age)
	}

	if err = rows.Err(); err != nil {
        return nil, err
    }

	return ages, nil
}


func (r *AgesRepository) FindById(c context.Context, id int) (models.Ages, error) {
	var ages models.Ages
	row := r.db.QueryRow(c, "select id, title, poster_url from ages where id = $1", id)
	err := row.Scan(&ages.Id, &ages.Title, &ages.PosterUrl)
	if err != nil {
		return models.Ages{}, err
	}
	return ages, nil
}

func (r *AgesRepository) Create(c context.Context, ages models.Ages) (int, error) {
	var id int
	row := r.db.QueryRow(c, "insert into ages (title, poster_url) values ($1, $2) returning id", ages.Title, ages.PosterUrl)
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *AgesRepository) Update(c context.Context, id int, ages models.Ages) error {
	_, err := r.db.Exec(c, "update ages set title=$1 where id=$2", ages.Title, id)
	if err != nil {
		return err
	}

	return nil
}


func (r *AgesRepository) Delete(c context.Context, id int) error {
	_, err := r.db.Exec(c, "delete from ages where id=$1", id)
	if err != nil {
		return err
	}
	return nil
}
