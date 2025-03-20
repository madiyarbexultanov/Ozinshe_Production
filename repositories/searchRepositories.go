package repositories

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SearchRepository struct {
	db *pgxpool.Pool
}


type SearchResult struct {
	Type        string 
	ID          int
	Title       string
	Description string
}


func NewSearchRepository(db *pgxpool.Pool) *SearchRepository {
	return &SearchRepository{db: db}
}


func (r *SearchRepository) SearchAll(c context.Context, query string) ([]SearchResult, error) {
	query = strings.ToLower(query)
	var results []SearchResult


	userQuery := `SELECT id, name, email FROM users WHERE LOWER(name) LIKE $1 OR LOWER(email) LIKE $1`
	rows, err := r.db.Query(c, userQuery, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var res SearchResult
		if err := rows.Scan(&res.ID, &res.Title, &res.Description); err != nil {
			return nil, err
		}
		res.Type = "User"
		results = append(results, res)
	}


	categoryQuery := `SELECT id, title FROM categories WHERE LOWER(title) LIKE $1`
	rows, err = r.db.Query(c, categoryQuery, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var res SearchResult
		if err := rows.Scan(&res.ID, &res.Title); err != nil {
			return nil, err
		}
		res.Type = "Category"
		results = append(results, res)
	}


	movieQuery := `SELECT id, title, description FROM movies WHERE LOWER(title) LIKE $1 OR LOWER(description) LIKE $1 OR EXISTS (SELECT 1 FROM unnest(key_words) AS kw WHERE LOWER(kw) LIKE $1)`
	rows, err = r.db.Query(c, movieQuery, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var res SearchResult
		if err := rows.Scan(&res.ID, &res.Title, &res.Description); err != nil {
			return nil, err
		}
		res.Type = "Movie"
		results = append(results, res)
	}

	return results, nil
}

