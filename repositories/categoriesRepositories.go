package repositories

import (
	"context"
	"ozinshe_production/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoriesRepository struct {
	db *pgxpool.Pool
}

func NewCategoriesRepository(conn *pgxpool.Pool) *CategoriesRepository {
	return &CategoriesRepository{db: conn}
}

func (r *CategoriesRepository) FindAll(c context.Context) ([]models.Category, error) {
	rows, err := r.db.Query(c, "select id, title from categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.Id, &category.Title)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *CategoriesRepository) FindAllByIds(c context.Context, ids []int) ([]models.Category, error) {
	rows, err := r.db.Query(c, "select id, title from categories where id = any($1)", ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]models.Category, 0)
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.Id, &category.Title)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
        return nil, err
    }

	return categories, nil
}

func (r *CategoriesRepository) FindById(c context.Context, id int) (models.Category, error) {
	var categories models.Category
	row := r.db.QueryRow(c, "select id, title from categories where id = $1", id)
	err := row.Scan(&categories.Id, &categories.Title)
	if err != nil {
		return models.Category{}, err
	}
	return categories, nil
}

func (r *CategoriesRepository) Create(c context.Context, categories models.Category) (int, error) {
	var id int
	row := r.db.QueryRow(c, "insert into categories (title) values ($1) returning id", categories.Title)
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *CategoriesRepository) Update(c context.Context, id int, category models.Category) error {
	_, err := r.db.Exec(c, "update categories set title=$1 where id=$2", category.Title, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *CategoriesRepository) Delete(c context.Context, id int) error {
	_, err := r.db.Exec(c, "delete from categories where id=$1", id)
	if err != nil {
		return err
	}
	return nil
}
