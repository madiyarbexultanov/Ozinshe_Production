package repositories

import (
	"context"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SearchRepository struct {
	db *pgxpool.Pool
}

// SearchResult будет хранить все результаты поиска
type SearchResult struct {
	Type   string      `json:"type"`   // Тип объекта (User, Category, Movie)
	ID     int         `json:"id"`     // ID объекта
	Entity interface{} `json:"entity"` // Данные объекта (пользователь, категория, фильм)
	URL    string      `json:"url"`    // URL для перехода на объект
}

type movieSearchResult struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Cover string `json:"cover"`
}

type userSearchResult struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type categorySearchResult struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
}

func NewSearchRepository(db *pgxpool.Pool) *SearchRepository {
	return &SearchRepository{db: db}
}

func (r *SearchRepository) SearchAll(c context.Context, query string) ([]SearchResult, error) {
	query = strings.ToLower(query)
	var results []SearchResult

	// Поиск пользователей
	userQuery := `SELECT id, name, email FROM users WHERE LOWER(name) LIKE $1 OR LOWER(email) LIKE $1`
	rows, err := r.db.Query(c, userQuery, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var res userSearchResult
		if err := rows.Scan(&res.Id, &res.Name, &res.Email); err != nil {
			return nil, err
		}
		results = append(results, SearchResult{
			Type:   "User",
			ID:     res.Id,
			Entity: res,
			URL:    "/users/" + strconv.Itoa(res.Id),
		})
	}

	// Поиск категорий
	categoryQuery := `SELECT id, title FROM categories WHERE LOWER(title) LIKE $1`
	rows, err = r.db.Query(c, categoryQuery, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var res categorySearchResult
		if err := rows.Scan(&res.Id, &res.Title); err != nil {
			return nil, err
		}
		results = append(results, SearchResult{
			Type:   "Category",
			ID:     res.Id,
			Entity: res,
			URL:    "/categories/" + strconv.Itoa(res.Id),
		})
	}

	// Поиск фильмов
	movieQuery := `SELECT id, title, cover FROM movies WHERE LOWER(title) LIKE $1 OR LOWER(description) LIKE $1 OR EXISTS (SELECT 1 FROM unnest(keywords) AS kw WHERE LOWER(kw) LIKE $1)`
	rows, err = r.db.Query(c, movieQuery, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var res movieSearchResult
		if err := rows.Scan(&res.Id, &res.Title, &res.Cover); err != nil {
			return nil, err
		}
		results = append(results, SearchResult{
			Type:   "Movie",
			ID:     res.Id,
			Entity: res,
			URL:    "/movies/" + strconv.Itoa(res.Id),
		})
	}

	return results, nil
}
